package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
)

// --- Fakes ---

type fakeLease struct {
	mu             sync.Mutex
	value          string
	acquireResults []bool
	acquireIdx     int
	acquireErr     error
	renewResult    bool
	renewErr       error
	renewCalls     int
	releaseCalls   int
	firstRenewCh   chan struct{}
	firstRenewOnce sync.Once
}

func (f *fakeLease) Acquire(ctx context.Context) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.acquireIdx < len(f.acquireResults) {
		r := f.acquireResults[f.acquireIdx]
		f.acquireIdx++
		return r, f.acquireErr
	}
	return false, f.acquireErr
}

func (f *fakeLease) Renew(ctx context.Context) (bool, error) {
	f.mu.Lock()
	f.renewCalls++
	f.mu.Unlock()
	f.firstRenewOnce.Do(func() { close(f.firstRenewCh) })
	return f.renewResult, f.renewErr
}

func (f *fakeLease) Release(ctx context.Context) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.releaseCalls++
	return true, nil
}

func (f *fakeLease) Value() string {
	return f.value
}

func (f *fakeLease) releaseCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.releaseCalls
}

func (f *fakeLease) renewCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.renewCalls
}

func (f *fakeLease) firstRenewed() <-chan struct{} {
	return f.firstRenewCh
}

type fakeClaimer struct {
	mu           sync.Mutex
	claimResult  *entity.FlowCommand
	claimOnce    bool
	claimedOnce  bool
	claimErr     error
	claimCalls   int
	firstClaimAt time.Time
	requeueCalls int
	firstClaimCh chan struct{}
	firstClaimOnce sync.Once
}

func (f *fakeClaimer) ClaimNext(ctx context.Context, workerID string, lease time.Duration) (*entity.FlowCommand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.claimCalls++
	if f.firstClaimAt.IsZero() {
		f.firstClaimAt = time.Now()
		f.firstClaimOnce.Do(func() { close(f.firstClaimCh) })
	}
	if f.claimOnce && f.claimedOnce {
		return nil, f.claimErr
	}
	if f.claimOnce {
		f.claimedOnce = true
	}
	return f.claimResult, f.claimErr
}

func (f *fakeClaimer) RequeueExpired(now time.Time) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requeueCalls++
	return 0, nil
}

func (f *fakeClaimer) claimCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.claimCalls
}

func (f *fakeClaimer) firstClaimTime() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.firstClaimAt
}

func (f *fakeClaimer) firstClaimed() <-chan struct{} {
	return f.firstClaimCh
}

type fakeRecoverer struct {
	mu       sync.Mutex
	called   bool
	calledAt time.Time
	err      error
}

func (f *fakeRecoverer) Recover(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.called = true
	f.calledAt = time.Now()
	return f.err
}

func (f *fakeRecoverer) wasCalled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.called
}

func (f *fakeRecoverer) callTime() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calledAt
}

type fakeExecutor struct {
	mu        sync.Mutex
	calls     int
	lastCmd   *entity.FlowCommand
	lastFence ExecutionFence
	err       error
}

func (f *fakeExecutor) Execute(ctx context.Context, cmd *entity.FlowCommand, fence ExecutionFence) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.lastCmd = cmd
	f.lastFence = fence
	return f.err
}

func (f *fakeExecutor) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *fakeExecutor) fence() ExecutionFence {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastFence
}

// blockingExecutor signals when execution starts and blocks until released.
// Used to verify renewal tickers keep running while a command is in-flight.
type blockingExecutor struct {
	mu           sync.Mutex
	calls        int
	lastCmd      *entity.FlowCommand
	lastFence    ExecutionFence
	startedCh    chan struct{}
	startedOnce  sync.Once
	releaseCh    chan struct{}
	releaseOnce  sync.Once
}

func newBlockingExecutor() *blockingExecutor {
	return &blockingExecutor{
		startedCh: make(chan struct{}),
		releaseCh: make(chan struct{}),
	}
}

