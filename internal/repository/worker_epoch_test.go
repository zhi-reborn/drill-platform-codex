package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupWorkerEpochRepo prepares an in-memory SQLite database with both the
// drill_flow_command and drill_worker_epoch tables so the fencing tests can
// exercise ClaimNext together with the fenced ownership methods.
func setupWorkerEpochRepo(t *testing.T) (*WorkerEpochRepo, *FlowCommandRepo, *gorm.DB) {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&entity.FlowCommand{}, &entity.WorkerEpoch{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return NewWorkerEpochRepo(db), NewFlowCommandRepo(db), db
}

func TestAdvanceEpochReturnsIncreasingValue(t *testing.T) {
	epochRepo, _, _ := setupWorkerEpochRepo(t)
	ctx := context.Background()
	ttl := 30 * time.Second

	first, err := epochRepo.AdvanceEpoch(ctx, "worker-A", ttl)
	if err != nil {
		t.Fatalf("first AdvanceEpoch: %v", err)
	}
	if first.Epoch == 0 {
		t.Fatalf("first epoch = 0, want > 0")
	}
	if first.WorkerID != "worker-A" {
		t.Fatalf("first worker_id = %q, want %q", first.WorkerID, "worker-A")
	}
	if first.LeaseUntil == nil {
		t.Fatalf("first lease_until is nil, want non-nil")
	}

	second, err := epochRepo.AdvanceEpoch(ctx, "worker-A", ttl)
	if err != nil {
		t.Fatalf("second AdvanceEpoch: %v", err)
	}
	if second.Epoch <= first.Epoch {
		t.Fatalf("second epoch = %d, want > %d", second.Epoch, first.Epoch)
	}

	third, err := epochRepo.AdvanceEpoch(ctx, "worker-A", ttl)
	if err != nil {
		t.Fatalf("third AdvanceEpoch: %v", err)
	}
	if third.Epoch <= second.Epoch {
		t.Fatalf("third epoch = %d, want > %d", third.Epoch, second.Epoch)
	}
}

func TestAdvanceEpochTransfersWorkerID(t *testing.T) {
	epochRepo, _, _ := setupWorkerEpochRepo(t)
	ctx := context.Background()
	ttl := 30 * time.Second

	first, err := epochRepo.AdvanceEpoch(ctx, "worker-A", ttl)
	if err != nil {
		t.Fatalf("first AdvanceEpoch: %v", err)
	}
	if first.WorkerID != "worker-A" {
		t.Fatalf("first worker_id = %q, want %q", first.WorkerID, "worker-A")
	}

	second, err := epochRepo.AdvanceEpoch(ctx, "worker-B", ttl)
	if err != nil {
		t.Fatalf("second AdvanceEpoch: %v", err)
	}
	if second.WorkerID != "worker-B" {
		t.Fatalf("second worker_id = %q, want %q (transfer failed)", second.WorkerID, "worker-B")
	}
	if second.Epoch <= first.Epoch {
		t.Fatalf("second epoch = %d, want > %d", second.Epoch, first.Epoch)
	}

	third, err := epochRepo.AdvanceEpoch(ctx, "worker-A", ttl)
	if err != nil {
		t.Fatalf("third AdvanceEpoch: %v", err)
	}
	if third.WorkerID != "worker-A" {
		t.Fatalf("third worker_id = %q, want %q (transfer back failed)", third.WorkerID, "worker-A")
	}
}

func TestClaimNextStoresOwnershipFields(t *testing.T) {
	_, cmdRepo, _ := setupWorkerEpochRepo(t)

	cmd, _, err := cmdRepo.CreateOrGet(newFlowCommand("claim-ownership-1", 10))
	if err != nil {
		t.Fatalf("CreateOrGet: %v", err)
	}

	ctx := context.Background()
	lease := 60 * time.Second
	claimed, err := cmdRepo.ClaimNext(ctx, "worker-A", lease)
	if err != nil {
		t.Fatalf("ClaimNext: %v", err)
	}
	if claimed == nil {
		t.Fatal("ClaimNext returned nil")
	}
	if claimed.ID != cmd.ID {
		t.Fatalf("claimed ID = %d, want %d", claimed.ID, cmd.ID)
	}

	if claimed.WorkerID == nil || *claimed.WorkerID != "worker-A" {
		t.Fatalf("worker_id = %v, want %q", claimed.WorkerID, "worker-A")
	}
	if claimed.LeaseToken == "" {
		t.Fatal("lease_token is empty, want non-empty token")
	}
	if claimed.LeaseUntil == nil {
		t.Fatal("lease_until is nil, want non-nil")
	}
	if claimed.WorkerEpoch == 0 {
		t.Fatal("worker_epoch = 0, want > 0 (epoch must be stamped on claim)")
	}
}

func TestMarkSucceededFencedRequiresMatchingOwnership(t *testing.T) {
	_, cmdRepo, _ := setupWorkerEpochRepo(t)

	if _, _, err := cmdRepo.CreateOrGet(newFlowCommand("fenced-success-1", 10)); err != nil {
		t.Fatalf("CreateOrGet: %v", err)
	}

	ctx := context.Background()
	claimed, err := cmdRepo.ClaimNext(ctx, "worker-A", 60*time.Second)
	if err != nil {
		t.Fatalf("ClaimNext: %v", err)
	}

	correct := CommandOwnership{
		WorkerID: "worker-A",
		Epoch:    claimed.WorkerEpoch,
		Token:    claimed.LeaseToken,
	}

	// Wrong worker_id: must return ErrOwnershipLost and update zero rows.
	wrongWorker := CommandOwnership{
		WorkerID: "worker-B",
		Epoch:    claimed.WorkerEpoch,
		Token:    claimed.LeaseToken,
	}
	if err := cmdRepo.MarkSucceededFenced(ctx, claimed.ID, wrongWorker, map[string]any{"ok": true}); !errors.Is(err, ErrOwnershipLost) {
		t.Fatalf("wrong worker err = %v, want ErrOwnershipLost", err)
	}
	got, err := cmdRepo.FindByID(claimed.ID)
	if err != nil {
		t.Fatalf("FindByID after wrong ownership: %v", err)
	}
	if got.Status != entity.FlowCommandProcessing {
		t.Fatalf("status = %s, want %s (zero rows should be updated)", got.Status, entity.FlowCommandProcessing)
	}

	// Wrong token: must return ErrOwnershipLost.
	wrongToken := CommandOwnership{
		WorkerID: "worker-A",
		Epoch:    claimed.WorkerEpoch,
		Token:    "deadbeef-not-the-real-token",
	}
	if err := cmdRepo.MarkSucceededFenced(ctx, claimed.ID, wrongToken, map[string]any{"ok": true}); !errors.Is(err, ErrOwnershipLost) {
		t.Fatalf("wrong token err = %v, want ErrOwnershipLost", err)
	}

	// Wrong epoch: must return ErrOwnershipLost.
	wrongEpoch := CommandOwnership{
		WorkerID: "worker-A",
		Epoch:    claimed.WorkerEpoch + 1,
		Token:    claimed.LeaseToken,
	}
	if err := cmdRepo.MarkSucceededFenced(ctx, claimed.ID, wrongEpoch, map[string]any{"ok": true}); !errors.Is(err, ErrOwnershipLost) {
		t.Fatalf("wrong epoch err = %v, want ErrOwnershipLost", err)
	}

	// Correct ownership: must succeed and flip status.
	if err := cmdRepo.MarkSucceededFenced(ctx, claimed.ID, correct, map[string]any{"ok": true}); err != nil {
		t.Fatalf("MarkSucceededFenced correct ownership: %v", err)
	}
	got, err = cmdRepo.FindByID(claimed.ID)
	if err != nil {
		t.Fatalf("FindByID after correct ownership: %v", err)
	}
	if got.Status != entity.FlowCommandSucceeded {
		t.Fatalf("status = %s, want %s", got.Status, entity.FlowCommandSucceeded)
	}
	if got.FinishedAt == nil {
		t.Fatal("finished_at is nil, want non-nil")
	}
}

func TestMarkFailedFencedRequiresMatchingOwnership(t *testing.T) {
	_, cmdRepo, _ := setupWorkerEpochRepo(t)

	if _, _, err := cmdRepo.CreateOrGet(newFlowCommand("fenced-fail-1", 10)); err != nil {
		t.Fatalf("CreateOrGet: %v", err)
	}

	ctx := context.Background()
	claimed, err := cmdRepo.ClaimNext(ctx, "worker-A", 60*time.Second)
	if err != nil {
		t.Fatalf("ClaimNext: %v", err)
	}

	correct := CommandOwnership{
		WorkerID: "worker-A",
		Epoch:    claimed.WorkerEpoch,
		Token:    claimed.LeaseToken,
	}

	// Wrong ownership: must return ErrOwnershipLost and update zero rows.
	wrong := CommandOwnership{
		WorkerID: "worker-B",
		Epoch:    claimed.WorkerEpoch,
		Token:    claimed.LeaseToken,
	}
	if err := cmdRepo.MarkFailedFenced(ctx, claimed.ID, wrong, "CODE", "msg"); !errors.Is(err, ErrOwnershipLost) {
		t.Fatalf("wrong ownership err = %v, want ErrOwnershipLost", err)
	}
	got, err := cmdRepo.FindByID(claimed.ID)
	if err != nil {
		t.Fatalf("FindByID after wrong ownership: %v", err)
	}
	if got.Status != entity.FlowCommandProcessing {
		t.Fatalf("status = %s, want %s (zero rows should be updated)", got.Status, entity.FlowCommandProcessing)
	}

	// Correct ownership: must succeed.
	if err := cmdRepo.MarkFailedFenced(ctx, claimed.ID, correct, "FLOW_ERROR", "flow failed"); err != nil {
		t.Fatalf("MarkFailedFenced correct ownership: %v", err)
	}
	got, err = cmdRepo.FindByID(claimed.ID)
	if err != nil {
		t.Fatalf("FindByID after correct ownership: %v", err)
	}
	if got.Status != entity.FlowCommandFailed {
		t.Fatalf("status = %s, want %s", got.Status, entity.FlowCommandFailed)
	}
	if got.ErrorCode == nil || *got.ErrorCode != "FLOW_ERROR" {
		t.Fatalf("error_code = %v, want FLOW_ERROR", got.ErrorCode)
	}
	if got.ErrorMessage == nil || *got.ErrorMessage != "flow failed" {
		t.Fatalf("error_message = %v, want flow failed", got.ErrorMessage)
	}
}

func TestExtendLeaseFencedRequiresMatchingOwnership(t *testing.T) {
	_, cmdRepo, _ := setupWorkerEpochRepo(t)

	if _, _, err := cmdRepo.CreateOrGet(newFlowCommand("fenced-extend-1", 10)); err != nil {
		t.Fatalf("CreateOrGet: %v", err)
	}

	ctx := context.Background()
	claimed, err := cmdRepo.ClaimNext(ctx, "worker-A", 60*time.Second)
	if err != nil {
		t.Fatalf("ClaimNext: %v", err)
	}

	correct := CommandOwnership{
		WorkerID: "worker-A",
		Epoch:    claimed.WorkerEpoch,
		Token:    claimed.LeaseToken,
	}

	// Wrong ownership: must return (false, ErrOwnershipLost).
	wrong := CommandOwnership{
		WorkerID: "worker-B",
		Epoch:    claimed.WorkerEpoch,
		Token:    claimed.LeaseToken,
	}
	ok, err := cmdRepo.ExtendLeaseFenced(ctx, claimed.ID, wrong, time.Now().Add(2*time.Minute))
	if !errors.Is(err, ErrOwnershipLost) {
		t.Fatalf("wrong ownership err = %v, want ErrOwnershipLost", err)
	}
	if ok {
		t.Fatal("wrong ownership ok = true, want false")
	}

	// Correct ownership: must return (true, nil).
	newUntil := time.Now().Add(2 * time.Minute)
	ok, err = cmdRepo.ExtendLeaseFenced(ctx, claimed.ID, correct, newUntil)
	if err != nil {
		t.Fatalf("ExtendLeaseFenced correct ownership: %v", err)
	}
	if !ok {
		t.Fatal("ExtendLeaseFenced correct ownership ok = false, want true")
	}
	got, err := cmdRepo.FindByID(claimed.ID)
	if err != nil {
		t.Fatalf("FindByID after extend: %v", err)
	}
	if got.LeaseUntil == nil {
		t.Fatal("lease_until is nil after extend")
	}
}

func TestStaleOwnershipUpdatesZeroRows(t *testing.T) {
	_, cmdRepo, db := setupWorkerEpochRepo(t)

	if _, _, err := cmdRepo.CreateOrGet(newFlowCommand("stale-1", 10)); err != nil {
		t.Fatalf("CreateOrGet: %v", err)
	}

	ctx := context.Background()
	claimed, err := cmdRepo.ClaimNext(ctx, "worker-A", 60*time.Second)
	if err != nil {
		t.Fatalf("ClaimNext: %v", err)
	}

	ownership := CommandOwnership{
		WorkerID: "worker-A",
		Epoch:    claimed.WorkerEpoch,
		Token:    claimed.LeaseToken,
	}

	// Expire the lease by moving lease_until into the past. The fenced WHERE
	// clause (lease_until > now) must reject every operation.
	pastTime := time.Now().Add(-time.Minute)
	if err := db.Model(&entity.FlowCommand{}).Where("id = ?", claimed.ID).Update("lease_until", pastTime).Error; err != nil {
		t.Fatalf("set expired lease: %v", err)
	}

	if err := cmdRepo.MarkSucceededFenced(ctx, claimed.ID, ownership, map[string]any{"ok": true}); !errors.Is(err, ErrOwnershipLost) {
		t.Fatalf("MarkSucceededFenced expired lease err = %v, want ErrOwnershipLost", err)
	}
	got, _ := cmdRepo.FindByID(claimed.ID)
	if got.Status != entity.FlowCommandProcessing {
		t.Fatalf("status = %s, want %s (expired lease must not flip status)", got.Status, entity.FlowCommandProcessing)
	}

	if err := cmdRepo.MarkFailedFenced(ctx, claimed.ID, ownership, "CODE", "msg"); !errors.Is(err, ErrOwnershipLost) {
		t.Fatalf("MarkFailedFenced expired lease err = %v, want ErrOwnershipLost", err)
	}

	ok, err := cmdRepo.ExtendLeaseFenced(ctx, claimed.ID, ownership, time.Now().Add(time.Minute))
	if !errors.Is(err, ErrOwnershipLost) {
		t.Fatalf("ExtendLeaseFenced expired lease err = %v, want ErrOwnershipLost", err)
	}
	if ok {
		t.Fatal("ExtendLeaseFenced expired lease ok = true, want false")
	}
}
