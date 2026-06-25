package worker

import (
	"context"
	"sync"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
)

// Config supplies the timing parameters used by the Worker loop.
// Defaults are exposed via DefaultConfig so callers can override only the
// fields they care about instead of repeating magic numbers.
type Config struct {
	LeaseTTL      time.Duration
	RenewInterval time.Duration
	CommandLease  time.Duration
	IdlePoll      time.Duration
}

// DefaultConfig returns the plan-approved defaults: 15s lease TTL, 5s renewal,
// 60s command lease, and 500ms idle polling.
func DefaultConfig() Config {
	return Config{
		LeaseTTL:      15 * time.Second,
		RenewInterval: 5 * time.Second,
		CommandLease:  60 * time.Second,
		IdlePoll:      500 * time.Millisecond,
	}
}

// LeaderGuard fences Worker leadership. redis.Lease satisfies this interface.
type LeaderGuard interface {
	Acquire(ctx context.Context) (bool, error)
	Renew(ctx context.Context) (bool, error)
	Release(ctx context.Context) (bool, error)
	Value() string
}

// CommandClaimer abstracts the durable command queue operations the Worker
// drives while it is the elected leader.
type CommandClaimer interface {
	ClaimNext(ctx context.Context, workerID string, lease time.Duration) (*entity.FlowCommand, error)
	RequeueExpired(now time.Time) (int64, error)
}

// Recoverer rebuilds in-memory flow state after a Worker wins leadership.
type Recoverer interface {
	Recover(ctx context.Context) error
}

// EpochRenewer advances and reads the singleton worker epoch used to fence
// command ownership. repository.WorkerEpochRepo satisfies this interface.
// A nil EpochRenewer disables epoch-based fencing (useful for tests that do
// not exercise the fencing path).
type EpochRenewer interface {
	AdvanceEpoch(ctx context.Context, workerID string, leaseTTL time.Duration) (*entity.WorkerEpoch, error)
	CurrentEpoch(ctx context.Context) (*entity.WorkerEpoch, error)
}

// CommandLeaseExtender pushes the lease_until forward on the currently
// executing command. repository.FlowCommandRepo satisfies this interface via
// ExtendLeaseFenced. A nil CommandLeaseExtender disables command-lease
// renewal (useful for tests that do not exercise long-running commands).
type CommandLeaseExtender interface {
	ExtendLeaseFenced(ctx context.Context, id uint64, ownership repository.CommandOwnership, until time.Time) (bool, error)
}

// ExecutionFence is the value object passed to the Executor for each command.
// It carries the worker_id, worker_epoch, and lease_token that the executor
// must use when calling fenced repository mutations. A stale fence (one whose
// epoch is less than the current singleton epoch) causes fenced updates to
// affect zero rows, which the executor observes as repository.ErrOwnershipLost.
type ExecutionFence struct {
	WorkerID    string
	WorkerEpoch uint64
	LeaseToken  string
}

// Executor maps a durable command to its transactional side effects. The
// fence carries the worker's current epoch and the command's lease_token;
// the executor must use them when committing results so that a stale worker
// cannot flip a command that has been re-claimed by a newer worker.
type Executor interface {
	Execute(ctx context.Context, cmd *entity.FlowCommand, fence ExecutionFence) error
}

// Worker is the singleton runtime that elects a leader via a fenced lease,
// recovers flow state, and drives the durable command queue.
type Worker struct {
	config        Config
	lease         LeaderGuard
	claimer       CommandClaimer
	recoverer     Recoverer
	executor      Executor
	epochRenewer  EpochRenewer
	leaseExtender CommandLeaseExtender
	workerID      string

	mu        sync.RWMutex
	status    Status
	runCancel context.CancelFunc

	// Leader-state, protected by mu. currentEpoch is bumped by serveAsLeader
	// on acquisition and by runEpochRenewal on each renewal tick. inflight
	// tracks the commands currently being executed so the command-lease
	// renewal ticker can extend their leases.
	currentEpoch uint64
	inflight     map[uint64]ExecutionFence

	// execWG tracks in-flight executor goroutines so Shutdown can wait for
	// them to observe context cancellation before releasing the lease.
	execWG sync.WaitGroup
}

// NewWorker constructs a Worker with the given dependencies. epochRenewer
// and leaseExtender may be nil to disable epoch fencing and command-lease
// renewal respectively (primarily for tests). The Worker is not started;
// call Run to begin the election loop.
func NewWorker(
	config Config,
	lease LeaderGuard,
	claimer CommandClaimer,
	recoverer Recoverer,
	executor Executor,
	epochRenewer EpochRenewer,
	leaseExtender CommandLeaseExtender,
	workerID string,
) *Worker {
	return &Worker{
		config:        config,
		lease:         lease,
		claimer:       claimer,
		recoverer:     recoverer,
		executor:      executor,
		epochRenewer:  epochRenewer,
		leaseExtender: leaseExtender,
		workerID:      workerID,
		status:        StatusStandby,
		inflight:      make(map[uint64]ExecutionFence),
	}
}

