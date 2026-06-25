//go:build integration

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
	redisinfra "drill-platform/internal/infrastructure/redis"
	"drill-platform/internal/repository"
	"drill-platform/internal/worker"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// --- Redis helpers (shared with websocket_pubsub_test.go) ---

// connectRedisRaw opens a go-redis client using REDIS_ADDR env var. Tests call
// t.Fatal if Redis is unreachable so the failure surfaces clearly.
func connectRedisRaw(t *testing.T) *goredis.Client {
	t.Helper()
	addr := redisAddr()
	rdb := goredis.NewClient(&goredis.Options{
		Addr:         addr,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		t.Fatalf("connect Redis %s: %v (is docker-compose.ha.yml running?)", addr, err)
	}
	return rdb
}

// redisLeaseStore adapts *goredis.Client to the redis.LeaseStore interface so
// integration tests can construct a redis.Lease without the package-level
// Client wrapper.
type redisLeaseStore struct {
	client *goredis.Client
}

func (s *redisLeaseStore) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, expiration).Result()
}

func (s *redisLeaseStore) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return s.client.Eval(ctx, script, keys, args...).Result()
}

// --- Drill instance helpers ---

// setupDrillInstanceSchema ensures the drill_instance table exists for
// failover recovery tests. Uses DisableForeignKeyConstraintWhenMigrating set in
// connectMySQL so related tables are not required.
func setupDrillInstanceSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&entity.DrillInstance{}); err != nil {
		t.Fatalf("auto migrate drill_instance: %v", err)
	}
}

// createRunningDrill inserts a drill instance with the given status and returns
// its ID. The drill is deleted after the test via t.Cleanup.
func createRunningDrill(t *testing.T, db *gorm.DB, name string) uint64 {
	t.Helper()
	drill := &entity.DrillInstance{
		TemplateID: 1,
		Name:       name,
		Status:     "running",
		CreatedBy:  1,
	}
	if err := db.Create(drill).Error; err != nil {
		t.Fatalf("create running drill: %v", err)
	}
	drillID := drill.ID
	t.Cleanup(func() {
		db.Exec("DELETE FROM drill_instance WHERE id = ?", drillID)
	})
	return drillID
}

// --- Fakes used by the failover tests ---

// drillRecoverer implements worker.Recoverer. It loads running/paused drills
// from MySQL so the test can assert that the standby recovered the in-flight
// drill after failover.
type drillRecoverer struct {
	mu           sync.Mutex
	called       bool
	recoveredIDs []uint64
	db           *gorm.DB
}

func newDrillRecoverer(db *gorm.DB) *drillRecoverer {
	return &drillRecoverer{db: db}
}

func (r *drillRecoverer) Recover(ctx context.Context) error {
	var drills []entity.DrillInstance
	if err := r.db.WithContext(ctx).Where("status IN ?", []string{"running", "paused"}).Find(&drills).Error; err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.called = true
	r.recoveredIDs = r.recoveredIDs[:0]
	for _, d := range drills {
		r.recoveredIDs = append(r.recoveredIDs, d.ID)
	}
	return nil
}

func (r *drillRecoverer) wasCalled() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.called
}

func (r *drillRecoverer) getRecoveredIDs() []uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := make([]uint64, len(r.recoveredIDs))
	copy(ids, r.recoveredIDs)
	return ids
}

// succeedingExecutor implements worker.Executor. It marks the claimed command
// as succeeded via the repo, simulating a real executor without the full
// flow-engine dependency graph.
type succeedingExecutor struct {
	repo   *repository.FlowCommandRepo
	mu     sync.Mutex
	called bool
	cmdID  uint64
}

func newSucceedingExecutor(repo *repository.FlowCommandRepo) *succeedingExecutor {
	return &succeedingExecutor{repo: repo}
}

func (e *succeedingExecutor) Execute(ctx context.Context, cmd *entity.FlowCommand) error {
	e.mu.Lock()
	e.called = true
	e.cmdID = cmd.ID
	e.mu.Unlock()
	return e.repo.MarkSucceeded(cmd.ID, map[string]any{"ok": true})
}