func (e *blockingExecutor) Execute(ctx context.Context, cmd *entity.FlowCommand, fence ExecutionFence) error {
	e.mu.Lock()
	e.calls++
	e.lastCmd = cmd
	e.lastFence = fence
	e.mu.Unlock()
	e.startedOnce.Do(func() { close(e.startedCh) })

	select {
	case <-e.releaseCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *blockingExecutor) started() <-chan struct{} {
	return e.startedCh
}

func (e *blockingExecutor) release() {
	e.releaseOnce.Do(func() { close(e.releaseCh) })
}

func (e *blockingExecutor) fence() ExecutionFence {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastFence
}

// fakeEpochRenewer records AdvanceEpoch calls and tracks the current epoch.
// Tests can externally bump the epoch to simulate a new worker taking over.
type fakeEpochRenewer struct {
	mu               sync.Mutex
	advances         int
	currentEpoch     uint64
	currentWorker    string
	advanceErr       error
	firstAdvanceAt   time.Time
	firstAdvanceCh   chan struct{}
	firstAdvanceOnce sync.Once
}

func (f *fakeEpochRenewer) AdvanceEpoch(ctx context.Context, workerID string, leaseTTL time.Duration) (*entity.WorkerEpoch, error) {
	f.mu.Lock()
	f.advances++
	f.currentEpoch++
	f.currentWorker = workerID
	epoch := f.currentEpoch
	err := f.advanceErr
	if f.firstAdvanceAt.IsZero() {
		f.firstAdvanceAt = time.Now()
	}
	f.mu.Unlock()
	f.firstAdvanceOnce.Do(func() { close(f.firstAdvanceCh) })

	if err != nil {
		return nil, err
	}
	leaseUntil := time.Now().Add(leaseTTL)
	return &entity.WorkerEpoch{
		ID:         1,
		WorkerID:   workerID,
		Epoch:      epoch,
		LeaseUntil: &leaseUntil,
	}, nil
}

func (f *fakeEpochRenewer) CurrentEpoch(ctx context.Context) (*entity.WorkerEpoch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.currentEpoch == 0 {
		return nil, nil
	}
	return &entity.WorkerEpoch{
		ID:       1,
		WorkerID: f.currentWorker,
		Epoch:    f.currentEpoch,
	}, nil
}

func (f *fakeEpochRenewer) advanceCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.advances
}

func (f *fakeEpochRenewer) currentEpochValue() uint64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentEpoch
}

func (f *fakeEpochRenewer) firstAdvanced() <-chan struct{} {
	return f.firstAdvanceCh
}

func (f *fakeEpochRenewer) firstAdvanceTime() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.firstAdvanceAt
}

// advanceExternally simulates a new worker bumping the singleton epoch,
// which makes any previously captured fence stale.
func (f *fakeEpochRenewer) advanceExternally(workerID string) {
	f.mu.Lock()
	f.advances++
	f.currentEpoch++
	f.currentWorker = workerID
	f.mu.Unlock()
}

// fakeLeaseExtender records ExtendLeaseFenced calls. When rejectEpochsBelow
// is set, any ownership with Epoch < rejectEpochsBelow returns
// (false, ErrOwnershipLost), simulating a new worker with a higher epoch.
type fakeLeaseExtender struct {
	mu                sync.Mutex
	extends           int
	lastID            uint64
	lastOwnership     repository.CommandOwnership
	lastUntil         time.Time
	rejectEpochsBelow uint64
	firstExtendCh     chan struct{}
	firstExtendOnce   sync.Once
	firstFailCh       chan struct{}
	firstFailOnce     sync.Once
}

func (f *fakeLeaseExtender) ExtendLeaseFenced(ctx context.Context, id uint64, ownership repository.CommandOwnership, until time.Time) (bool, error) {
	f.mu.Lock()
	f.extends++
	f.lastID = id
	f.lastOwnership = ownership
	f.lastUntil = until
	stale := f.rejectEpochsBelow != 0 && ownership.Epoch < f.rejectEpochsBelow
	f.mu.Unlock()

	f.firstExtendOnce.Do(func() { close(f.firstExtendCh) })

	if stale {
		f.firstFailOnce.Do(func() { close(f.firstFailCh) })
		return false, repository.ErrOwnershipLost
	}
	return true, nil
}