// Status reports the current Worker state. Safe for concurrent use.
func (w *Worker) Status() Status {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status
}

func (w *Worker) setStatus(s Status) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = s
}

func (w *Worker) isStopping() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status == StatusStopping
}

// Shutdown marks the Worker as stopping, cancels the Run loop, waits for
// in-flight executor goroutines to observe the cancellation, and releases
// the current lease token. It releases only the lease this Worker holds.
func (w *Worker) Shutdown(ctx context.Context) error {
	w.mu.Lock()
	w.status = StatusStopping
	if w.runCancel != nil {
		w.runCancel()
	}
	w.mu.Unlock()

	// Wait for in-flight executions to finish (or the shutdown context to
	// expire). Correctness does not depend on this wait — fencing protects
	// against stale commits — but it avoids goroutine leaks.
	waitDone := make(chan struct{})
	go func() {
		w.execWG.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-ctx.Done():
	}

	_, err := w.lease.Release(ctx)
	return err
}

// Run performs the leader election and command processing loop:
//  1. lease acquisition loop (standby while not leader);
//  2. Recover after acquisition;
//  3. MySQL epoch acquisition (fencing off stale workers);
//  4. expired command requeue;
//  5. claim/execute loop with parallel renewal tickers;
//  6. immediate demotion on any renewal failure.
//
// Run blocks until ctx is cancelled or Shutdown is called.
func (w *Worker) Run(ctx context.Context) error {
	if w.isStopping() {
		return nil
	}

	runCtx, runCancel := context.WithCancel(ctx)
	w.mu.Lock()
	w.runCancel = runCancel
	w.status = StatusStandby
	w.mu.Unlock()
	defer func() {
		w.mu.Lock()
		w.runCancel = nil
		w.mu.Unlock()
	}()

	for {
		if w.isStopping() {
			return nil
		}

		w.setStatus(StatusStandby)
		acquired, err := w.lease.Acquire(runCtx)
		if err != nil || !acquired {
			if err := w.wait(runCtx); err != nil {
				return w.exitErr(err)
			}
			continue
		}

		if err := w.serveAsLeader(runCtx); err != nil {
			return w.exitErr(err)
		}
	}
}

// serveAsLeader runs the recover -> epoch-acquire -> requeue -> leader loop.
// It returns a non-nil error only when the context is cancelled; renewal
// failure demotes back to standby without an error so the outer loop retries.
func (w *Worker) serveAsLeader(runCtx context.Context) error {
	w.setStatus(StatusRecovering)
	if w.recoverer != nil {
		if err := w.recoverer.Recover(runCtx); err != nil {
			_, _ = w.lease.Release(runCtx)
			w.setStatus(StatusStandby)
			return w.wait(runCtx)
		}
	}

	// Acquire the MySQL epoch BEFORE claiming any command. This guarantees
	// every command we claim is stamped with an epoch strictly greater than
	// any prior worker's, so a stale worker cannot commit after a newer
	// worker has taken over.
	if w.epochRenewer != nil {
		epoch, err := w.epochRenewer.AdvanceEpoch(runCtx, w.workerID, w.config.LeaseTTL)
		if err != nil || epoch == nil {
			_, _ = w.lease.Release(runCtx)
			w.setStatus(StatusStandby)
			return w.wait(runCtx)
		}
		w.mu.Lock()
		w.currentEpoch = epoch.Epoch
		w.mu.Unlock()
	}

	if _, err := w.claimer.RequeueExpired(time.Now()); err != nil {
		_, _ = w.lease.Release(runCtx)
		w.setStatus(StatusStandby)
		return w.wait(runCtx)
	}

	w.setStatus(StatusLeaderReady)
	return w.runLeaderLoop(runCtx)
}

// runLeaderLoop claims and dispatches commands while three parallel renewal
// tickers (Redis lease, MySQL epoch, command lease) keep the worker's
// authority current. Any renewal failure triggers demotion so the outer
// loop re-acquires leadership.
func (w *Worker) runLeaderLoop(runCtx context.Context) error {
	leaderCtx, cancelLeader := context.WithCancel(runCtx)
	defer cancelLeader()

	demotionCh := make(chan struct{})
	var demotionOnce sync.Once
	signalDemotion := func() {
		demotionOnce.Do(func() { close(demotionCh) })
	}

	// Renewal goroutine: runs Redis, epoch, and command-lease renewals in
	// parallel. Any failure triggers demotion. Execution happens in
	// separate goroutines (see claimAndDispatch) so renewal never blocks
	// on a long-running command.
	renewalDone := make(chan struct{})
	go func() {
		defer close(renewalDone)
		w.runRenewals(leaderCtx, signalDemotion)
	}()

	idleTimer := time.NewTimer(w.config.IdlePoll)
	defer idleTimer.Stop()

	for {
		select {
		case <-leaderCtx.Done():
			return leaderCtx.Err()
		case <-demotionCh:
			// Stop the renewals and wait for them to exit before
			// returning, so the next acquisition starts clean.
			cancelLeader()
			<-renewalDone
			w.setStatus(StatusStandby)
			return nil
		case <-idleTimer.C:
			w.claimAndDispatch(leaderCtx)
			idleTimer.Reset(w.config.IdlePoll)
		}
	}
}

