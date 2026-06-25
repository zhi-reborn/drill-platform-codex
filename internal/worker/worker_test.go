package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
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

type fakeClaimer struct {
	mu           sync.Mutex
	claimResult  *entity.FlowCommand
	claimErr     error
	claimCalls   int
	firstClaimAt time.Time
	requeueCalls int
}

func (f *fakeClaimer) ClaimNext(ctx context.Context, workerID string, lease time.Duration) (*entity.FlowCommand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.claimCalls++
	if f.firstClaimAt.IsZero() {
		f.firstClaimAt = time.Now()
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
	mu      sync.Mutex
	calls   int
	lastCmd *entity.FlowCommand
	err     error
}

func (f *fakeExecutor) Execute(ctx context.Context, cmd *entity.FlowCommand) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.lastCmd = cmd
	return f.err
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

// --- Tests ---

func TestStandbyDoesNotClaim(t *testing.T) {
	lease := &fakeLease{
		value:          "worker-standby/token",
		acquireResults: []bool{},
		renewResult:    true,
	}
	claimer := &fakeClaimer{claimResult: nil}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, "worker-standby")

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
	}
	claimer := &fakeClaimer{claimResult: nil}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, "worker-rec")

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
	}
	claimer := &fakeClaimer{claimResult: nil}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, "worker-renew")

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
	}
	claimer := &fakeClaimer{claimResult: nil}
	recoverer := &fakeRecoverer{}
	executor := &fakeExecutor{}

	w := NewWorker(newTestConfig(), lease, claimer, recoverer, executor, "worker-shutdown")

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