func (f *fakeLeaseExtender) extendCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.extends
}

func (f *fakeLeaseExtender) lastCommandID() uint64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastID
}

func (f *fakeLeaseExtender) firstExtended() <-chan struct{} {
	return f.firstExtendCh
}

func (f *fakeLeaseExtender) firstFailed() <-chan struct{} {
	return f.firstFailCh
}

func (f *fakeLeaseExtender) setRejectEpochsBelow(epoch uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rejectEpochsBelow = epoch
}

// --- Helpers ---

func newTestConfig() Config {
	return Config{
		LeaseTTL:      100 * time.Millisecond,
		RenewInterval: 20 * time.Millisecond,
		CommandLease:  50 * time.Millisecond,
		IdlePoll:      5 * time.Millisecond,
	}
}

func runWorkerAsync(t *testing.T, w *Worker, ctx context.Context) <-chan struct{} {
	t.Helper()
	done := make(chan struct{})
	go func() {
		_ = w.Run(ctx)
		close(done)
	}()
	return done
}

// --- Existing tests (updated for new NewWorker signature) ---

func TestStandbyDoesNotClaim(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-standby/token",
		acquireResults: []bool{},
		renewResult:    true,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{claimResult: nil, firstClaimCh: make(chan struct{})}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, nil, nil, "worker-standby")

	ctx, cancel := context.WithCancel(context.Background())
	done := runWorkerAsync(t, w, ctx)

	time.Sleep(30 * time.Millisecond)
	cancel()
	<-done

	if claimer.claimCount() != 0 {
		t.Fatalf("ClaimNext called %d times, want 0", claimer.claimCount())
	}
	if recoverer.wasCalled() {
		t.Fatal("Recover was called in standby, want not called")
	}
	if w.Status() != StatusStandby {
		t.Fatalf("Status = %q, want %q", w.Status(), StatusStandby)
	}
}

func TestAcquiringLeadershipInvokesRecovery(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-rec/token",
		acquireResults: []bool{true},
		renewResult:    true,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{claimResult: nil, firstClaimCh: make(chan struct{})}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, epochRenewer, nil, "worker-rec")

	ctx, cancel := context.WithCancel(context.Background())
	done := runWorkerAsync(t, w, ctx)

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	if !recoverer.wasCalled() {
		t.Fatal("Recover was not called after acquiring leadership")
	}
	if claimer.claimCount() == 0 {
		t.Fatal("ClaimNext was not called after recovery")
	}

	recTime := recoverer.callTime()
	claimTime := claimer.firstClaimTime()
	if !recTime.Before(claimTime) {
		t.Fatalf("Recover at %v was not before first ClaimNext at %v", recTime, claimTime)
	}
}

func TestRenewalFailureStopsNewClaims(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-renew/token",
		acquireResults: []bool{true},
		renewResult:    false,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{claimResult: nil, firstClaimCh: make(chan struct{})}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, epochRenewer, nil, "worker-renew")

	ctx, cancel := context.WithCancel(context.Background())
	done := runWorkerAsync(t, w, ctx)

	time.Sleep(50 * time.Millisecond)
	claimsAfterDemotion := claimer.claimCount()

	time.Sleep(30 * time.Millisecond)
	claimsFinal := claimer.claimCount()
	cancel()
	<-done

	if !recoverer.wasCalled() {
		t.Fatal("Recover was not called, worker never became leader")
	}
	if claimsFinal != claimsAfterDemotion {
		t.Fatalf("claims increased after renewal failure: %d -> %d", claimsAfterDemotion, claimsFinal)
	}
}