func (e *succeedingExecutor) wasCalled() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.called
}

// --- Worker helpers ---

// failoverConfig returns a Worker config with short timings so the failover
// test completes within a few seconds instead of the production 15s lease.
func failoverConfig() worker.Config {
	return worker.Config{
		LeaseTTL:      2 * time.Second,
		RenewInterval: 500 * time.Millisecond,
		CommandLease:  5 * time.Second,
		IdlePoll:      100 * time.Millisecond,
	}
}

// runWorkerAsync starts the Worker's Run loop in a goroutine and returns a
// done channel that is closed when Run returns.
func runWorkerAsync(t *testing.T, w *worker.Worker, ctx context.Context) <-chan struct{} {
	t.Helper()
	done := make(chan struct{})
	go func() {
		_ = w.Run(ctx)
		close(done)
	}()
	return done
}

// waitForStatus polls the worker status until it matches want or the deadline
// expires.
func waitForStatus(t *testing.T, w *worker.Worker, want worker.Status, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if w.Status() == want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("worker status = %s, want %s (timeout %s)", w.Status(), want, timeout)
}

// --- Failover tests ---

// TestTwoWorkersExactlyOneLeader starts two Workers against the same Redis
// lease key and asserts exactly one reaches leader-ready while the other stays
// standby.
func TestTwoWorkersExactlyOneLeader(t *testing.T) {
	db := connectMySQL(t)
	setupFlowCommandSchema(t, db)
	prepareFlowCommands(t, db)

	rdb := connectRedisRaw(t)
	defer rdb.Close()

	leaseKey := "drill:worker:leader:test-two-leaders"
	ctx := context.Background()
	if err := rdb.Del(ctx, leaseKey).Err(); err != nil {
		t.Fatalf("delete lease key: %v", err)
	}

	repo := repository.NewFlowCommandRepo(db)
	store := &redisLeaseStore{client: rdb}

	config := failoverConfig()
	leaseA := redisinfra.NewLease(store, leaseKey, "worker-a", config.LeaseTTL)
	leaseB := redisinfra.NewLease(store, leaseKey, "worker-b", config.LeaseTTL)

	wA := worker.NewWorker(config, leaseA, repo, nil, nil, "worker-a")
	wB := worker.NewWorker(config, leaseB, repo, nil, nil, "worker-b")

	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())
	defer cancelA()
	defer cancelB()

	doneA := runWorkerAsync(t, wA, ctxA)
	doneB := runWorkerAsync(t, wB, ctxB)

	// Wait for one worker to win leadership.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if wA.Status() == worker.StatusLeaderReady || wB.Status() == worker.StatusLeaderReady {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	leaderCount := 0
	if wA.Status() == worker.StatusLeaderReady {
		leaderCount++
	}
	if wB.Status() == worker.StatusLeaderReady {
		leaderCount++
	}
	if leaderCount != 1 {
		t.Fatalf("expected exactly 1 leader, got %d (A=%s B=%s)", leaderCount, wA.Status(), wB.Status())
	}

	// Shut down both workers.
	cancelA()
	cancelB()
	<-doneA
	<-doneB
}

