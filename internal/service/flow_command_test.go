package service

import (
	"context"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFlowCommandServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&entity.FlowCommand{}); err != nil {
		t.Fatalf("migrate flow command: %v", err)
	}
	return db
}

func newFlowCommandServiceForTest(t *testing.T, waitTimeout time.Duration) (*FlowCommandService, *repository.FlowCommandRepo) {
	t.Helper()
	db := setupFlowCommandServiceTestDB(t)
	repo := repository.NewFlowCommandRepo(db)
	svc := NewFlowCommandService(repo, waitTimeout)
	svc.pollInterval = time.Millisecond
	return svc, repo
}

func TestFlowCommandSubmitPreservesSuppliedIdempotencyKey(t *testing.T) {
	svc, _ := newFlowCommandServiceForTest(t, time.Millisecond)

	result, err := svc.SubmitAndWait(context.Background(), SubmitCommandRequest{
		CommandType:     "start_drill",
		DrillInstanceID: 10,
		OperatorID:      7,
		IdempotencyKey:  "client-key-1",
		Payload:         map[string]any{"reason": "manual"},
	})
	if err != nil {
		t.Fatalf("SubmitAndWait: %v", err)
	}
	if result.Command.IdempotencyKey != "client-key-1" {
		t.Fatalf("expected supplied idempotency key to be preserved, got %q", result.Command.IdempotencyKey)
	}
	if result.Command.Payload != `{"reason":"manual"}` {
		t.Fatalf("expected payload to be stored as JSON, got %s", result.Command.Payload)
	}
}

func TestFlowCommandSubmitGeneratesIdempotencyKeyWhenMissing(t *testing.T) {
	svc, _ := newFlowCommandServiceForTest(t, time.Millisecond)

	result, err := svc.SubmitAndWait(context.Background(), SubmitCommandRequest{
		CommandType:     "start_drill",
		DrillInstanceID: 10,
		OperatorID:      7,
		Payload:         map[string]any{"reason": "manual"},
	})
	if err != nil {
		t.Fatalf("SubmitAndWait: %v", err)
	}
	if result.Command.IdempotencyKey == "" {
		t.Fatalf("expected generated idempotency key")
	}
}

func TestFlowCommandSubmitDuplicateSubmissionReturnsExistingCommand(t *testing.T) {
	svc, _ := newFlowCommandServiceForTest(t, time.Millisecond)
	req := SubmitCommandRequest{
		CommandType:     "start_drill",
		DrillInstanceID: 10,
		OperatorID:      7,
		IdempotencyKey:  "duplicate-key",
		Payload:         map[string]any{"attempt": 1},
	}

	first, err := svc.SubmitAndWait(context.Background(), req)
	if err != nil {
		t.Fatalf("first SubmitAndWait: %v", err)
	}
	second, err := svc.SubmitAndWait(context.Background(), req)
	if err != nil {
		t.Fatalf("second SubmitAndWait: %v", err)
	}
	if first.Command.ID == 0 || second.Command.ID != first.Command.ID {
		t.Fatalf("expected duplicate submission to return existing command id %d, got %d", first.Command.ID, second.Command.ID)
	}
}

func TestFlowCommandSubmitTerminalCommandsReturnImmediately(t *testing.T) {
	svc, repo := newFlowCommandServiceForTest(t, 100*time.Millisecond)
	cmd, _, err := repo.CreateOrGet(&entity.FlowCommand{
		CommandType:     "start_drill",
		DrillInstanceID: 10,
		OperatorID:      7,
		IdempotencyKey:  "terminal-key",
		Payload:         `{}`,
		Status:          entity.FlowCommandSucceeded,
	})
	if err != nil {
		t.Fatalf("seed command: %v", err)
	}

	started := time.Now()
	result, err := svc.SubmitAndWait(context.Background(), SubmitCommandRequest{
		CommandType:     "start_drill",
		DrillInstanceID: 10,
		OperatorID:      7,
		IdempotencyKey:  "terminal-key",
		Payload:         map[string]any{},
	})
	if err != nil {
		t.Fatalf("SubmitAndWait: %v", err)
	}
	if time.Since(started) >= 50*time.Millisecond {
		t.Fatalf("expected terminal command to return immediately")
	}
	if result.Pending {
		t.Fatalf("expected terminal command pending=false")
	}
	if result.Command.ID != cmd.ID || result.Command.Status != entity.FlowCommandSucceeded {
		t.Fatalf("expected existing terminal command, got %#v", result.Command)
	}
}

func TestFlowCommandSubmitPendingCommandsReturnAfterConfiguredWaitTimeout(t *testing.T) {
	svc, _ := newFlowCommandServiceForTest(t, 10*time.Millisecond)

	started := time.Now()
	result, err := svc.SubmitAndWait(context.Background(), SubmitCommandRequest{
		CommandType:     "start_drill",
		DrillInstanceID: 10,
		OperatorID:      7,
		IdempotencyKey:  "pending-key",
		Payload:         map[string]any{},
	})
	if err != nil {
		t.Fatalf("SubmitAndWait: %v", err)
	}
	if !result.Pending {
		t.Fatalf("expected pending command to return pending=true")
	}
	if result.Command.Status != entity.FlowCommandPending {
		t.Fatalf("expected pending status, got %s", result.Command.Status)
	}
	if time.Since(started) < 10*time.Millisecond {
		t.Fatalf("expected SubmitAndWait to wait for configured timeout")
	}
}