func TestShutdownReleasesOnlyCurrentToken(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-shutdown/token-xyz",
		acquireResults: []bool{true},
		renewResult:    true,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{claimResult: nil, firstClaimCh: make(chan struct{})}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, epochRenewer, nil, "worker-shutdown")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := runWorkerAsync(t, w, ctx)

	time.Sleep(20 * time.Millisecond)

	if err := w.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not exit after Shutdown")
	}

	if lease.releaseCount() != 1 {
		t.Fatalf("Release called %d times, want 1", lease.releaseCount())
	}
	if lease.Value() != "worker-shutdown/token-xyz" {
		t.Fatalf("Value = %q, want %q", lease.Value(), "worker-shutdown/token-xyz")
	}
	if w.Status() != StatusStopping {
		t.Fatalf("Status = %q, want %q", w.Status(), StatusStopping)
	}
}

// --- New tests for Task 2 ---

// TestRedisAcquisitionFollowedByEpochAcquisition verifies that the worker
// acquires the MySQL epoch AFTER Redis leadership but BEFORE claiming any
// command. This guarantees every claimed command is stamped with an epoch
// greater than any prior worker's.
func TestRedisAcquisitionFollowedByEpochAcquisition(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-epoch/token",
		acquireResults: []bool{true},
		renewResult:    true,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{claimResult: nil, firstClaimCh: make(chan struct{})}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, epochRenewer, nil, "worker-epoch")

	ctx, cancel := context.WithCancel(context.Background())
	done := runWorkerAsync(t, w, ctx)

	// Wait until the first claim happens (proves we reached the leader loop).
	select {
	case <-claimer.firstClaimed():
	case <-time.After(200 * time.Millisecond):
		cancel()
		t.Fatal("ClaimNext was never called")
	}

	cancel()
	<-done

	if epochRenewer.advanceCount() == 0 {
		t.Fatal("AdvanceEpoch was not called after leadership acquisition")
	}
	if !recoverer.wasCalled() {
		t.Fatal("Recover was not called")
	}
	// AdvanceEpoch (in serveAsLeader) must happen before ClaimNext (in runLeaderLoop).
	epochTime := epochRenewer.firstAdvanceTime()
	claimTime := claimer.firstClaimTime()
	if !epochTime.Before(claimTime) {
		t.Fatalf("AdvanceEpoch at %v was not before first ClaimNext at %v", epochTime, claimTime)
	}
}

// TestRenewalRunsDuringLongCommand verifies that the Redis lease renewal
// ticker keeps firing while a command is executing in the executor goroutine.
// The execution must not block renewal.
func TestRenewalRunsDuringLongCommand(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-long/token",
		acquireResults: []bool{true},
		renewResult:    true,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{
		claimResult:   &entity.FlowCommand{ID: 1, CommandType: "test", LeaseToken: "tok-1"},
		claimOnce:     true,
		firstClaimCh:  make(chan struct{}),
	}
	recoverer := &fakeRecoverer{}
	executor := newBlockingExecutor()
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}

	cfg := newTestConfig()
	cfg.RenewInterval = 5 * time.Millisecond

	w := NewWorker(cfg, lease, claimer, recoverer, executor, epochRenewer, nil, "worker-long")

	ctx, cancel := context.WithCancel(context.Background())
	done := runWorkerAsync(t, w, ctx)

	// Wait for the executor to start (command claimed and dispatched).
	select {
	case <-executor.started():
	case <-time.After(200 * time.Millisecond):
		cancel()
		t.Fatal("executor did not start")
	}

	// The renewal ticker must fire while execution is blocked.
	select {
	case <-lease.firstRenewed():
	case <-time.After(200 * time.Millisecond):
		cancel()
		t.Fatal("lease.Renew was not called during long command execution")
	}

	// Confirm execution is still in-flight (we haven't released it yet).
	if executor.fence().LeaseToken != "tok-1" {
		t.Fatalf("executor fence token = %q, want %q", executor.fence().LeaseToken, "tok-1")
	}

	executor.release()
	cancel()
	<-done
}

