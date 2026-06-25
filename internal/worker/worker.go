package worker

import (
	"context"
	"sync"
	"time"

	"drill-platform/internal/domain/entity"
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

// Executor maps a durable command to its transactional side effects.
// It is implemented in Task 8; a nil Executor makes the Worker claim-only.
type Executor interface {
	Execute(ctx context.Context, cmd *entity.FlowCommand) error
}

// Worker is the singleton runtime that elects a leader via a fenced lease,
// recovers flow state, and drives the durable command queue.
type Worker struct {
	config    Config
	lease     LeaderGuard
	claimer   CommandClaimer
	recoverer Recoverer
	executor  Executor
	workerID  string

	mu        sync.RWMutex
	status    Status
	runCancel context.CancelFunc
}

// NewWorker constructs a Worker with the given dependencies. The Worker is
// not started; call Run to begin the election loop.
func NewWorker(config Config, lease LeaderGuard, claimer CommandClaimer, recoverer Recoverer, executor Executor, workerID string) *Worker {
	return &Worker{
		config:    config,
		lease:     lease,
		claimer:   claimer,
		recoverer: recoverer,
		executor:  executor,
		workerID:  workerID,
		status:    StatusStandby,
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

// Shutdown marks the Worker as stopping, cancels the Run loop, and releases
// the current lease token. It releases only the lease this Worker holds.
func (w *Worker) Shutdown(ctx context.Context) error {
	w.mu.Lock()
	w.status = StatusStopping
	if w.runCancel != nil {
		w.runCancel()
	}
	w.mu.Unlock()

	_, err := w.lease.Release(ctx)
	return err
}

// Run performs the leader election and command processing loop:
//  1. lease acquisition loop (standby while not leader);
//  2. Recover after acquisition;
//  3. expired command requeue;
//  4. claim/execute loop;
//  5. renewal ticker;
//  6. immediate demotion on renewal failure.
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

// serveAsLeader runs the recover -> requeue -> claim/execute -> renew cycle.
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

	if _, err := w.claimer.RequeueExpired(time.Now()); err != nil {
		_, _ = w.lease.Release(runCtx)
		w.setStatus(StatusStandby)
		return w.wait(runCtx)
	}

	w.setStatus(StatusLeaderReady)
	return w.runLeaderLoop(runCtx)
}

// runLeaderLoop interleaves command claiming with lease renewal. On renewal
// failure it demotes to standby and returns nil so the outer loop re-acquires.
func (w *Worker) runLeaderLoop(runCtx context.Context) error {
	renewTicker := time.NewTicker(w.config.RenewInterval)
	defer renewTicker.Stop()

	idleTimer := time.NewTimer(w.config.IdlePoll)
	defer idleTimer.Stop()

	for {
		select {
		case <-runCtx.Done():
			return runCtx.Err()
		case <-renewTicker.C:
			ok, err := w.lease.Renew(runCtx)
			if err != nil {
				w.setStatus(StatusStandby)
				return nil
			}
			if !ok {
				w.setStatus(StatusStandby)
				return nil
			}
		case <-idleTimer.C:
			cmd, err := w.claimer.ClaimNext(runCtx, w.workerID, w.config.CommandLease)
			if err != nil {
				w.setStatus(StatusStandby)
				return nil
			}
			if cmd != nil && w.executor != nil {
				_ = w.executor.Execute(runCtx, cmd)
			}
			idleTimer.Reset(w.config.IdlePoll)
		}
	}
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