// runRenewals spawns three parallel renewal goroutines and blocks until all
// of them exit (either because ctx was cancelled or signalDemotion was
// triggered by one of them).
func (w *Worker) runRenewals(ctx context.Context, signalDemotion func()) {
	redisTicker := time.NewTicker(w.config.RenewInterval)
	epochTicker := time.NewTicker(w.config.RenewInterval)
	cmdLeaseTicker := time.NewTicker(w.config.RenewInterval)
	defer redisTicker.Stop()
	defer epochTicker.Stop()
	defer cmdLeaseTicker.Stop()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		w.runRedisRenewal(ctx, redisTicker, signalDemotion)
	}()
	go func() {
		defer wg.Done()
		w.runEpochRenewal(ctx, epochTicker, signalDemotion)
	}()
	go func() {
		defer wg.Done()
		w.runCommandLeaseRenewal(ctx, cmdLeaseTicker, signalDemotion)
	}()
	wg.Wait()
}

// runRedisRenewal renews the Redis leadership lease on each tick. A failed
// renewal (error or false) means another worker has taken over; we signal
// demotion so the outer loop re-acquires.
func (w *Worker) runRedisRenewal(ctx context.Context, ticker *time.Ticker, signalDemotion func()) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok, err := w.lease.Renew(ctx)
			if err != nil || !ok {
				signalDemotion()
				return
			}
		}
	}
}

// runEpochRenewal bumps the singleton MySQL epoch on each tick, refreshing
// the lease_until and re-asserting this worker as the current owner. A
// failure means we can no longer fence off stale workers, so we demote.
func (w *Worker) runEpochRenewal(ctx context.Context, ticker *time.Ticker, signalDemotion func()) {
	if w.epochRenewer == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			epoch, err := w.epochRenewer.AdvanceEpoch(ctx, w.workerID, w.config.LeaseTTL)
			if err != nil || epoch == nil {
				signalDemotion()
				return
			}
			w.mu.Lock()
			w.currentEpoch = epoch.Epoch
			w.mu.Unlock()
		}
	}
}

// runCommandLeaseRenewal extends the lease_until of every in-flight command
// on each tick. If any extend fails (ownership lost because a newer worker
// bumped the epoch or the lease expired), we demote. Correctness comes from
// the fenced WHERE clause in ExtendLeaseFenced, not from this demotion.
func (w *Worker) runCommandLeaseRenewal(ctx context.Context, ticker *time.Ticker, signalDemotion func()) {
	if w.leaseExtender == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := w.snapshotInflight()
			for id, fence := range snapshot {
				ownership := repository.CommandOwnership{
					WorkerID: fence.WorkerID,
					Epoch:    fence.WorkerEpoch,
					Token:    fence.LeaseToken,
				}
				ok, err := w.leaseExtender.ExtendLeaseFenced(ctx, id, ownership, time.Now().Add(w.config.CommandLease))
				if err != nil || !ok {
					signalDemotion()
					return
				}
			}
		}
	}
}

// claimAndDispatch claims the next pending command (if any) and dispatches
// it to the executor in a separate goroutine. The goroutine receives a
// fenced context (cancelled on demotion) and an ExecutionFence carrying the
// worker's current epoch and the command's lease_token. The executor must
// use the fence for all mutations; cancellation is a hint, not a correctness
// mechanism.
func (w *Worker) claimAndDispatch(ctx context.Context) {
	cmd, err := w.claimer.ClaimNext(ctx, w.workerID, w.config.CommandLease)
	if err != nil || cmd == nil || w.executor == nil {
		return
	}

	fence := ExecutionFence{
		WorkerID:    w.workerID,
		WorkerEpoch: w.currentEpochValue(),
		LeaseToken:  cmd.LeaseToken,
	}
	w.addInflight(cmd, fence)

	w.execWG.Add(1)
	go func() {
		defer w.execWG.Done()
		defer w.removeInflight(cmd.ID)
		_ = w.executor.Execute(ctx, cmd, fence)
	}()
}

func (w *Worker) currentEpochValue() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentEpoch
}

func (w *Worker) addInflight(cmd *entity.FlowCommand, fence ExecutionFence) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.inflight[cmd.ID] = fence
}

func (w *Worker) removeInflight(id uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.inflight, id)
}

func (w *Worker) snapshotInflight() map[uint64]ExecutionFence {
	w.mu.RLock()
	defer w.mu.RUnlock()
	cp := make(map[uint64]ExecutionFence, len(w.inflight))
	for k, v := range w.inflight {
		cp[k] = v
	}
	return cp
}

// wait blocks for the idle poll interval or until the context is cancelled.
func (w *Worker) wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(w.config.IdlePoll):
		return nil
	}
}

// exitErr maps context cancellation during a clean shutdown to nil.
func (w *Worker) exitErr(err error) error {
	if w.isStopping() {
		return nil
	}
	return err
}