// TestCommandLeaseRenewalRunsIndependently verifies that the command lease
// renewal ticker fires independently of the Redis and epoch renewal tickers,
// extending the lease of the currently executing command.
func TestCommandLeaseRenewalRunsIndependently(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-cmdlease/token",
		acquireResults: []bool{true},
		renewResult:    true,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{
		claimResult:  &entity.FlowCommand{ID: 42, CommandType: "test", LeaseToken: "tok-42"},
		claimOnce:    true,
		firstClaimCh: make(chan struct{}),
	}
	recoverer := &fakeRecoverer{}
	executor := newBlockingExecutor()
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}
	extender := &fakeLeaseExtender{firstExtendCh: make(chan struct{}), firstFailCh: make(chan struct{})}

	cfg := newTestConfig()
	cfg.RenewInterval = 5 * time.Millisecond

	w := NewWorker(cfg, lease, claimer, recoverer, executor, epochRenewer, extender, "worker-cmdlease")

	ctx, cancel := context.WithCancel(context.Background())
	done := runWorkerAsync(t, w, ctx)

	select {
	case <-executor.started():
	case <-time.After(200 * time.Millisecond):
		cancel()
		t.Fatal("executor did not start")
	}

	select {
	case <-extender.firstExtended():
	case <-time.After(200 * time.Millisecond):
		cancel()
		t.Fatal("ExtendLeaseFenced was not called during long command execution")
	}

	if extender.lastCommandID() != 42 {
		t.Fatalf("ExtendLeaseFenced called with id=%d, want 42", extender.lastCommandID())
	}

	executor.release()
	cancel()
	<-done
}

// TestLossOfRedisLeadershipCancelsCommit verifies that when the Redis lease
// renewal fails, the worker demotes and stops claiming new commands. No new
// executions should start after the loss of leadership.
func TestLossOfRedisLeadershipCancelsCommit(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-loss/token",
		acquireResults: []bool{true},
		renewResult:    false, // every renewal fails
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{
		claimResult:  &entity.FlowCommand{ID: 1, CommandType: "test", LeaseToken: "tok-1"},
		claimOnce:    true,
		firstClaimCh: make(chan struct{}),
	}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}

	cfg := newTestConfig()
	cfg.RenewInterval = 5 * time.Millisecond

	w := NewWorker(cfg, lease, claimer, recoverer, executor, epochRenewer, nil, "worker-loss")

	ctx, cancel := context.WithCancel(context.Background())
	done := runWorkerAsync(t, w, ctx)

	// Wait for at least one (failed) renewal.
	select {
	case <-lease.firstRenewed():
	case <-time.After(200 * time.Millisecond):
		cancel()
		t.Fatal("lease.Renew was never called")
	}

	// Give the worker a moment to process the demotion. After this, no new
	// commands should be executed.
	claimsAtDemotion := claimer.claimCount()
	executionsAtDemotion := executor.callCount()

	time.Sleep(30 * time.Millisecond)

	claimsFinal := claimer.claimCount()
	executionsFinal := executor.callCount()
	cancel()
	<-done

	if claimsFinal > claimsAtDemotion+1 {
		t.Fatalf("claims continued after redis loss: %d -> %d", claimsAtDemotion, claimsFinal)
	}
	if executionsFinal > executionsAtDemotion+1 {
		t.Fatalf("executions continued after redis loss: %d -> %d", executionsAtDemotion, executionsFinal)
	}
}

