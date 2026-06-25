package repository

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFlowCommandRepo(t *testing.T) *FlowCommandRepo {
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

	if err := db.AutoMigrate(&entity.FlowCommand{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return NewFlowCommandRepo(db)
}

func newFlowCommand(key string, operatorID uint64) *entity.FlowCommand {
	return &entity.FlowCommand{
		CommandType:     "start_drill",
		DrillInstanceID: 1001,
		OperatorID:      operatorID,
		IdempotencyKey:  key,
		Payload:         `{"action":"start"}`,
		Status:          entity.FlowCommandPending,
	}
}

func TestFlowCommandRepoCreateOrGetIsIdempotentByIdempotencyKey(t *testing.T) {
	repo := setupFlowCommandRepo(t)

	first, created, err := repo.CreateOrGet(newFlowCommand("idem-1", 10))
	if err != nil {
		t.Fatalf("first CreateOrGet error: %v", err)
	}
	if !created {
		t.Fatalf("first CreateOrGet created = false, want true")
	}
	if first.ID == 0 {
		t.Fatalf("first command ID = 0")
	}

	second, created, err := repo.CreateOrGet(newFlowCommand("idem-1", 10))
	if err != nil {
		t.Fatalf("duplicate CreateOrGet error: %v", err)
	}
	if created {
		t.Fatalf("duplicate CreateOrGet created = true, want false")
	}
	if second.ID != first.ID {
		t.Fatalf("duplicate command ID = %d, want %d", second.ID, first.ID)
	}
}

func TestFlowCommandRepoCreateOrGetReturnsNonIdempotencyInsertError(t *testing.T) {
	repo := setupFlowCommandRepo(t)
	first, _, err := repo.CreateOrGet(newFlowCommand("non-idem-error-1", 10))
	if err != nil {
		t.Fatalf("CreateOrGet first error: %v", err)
	}

	_, _, err = repo.CreateOrGet(&entity.FlowCommand{
		ID:              first.ID,
		CommandType:     "start_drill",
		DrillInstanceID: 1001,
		OperatorID:      10,
		IdempotencyKey:  "different-key",
		Payload:         `{"action":"start"}`,
		Status:          entity.FlowCommandPending,
	})
	if err == nil {
		t.Fatalf("CreateOrGet with duplicate primary key returned nil error")
	}
}

func TestFlowCommandRepoMarkSucceeded(t *testing.T) {
	repo := setupFlowCommandRepo(t)
	cmd, _, err := repo.CreateOrGet(newFlowCommand("success-1", 10))
	if err != nil {
		t.Fatalf("CreateOrGet error: %v", err)
	}

	if err := repo.MarkSucceeded(cmd.ID, map[string]any{"ok": true, "step": "done"}); err != nil {
		t.Fatalf("MarkSucceeded error: %v", err)
	}

	got, err := repo.FindByID(cmd.ID)
	if err != nil {
		t.Fatalf("FindByID error: %v", err)
	}
	if got.Status != entity.FlowCommandSucceeded {
		t.Fatalf("status = %s, want %s", got.Status, entity.FlowCommandSucceeded)
	}
	if got.Result == nil || !json.Valid([]byte(*got.Result)) {
		t.Fatalf("result = %v, want valid JSON", got.Result)
	}
	if got.FinishedAt == nil {
		t.Fatalf("finished_at is nil")
	}
}

func TestFlowCommandRepoMarkFailed(t *testing.T) {
	repo := setupFlowCommandRepo(t)
	cmd, _, err := repo.CreateOrGet(newFlowCommand("failed-1", 10))
	if err != nil {
		t.Fatalf("CreateOrGet error: %v", err)
	}

	if err := repo.MarkFailed(cmd.ID, "FLOW_ERROR", "flow failed"); err != nil {
		t.Fatalf("MarkFailed error: %v", err)
	}

	got, err := repo.FindByID(cmd.ID)
	if err != nil {
		t.Fatalf("FindByID error: %v", err)
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
	if got.FinishedAt == nil {
		t.Fatalf("finished_at is nil")
	}
}

func TestFlowCommandRepoFindByIDForOperator(t *testing.T) {
	repo := setupFlowCommandRepo(t)
	cmd, _, err := repo.CreateOrGet(newFlowCommand("operator-1", 10))
	if err != nil {
		t.Fatalf("CreateOrGet error: %v", err)
	}

	got, err := repo.FindByIDForOperator(cmd.ID, 10)
	if err != nil {
		t.Fatalf("FindByIDForOperator owner error: %v", err)
	}
	if got.ID != cmd.ID {
		t.Fatalf("owner got ID = %d, want %d", got.ID, cmd.ID)
	}

	if _, err := repo.FindByIDForOperator(cmd.ID, 11); err == nil {
		t.Fatalf("FindByIDForOperator with different operator returned nil error")
	}
}

func TestFlowCommandRepoExtendLeaseRequiresMatchingProcessingWorker(t *testing.T) {
	repo := setupFlowCommandRepo(t)
	cmd, _, err := repo.CreateOrGet(newFlowCommand("extend-1", 10))
	if err != nil {
		t.Fatalf("CreateOrGet error: %v", err)
	}
	worker := "worker-1"
	if err := repo.db.Model(&entity.FlowCommand{}).Where("id = ?", cmd.ID).Updates(map[string]any{
		"status":    entity.FlowCommandProcessing,
		"worker_id": worker,
	}).Error; err != nil {
		t.Fatalf("set processing: %v", err)
	}

	until := time.Now().UTC().Add(time.Minute)
	ok, err := repo.ExtendLease(context.Background(), cmd.ID, worker, until)
	if err != nil {
		t.Fatalf("ExtendLease owner error: %v", err)
	}
	if !ok {
		t.Fatalf("ExtendLease owner ok = false, want true")
	}

	ok, err = repo.ExtendLease(t.Context(), cmd.ID, "worker-2", until.Add(time.Minute))
	if err != nil {
		t.Fatalf("ExtendLease other worker error: %v", err)
	}
	if ok {
		t.Fatalf("ExtendLease other worker ok = true, want false")
	}

	if err := repo.MarkSucceeded(cmd.ID, map[string]any{"ok": true}); err != nil {
		t.Fatalf("MarkSucceeded error: %v", err)
	}
	ok, err = repo.ExtendLease(t.Context(), cmd.ID, worker, until.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("ExtendLease terminal error: %v", err)
	}
	if ok {
		t.Fatalf("ExtendLease terminal ok = true, want false")
	}
}

func TestFlowCommandRepoRequeueExpired(t *testing.T) {
	repo := setupFlowCommandRepo(t)
	now := time.Now().UTC()
	worker := "worker-1"
	expiredLease := now.Add(-time.Minute)
	activeLease := now.Add(time.Minute)

	expired, _, err := repo.CreateOrGet(newFlowCommand("expired-1", 10))
	if err != nil {
		t.Fatalf("CreateOrGet expired error: %v", err)
	}
	active, _, err := repo.CreateOrGet(newFlowCommand("active-1", 10))
	if err != nil {
		t.Fatalf("CreateOrGet active error: %v", err)
	}

	if err := repo.db.Model(&entity.FlowCommand{}).Where("id = ?", expired.ID).Updates(map[string]any{
		"status":      entity.FlowCommandProcessing,
		"worker_id":   worker,
		"lease_until": expiredLease,
	}).Error; err != nil {
		t.Fatalf("set expired processing: %v", err)
	}
	if err := repo.db.Model(&entity.FlowCommand{}).Where("id = ?", active.ID).Updates(map[string]any{
		"status":      entity.FlowCommandProcessing,
		"worker_id":   worker,
		"lease_until": activeLease,
	}).Error; err != nil {
		t.Fatalf("set active processing: %v", err)
	}

	count, err := repo.RequeueExpired(now)
	if err != nil {
		t.Fatalf("RequeueExpired error: %v", err)
	}
	if count != 1 {
		t.Fatalf("requeued count = %d, want 1", count)
	}

	gotExpired, err := repo.FindByID(expired.ID)
	if err != nil {
		t.Fatalf("FindByID expired error: %v", err)
	}
	if gotExpired.Status != entity.FlowCommandPending {
		t.Fatalf("expired status = %s, want %s", gotExpired.Status, entity.FlowCommandPending)
	}
	if gotExpired.WorkerID != nil {
		t.Fatalf("expired worker_id = %v, want nil", gotExpired.WorkerID)
	}
	if gotExpired.LeaseUntil != nil {
		t.Fatalf("expired lease_until = %v, want nil", gotExpired.LeaseUntil)
	}

	gotActive, err := repo.FindByID(active.ID)
	if err != nil {
		t.Fatalf("FindByID active error: %v", err)
	}
	if gotActive.Status != entity.FlowCommandProcessing {
		t.Fatalf("active status = %s, want %s", gotActive.Status, entity.FlowCommandProcessing)
	}
}