// TestFailoverRecoversRunningDrill verifies that when the leader Worker is
// cancelled (simulating a crash), the standby acquires the lease after TTL
// expiry, recovers the running drill, and completes the pending command.
func TestFailoverRecoversRunningDrill(t *testing.T) {
	db := connectMySQL(t)
	setupFlowCommandSchema(t, db)
	prepareFlowCommands(t, db)
	setupDrillInstanceSchema(t, db)

	rdb := connectRedisRaw(t)
	defer rdb.Close()

	leaseKey := "drill:worker:leader:test-failover"
	ctx := context.Background()
	if err := rdb.Del(ctx, leaseKey).Err(); err != nil {
		t.Fatalf("delete lease key: %v", err)
	}

	// Seed a running drill and a pending command for it.
	drillID := createRunningDrill(t, db, "failover-test-drill")
	repo := repository.NewFlowCommandRepo(db)
	cmd, _, err := repo.CreateOrGet(newPendingCommand(uniqueKey("failover-cmd"), drillID))
	if err != nil {
		t.Fatalf("create pending command: %v", err)
	}
	commandID := cmd.ID

	config := failoverConfig()
	store := &redisLeaseStore{client: rdb}

	recovererA := newDrillRecoverer(db)
	recovererB := newDrillRecoverer(db)
	executorA := newSucceedingExecutor(repo)
	executorB := newSucceedingExecutor(repo)

	leaseA := redisinfra.NewLease(store, leaseKey, "worker-a", config.LeaseTTL)
	leaseB := redisinfra.NewLease(store, leaseKey, "worker-b", config.LeaseTTL)

	wA := worker.NewWorker(config, leaseA, repo, recovererA, executorA, "worker-a")
	wB := worker.NewWorker(config, leaseB, repo, recovererB, executorB, "worker-b")

	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())

	doneA := runWorkerAsync(t, wA, ctxA)
	doneB := runWorkerAsync(t, wB, ctxB)

	// Identify which worker became the leader.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if wA.Status() == worker.StatusLeaderReady || wB.Status() == worker.StatusLeaderReady {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	var (
		leaderCtxCancel  context.CancelFunc
		standbyWorker    *worker.Worker
		standbyCancel    context.CancelFunc
		standbyRecoverer *drillRecoverer
		standbyDone      <-chan struct{}
		leaderDone       <-chan struct{}
	)
	if wA.Status() == worker.StatusLeaderReady {
		leaderCtxCancel = cancelA
		leaderDone = doneA
		standbyWorker = wB
		standbyCancel = cancelB
		standbyRecoverer = recovererB
		standbyDone = doneB
	} else if wB.Status() == worker.StatusLeaderReady {
		leaderCtxCancel = cancelB
		leaderDone = doneB
		standbyWorker = wA
		standbyCancel = cancelA
		standbyRecoverer = recovererA
		standbyDone = doneA
	} else {
		cancelA()
		cancelB()
		t.Fatalf("neither worker became leader (A=%s B=%s)", wA.Status(), wB.Status())
	}

	// Cancel the leader context WITHOUT calling Shutdown so the Redis lease is
	// not released. The standby must wait for the lease TTL to expire before
	// acquiring leadership.
	leaderCtxCancel()
	<-leaderDone

	// Wait for the standby to acquire leadership after lease expiry.
	// LeaseTTL is 2s; allow generous headroom for the renewal/acquisition cycle.
	waitForStatus(t, standbyWorker, worker.StatusLeaderReady, 10*time.Second)

	// Assert the standby recovered the running drill.
	if !standbyRecoverer.wasCalled() {
		t.Fatal("standby recoverer was not called after failover")
	}
	recoveredIDs := standbyRecoverer.getRecoveredIDs()
	found := false
	for _, id := range recoveredIDs {
		if id == drillID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("drill %d not recovered by standby; recovered IDs: %v", drillID, recoveredIDs)
	}

	// Wait for the pending command to be executed by the standby.
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		got, err := repo.FindByID(commandID)
		if err != nil {
			t.Fatalf("find command %d: %v", commandID, err)
		}
		if got.IsTerminal() {
			if got.Status != entity.FlowCommandSucceeded {
				t.Fatalf("command %d status = %s, want %s", commandID, got.Status, entity.FlowCommandSucceeded)
			}
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	got, err := repo.FindByID(commandID)
	if err != nil {
		t.Fatalf("final find command %d: %v", commandID, err)
	}
	if got.Status != entity.FlowCommandSucceeded {
		t.Fatalf("command %d status = %s, want %s", commandID, got.Status, entity.FlowCommandSucceeded)
	}

	// Clean up the standby.
	standbyCancel()
	select {
	case <-standbyDone:
	case <-time.After(3 * time.Second):
		t.Fatal("standby worker did not exit after cancel")
	}
}