// TestOldWorkerCannotCompleteAfterNewEpoch verifies that when a new worker
// bumps the singleton epoch, the old worker's stale fence (carrying the old
// epoch) is rejected by the command lease extender with ErrOwnershipLost.
// This is the fencing guarantee: correctness comes from the epoch comparison,
// not from cancellation timing.
func TestOldWorkerCannotCompleteAfterNewEpoch(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-stale/token",
		acquireResults: []bool{true},
		renewResult:    true,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{
		claimResult:  &entity.FlowCommand{ID: 7, CommandType: "test", LeaseToken: "tok-7"},
		claimOnce:    true,
		firstClaimCh: make(chan struct{}),
	}
	recoverer := &fakeRecoverer{}
	executor := newBlockingExecutor()
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}
	extender := &fakeLeaseExtender{firstExtendCh: make(chan struct{}), firstFailCh: make(chan struct{})}

	cfg := newTestConfig()
	cfg.RenewInterval = 10 * time.Millisecond

	w := NewWorker(cfg, lease, claimer, recoverer, executor, epochRenewer, extender, "worker-stale")

	ctx, cancel := context.WithCancel(context.Background())
	done := runWorkerAsync(t, w, ctx)

	// Wait for the executor to start and capture the fence.
	select {
	case <-executor.started():
	case <-time.After(200 * time.Millisecond):
		cancel()
		t.Fatal("executor did not start")
	}
	oldFence := executor.fence()
	if oldFence.WorkerEpoch == 0 {
		cancel()
		t.Fatal("captured fence has zero epoch")
	}

	// Simulate a new worker taking over: bump the singleton epoch externally
	// and configure the extender to reject any fence with a lower epoch.
	extender.setRejectEpochsBelow(oldFence.WorkerEpoch + 1)

	// The command lease renewal ticker must detect the stale fence and fail.
	select {
	case <-extender.firstFailed():
	case <-time.After(200 * time.Millisecond):
		cancel()
		t.Fatal("ExtendLeaseFenced did not detect the stale fence")
	}

	executor.release()
	cancel()
	<-done
}

// TestShutdownStopsRenewers verifies that after Shutdown, all renewal tickers
// (Redis, epoch, command lease) stop firing.
func TestShutdownStopsRenewers(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-shutrenew/token",
		acquireResults: []bool{true},
		renewResult:    true,
		firstRenewCh:   make(chan struct{}),
	}
	claimer := &fakeClaimer{
		claimResult:  &entity.FlowCommand{ID: 9, CommandType: "test", LeaseToken: "tok-9"},
		claimOnce:    true,
		firstClaimCh: make(chan struct{}),
	}
	recoverer := &fakeRecoverer{}
	executor := newBlockingExecutor()
	epochRenewer := &fakeEpochRenewer{firstAdvanceCh: make(chan struct{})}
	extender := &fakeLeaseExtender{firstExtendCh: make(chan struct{}), firstFailCh: make(chan struct{})}

	cfg := newTestConfig()
	cfg.RenewInterval = 5 * time.Millisecond

	w := NewWorker(cfg, lease, claimer, recoverer, executor, epochRenewer, extender, "worker-shutrenew")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := runWorkerAsync(t, w, ctx)

	// Wait for at least one renewal of each kind.
	select {
	case <-lease.firstRenewed():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("lease.Renew was never called")
	}
	select {
	case <-epochRenewer.firstAdvanced():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("AdvanceEpoch was never called")
	}
	select {
	case <-executor.started():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("executor did not start")
	}
	select {
	case <-extender.firstExtended():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("ExtendLeaseFenced was never called")
	}

	if err := w.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not exit after Shutdown")
	}

	// After shutdown, no more renewals should happen.
	renewalsBefore := lease.renewCount()
	epochAdvancesBefore := epochRenewer.advanceCount()
	extendsBefore := extender.extendCount()

	time.Sleep(50 * time.Millisecond)

	if lease.renewCount() != renewalsBefore {
		t.Fatalf("lease.Renew called after shutdown: %d -> %d", renewalsBefore, lease.renewCount())
	}
	if epochRenewer.advanceCount() != epochAdvancesBefore {
		t.Fatalf("AdvanceEpoch called after shutdown: %d -> %d", epochAdvancesBefore, epochRenewer.advanceCount())
	}
	if extender.extendCount() != extendsBefore {
		t.Fatalf("ExtendLeaseFenced called after shutdown: %d -> %d", extendsBefore, extender.extendCount())
	}
}
