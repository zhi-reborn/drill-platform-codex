//go:build integration

// Package integration contains end-to-end integration tests that require real
// MySQL and Redis instances. The tests are gated behind the "integration" build
// tag so they do not run during normal `go test ./...`. Use
// `go test -tags=integration ./internal/integration` to run them, typically via
// scripts/test-ha.sh which provisions the dependencies via docker-compose.ha.yml.
package integration

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Shared environment helpers ---

// envOr returns the named environment variable or fallback when unset/empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// mysqlDSN builds a MySQL DSN from DATABASE_* environment variables. Defaults
// match docker-compose.ha.yml host port mappings.
func mysqlDSN() string {
	user := envOr("DATABASE_USER", "drill")
	password := envOr("DATABASE_PASSWORD", "drill123")
	host := envOr("DATABASE_HOST", "127.0.0.1")
	port := envOr("DATABASE_PORT", "13306")
	name := envOr("DATABASE_NAME", "drill_platform")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, name)
}

// redisAddr returns the Redis address from REDIS_ADDR. Default matches
// docker-compose.ha.yml host port mapping.
func redisAddr() string {
	return envOr("REDIS_ADDR", "127.0.0.1:16379")
}

// connectMySQL opens a GORM connection to MySQL using DATABASE_* env vars. The
// connection disables foreign key constraint migration so AutoMigrate works
// against partial schemas in test databases. Tests call t.Fatal if MySQL is
// unreachable so the failure surfaces clearly when dependencies are missing.
func connectMySQL(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := mysqlDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		host := envOr("DATABASE_HOST", "127.0.0.1")
		port := envOr("DATABASE_PORT", "13306")
		name := envOr("DATABASE_NAME", "drill_platform")
		t.Fatalf("connect MySQL %s:%s/%s: %v (is docker-compose.ha.yml running?)", host, port, name, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	return db
}

// setupFlowCommandSchema ensures the drill_flow_command table exists.
func setupFlowCommandSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&entity.FlowCommand{}); err != nil {
		t.Fatalf("auto migrate flow_command: %v", err)
	}
}

// prepareFlowCommands cleans the flow_command table before and after the test
// so each test starts from a known empty state.
func prepareFlowCommands(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("DELETE FROM drill_flow_command").Error; err != nil {
		t.Fatalf("cleanup flow_command before test: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Exec("DELETE FROM drill_flow_command").Error; err != nil {
			t.Errorf("cleanup flow_command after test: %v", err)
		}
	})
}

// newPendingCommand constructs a pending FlowCommand with the given key and drill.
func newPendingCommand(key string, drillID uint64) *entity.FlowCommand {
	return &entity.FlowCommand{
		CommandType:     "start_drill",
		DrillInstanceID: drillID,
		OperatorID:      1,
		IdempotencyKey:  key,
		Payload:         `{}`,
		Status:          entity.FlowCommandPending,
	}
}

// uniqueKey returns an idempotency key that is unique across test runs.
func uniqueKey(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), time.Now().Nanosecond())
}

// --- MySQL integration tests ---

// TestConcurrentCreateOrGetReturnsOneCommand verifies that concurrent
// CreateOrGet calls with the same idempotency key all resolve to the same
// command ID. The MySQL unique index on idempotency_key is the source of truth.
func TestConcurrentCreateOrGetReturnsOneCommand(t *testing.T) {
	db := connectMySQL(t)
	setupFlowCommandSchema(t, db)
	prepareFlowCommands(t, db)
	repo := repository.NewFlowCommandRepo(db)

	key := uniqueKey("concurrent-create")
	const goroutines = 10
	results := make([]*entity.FlowCommand, goroutines)
	errs := make([]error, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			result, _, err := repo.CreateOrGet(newPendingCommand(key, 9001))
			results[idx] = result
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d error: %v", i, err)
		}
	}

	var firstID uint64
	for i, r := range results {
		if r == nil {
			t.Fatalf("goroutine %d returned nil result", i)
		}
		if i == 0 {
			firstID = r.ID
		}
		if r.ID != firstID {
			t.Fatalf("goroutine %d returned command ID %d, want %d (duplicate creation detected)", i, r.ID, firstID)
		}
	}
}

// TestTwoClaimersCannotClaimSameCommand verifies that two concurrent ClaimNext
// calls cannot both claim the same pending command. MySQL's FOR UPDATE SKIP
// LOCKED ensures only one claimer wins.
func TestTwoClaimersCannotClaimSameCommand(t *testing.T) {
	db := connectMySQL(t)
	setupFlowCommandSchema(t, db)
	prepareFlowCommands(t, db)
	repo := repository.NewFlowCommandRepo(db)

	if _, _, err := repo.CreateOrGet(newPendingCommand(uniqueKey("claim-exclusive"), 9002)); err != nil {
		t.Fatalf("seed command: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		mu        sync.Mutex
		claimed   []*entity.FlowCommand
		claimErrs []error
	)
	var wg sync.WaitGroup
	wg.Add(2)
	start := make(chan struct{})

	for i := 0; i < 2; i++ {
		workerID := fmt.Sprintf("claimer-%d", i)
		go func(wid string) {
			defer wg.Done()
			<-start
			cmd, err := repo.ClaimNext(ctx, wid, time.Minute)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				claimErrs = append(claimErrs, err)
				return
			}
			if cmd != nil {
				claimed = append(claimed, cmd)
			}
		}(workerID)
	}
	close(start)
	wg.Wait()

	for _, err := range claimErrs {
		if err != nil {
			t.Fatalf("claim error: %v", err)
		}
	}

	if len(claimed) != 1 {
		t.Fatalf("expected exactly 1 claim, got %d", len(claimed))
	}
}

// TestExpiredProcessingCommandsReclaimed verifies that a processing command
// whose lease has expired is requeued by RequeueExpired and can subsequently be
// claimed by a new worker.
func TestExpiredProcessingCommandsReclaimed(t *testing.T) {
	db := connectMySQL(t)
	setupFlowCommandSchema(t, db)
	prepareFlowCommands(t, db)
	repo := repository.NewFlowCommandRepo(db)

	cmd, _, err := repo.CreateOrGet(newPendingCommand(uniqueKey("expired-reclaim"), 9003))
	if err != nil {
		t.Fatalf("create command: %v", err)
	}

	// Simulate a worker that claimed the command but died before completing it,
	// leaving an expired lease.
	expiredLease := time.Now().Add(-time.Minute)
	if err := db.Model(&entity.FlowCommand{}).Where("id = ?", cmd.ID).Updates(map[string]any{
		"status":      entity.FlowCommandProcessing,
		"worker_id":   "dead-worker",
		"lease_until": expiredLease,
	}).Error; err != nil {
		t.Fatalf("set expired processing: %v", err)
	}

	count, err := repo.RequeueExpired(time.Now())
	if err != nil {
		t.Fatalf("requeue expired: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected at least 1 requeued command, got %d", count)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	claimed, err := repo.ClaimNext(ctx, "new-worker", time.Minute)
	if err != nil {
		t.Fatalf("claim after requeue: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected to reclaim the expired command, got nil")
	}
	if claimed.ID != cmd.ID {
		t.Fatalf("claimed command ID %d, want %d", claimed.ID, cmd.ID)
	}
	if claimed.Attempts < 2 {
		t.Fatalf("attempts = %d, want >= 2 (one original claim plus reclaim)", claimed.Attempts)
	}
}

// TestCommandsForTwoDrillsProceedIndependently verifies that pending commands
// for two different drills can both be claimed, proving the command queue does
// not serialize unrelated drills.
func TestCommandsForTwoDrillsProceedIndependently(t *testing.T) {
	db := connectMySQL(t)
	setupFlowCommandSchema(t, db)
	prepareFlowCommands(t, db)
	repo := repository.NewFlowCommandRepo(db)

	base := uniqueKey("drill-pair")
	drillA := uint64(9100)
	drillB := uint64(9101)

	cmdA, _, err := repo.CreateOrGet(newPendingCommand(base+"-a", drillA))
	if err != nil {
		t.Fatalf("create command A: %v", err)
	}
	cmdB, _, err := repo.CreateOrGet(newPendingCommand(base+"-b", drillB))
	if err != nil {
		t.Fatalf("create command B: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	claimedA, err := repo.ClaimNext(ctx, "worker-a", time.Minute)
	if err != nil {
		t.Fatalf("claim A: %v", err)
	}
	claimedB, err := repo.ClaimNext(ctx, "worker-b", time.Minute)
	if err != nil {
		t.Fatalf("claim B: %v", err)
	}
	if claimedA == nil || claimedB == nil {
		t.Fatalf("expected both commands claimed: A=%v B=%v", claimedA, claimedB)
	}

	ids := map[uint64]bool{claimedA.ID: true, claimedB.ID: true}
	if len(ids) != 2 {
		t.Fatalf("expected 2 distinct claimed commands, got %d (A=%d B=%d)", len(ids), claimedA.ID, claimedB.ID)
	}
	if !ids[cmdA.ID] || !ids[cmdB.ID] {
		t.Fatalf("claimed IDs do not match seeded commands: got %v, want cmdA=%d cmdB=%d", ids, cmdA.ID, cmdB.ID)
	}
}
