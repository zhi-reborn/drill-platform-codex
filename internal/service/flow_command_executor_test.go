package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/events"
	"drill-platform/internal/repository"
	"drill-platform/internal/worker"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// fakePublisher records events for assertion.
type fakePublisher struct {
	mu      sync.Mutex
	events  []events.WSMessage
	onEvent func(events.WSMessage)
}

func (p *fakePublisher) Publish(_ context.Context, event events.WSMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.onEvent != nil {
		p.onEvent(event)
	}
	p.events = append(p.events, event)
	return nil
}

func (p *fakePublisher) Events() []events.WSMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	cp := make([]events.WSMessage, len(p.events))
	copy(cp, p.events)
	return cp
}

// fakeLeaderGuard is retained for backward compatibility with any test that
// still references it; the executor no longer takes a LeaderGuard.
type fakeLeaderGuard struct {
	value string
}

func (g *fakeLeaderGuard) Acquire(_ context.Context) (bool, error) { return true, nil }
func (g *fakeLeaderGuard) Renew(_ context.Context) (bool, error)   { return true, nil }
func (g *fakeLeaderGuard) Release(_ context.Context) (bool, error) { return true, nil }
func (g *fakeLeaderGuard) Value() string                           { return g.value }

// testFence returns the ExecutionFence matching the ownership fields stamped
// by createExecutorCommand. Tests that need a stale fence should override the
// command's ownership columns after creation.
func testFence() worker.ExecutionFence {
	return worker.ExecutionFence{
		WorkerID:    "test-worker",
		WorkerEpoch: 1,
		LeaseToken:  "test-lease-token",
	}
}

func setupExecutorTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec(`
		CREATE TABLE drill_flow_command (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command_type TEXT NOT NULL,
			drill_instance_id INTEGER NOT NULL,
			step_instance_id INTEGER NULL,
			operator_id INTEGER NOT NULL,
			idempotency_key TEXT NOT NULL UNIQUE,
			payload TEXT NOT NULL,
			status TEXT NOT NULL,
			worker_id TEXT,
			worker_epoch INTEGER DEFAULT 0,
			lease_token TEXT DEFAULT '',
			lease_until DATETIME NULL,
			attempts INTEGER DEFAULT 0,
			attempt_count INTEGER DEFAULT 0,
			result TEXT,
			error_code TEXT,
			error_message TEXT,
			created_at DATETIME,
			started_at DATETIME NULL,
			finished_at DATETIME NULL,
			updated_at DATETIME
		);
		CREATE TABLE drill_instance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			template_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL,
			start_time DATETIME NULL,
			end_time DATETIME NULL,
			planned_start DATETIME NULL,
			current_task_id INTEGER NULL,
			progress_pct INTEGER DEFAULT 0,
			created_by INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		);
		CREATE TABLE drill_instance_step (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			drill_instance_id INTEGER NOT NULL,
			parent_step_id INTEGER NULL,
			template_step_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			seq INTEGER NOT NULL,
			status TEXT NOT NULL,
			assignee_ids TEXT NOT NULL,
			actual_operator INTEGER NULL,
			start_time DATETIME NULL,
			end_time DATETIME NULL,
			timeout_at DATETIME NULL,
			remark TEXT,
			issue_desc TEXT,
			step_type TEXT,
			timeout_minutes INTEGER,
			default_assignee_role TEXT,
			executor_team TEXT,
			phase TEXT,
			phase_step TEXT,
			pre_step_ids TEXT,
			estimated_duration_minutes INTEGER NULL,
			estimated_start_offset INTEGER NULL,
			action_params TEXT,
			created_at DATETIME
		);
		CREATE TABLE drill_instance_step_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			drill_instance_id INTEGER NOT NULL,
			task_instance_id INTEGER NULL,
			command_id INTEGER NULL,
			action TEXT NOT NULL,
			operator_id INTEGER,
			operator_name TEXT,
			content TEXT,
			created_at DATETIME
		);
		CREATE TABLE notification (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			type TEXT NOT NULL,
			command_id INTEGER NULL,
			title TEXT NOT NULL,
			content TEXT,
			drill_id INTEGER NULL,
			drill_name TEXT,
			step_id INTEGER NULL,
			step_name TEXT,
			is_read INTEGER DEFAULT 0,
			created_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newExecutorForTest(t *testing.T) (*FlowCommandExecutor, *gorm.DB, *fakePublisher) {
	t.Helper()
	db := setupExecutorTestDB(t)
	repo := newFlowCommandRepoForExecutorTest(db)
	publisher := &fakePublisher{}
	executor := NewFlowCommandExecutor(db, repo, publisher)
	return executor, db, publisher
}

func newFlowCommandRepoForExecutorTest(db *gorm.DB) *flowCommandRepoForTest {
	return &flowCommandRepoForTest{db: db}
}

type flowCommandRepoForTest struct {
	db *gorm.DB
}

func (r *flowCommandRepoForTest) CreateOrGet(cmd *entity.FlowCommand) (*entity.FlowCommand, bool, error) {
	if err := r.db.Create(cmd).Error; err != nil {
		var existing entity.FlowCommand
		if findErr := r.db.Where("idempotency_key = ?", cmd.IdempotencyKey).First(&existing).Error; findErr != nil {
			return nil, false, findErr
		}
		return &existing, false, nil
	}
	return cmd, true, nil
}

func (r *flowCommandRepoForTest) FindByID(id uint64) (*entity.FlowCommand, error) {
	var cmd entity.FlowCommand
	if err := r.db.First(&cmd, id).Error; err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (r *flowCommandRepoForTest) MarkSucceededFenced(ctx context.Context, id uint64, ownership repository.CommandOwnership, result any) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&entity.FlowCommand{}).
		Where("id = ? AND status = ? AND worker_id = ? AND worker_epoch = ? AND lease_token = ? AND lease_until > ?",
			id, entity.FlowCommandProcessing, ownership.WorkerID, ownership.Epoch, ownership.Token, now).
		Updates(map[string]any{
			"status":      entity.FlowCommandSucceeded,
			"finished_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrOwnershipLost
	}
	return nil
}

func (r *flowCommandRepoForTest) MarkFailedFenced(ctx context.Context, id uint64, ownership repository.CommandOwnership, code, message string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&entity.FlowCommand{}).
		Where("id = ? AND status = ? AND worker_id = ? AND worker_epoch = ? AND lease_token = ? AND lease_until > ?",
			id, entity.FlowCommandProcessing, ownership.WorkerID, ownership.Epoch, ownership.Token, now).
		Updates(map[string]any{
			"status":        entity.FlowCommandFailed,
			"error_code":    code,
			"error_message": message,
			"finished_at":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrOwnershipLost
	}
	return nil
}

func seedDrillAndStepForExecutorTest(t *testing.T, db *gorm.DB, stepStatus string) (uint64, uint64) {
	t.Helper()
	drill := entity.DrillInstance{
		ID: 1, TemplateID: 1, Name: "test-drill", Status: "running", CreatedBy: 100,
	}
	if err := db.Create(&drill).Error; err != nil {
		t.Fatalf("create drill: %v", err)
	}
	step := entity.StepInstance{
		ID: 10, DrillInstanceID: 1, StepTemplateID: 101, Name: "test-step",
		Seq: 1, Status: stepStatus, AssigneeIDs: "[]",
	}
	if err := db.Create(&step).Error; err != nil {
		t.Fatalf("create step: %v", err)
	}
	return drill.ID, step.ID
}

func createExecutorCommand(t *testing.T, db *gorm.DB, cmdType string, drillID uint64, stepID *uint64, payload any) *entity.FlowCommand {
	t.Helper()
	payloadBytes, _ := json.Marshal(payload)
	workerID := "test-worker"
	leaseUntil := time.Now().Add(time.Hour)
	cmd := &entity.FlowCommand{
		CommandType:     cmdType,
		DrillInstanceID: drillID,
		StepInstanceID:  stepID,
		OperatorID:      100,
		IdempotencyKey:  fmt.Sprintf("key-%s-%d", cmdType, time.Now().UnixNano()),
		Payload:         string(payloadBytes),
		Status:          entity.FlowCommandProcessing,
		WorkerID:        &workerID,
		WorkerEpoch:     1,
		LeaseToken:      "test-lease-token",
		LeaseUntil:      &leaseUntil,
	}
	if err := db.Create(cmd).Error; err != nil {
		t.Fatalf("create command: %v", err)
	}
	return cmd
}

func TestExecuteCompleteStepOnRunningStepChangesItOnce(t *testing.T) {
	executor, db, _ := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")

	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})
	if err := executor.Execute(context.Background(), cmd, testFence()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var step entity.StepInstance
	if err := db.First(&step, stepID).Error; err != nil {
		t.Fatalf("load step: %v", err)
	}
	if step.Status != "completed" {
		t.Fatalf("expected step status completed, got %s", step.Status)
	}
	if step.Remark != "done" {
		t.Fatalf("expected step remark 'done', got %q", step.Remark)
	}

	var logCount int64
	db.Model(&entity.DrillInstanceLog{}).Where("task_instance_id = ?", stepID).Count(&logCount)
	if logCount != 1 {
		t.Fatalf("expected 1 log, got %d", logCount)
	}

	var notifCount int64
	db.Model(&entity.Notification{}).Where("step_id = ?", stepID).Count(&notifCount)
	if notifCount != 1 {
		t.Fatalf("expected 1 notification, got %d", notifCount)
	}

	var updated entity.FlowCommand
	db.First(&updated, cmd.ID)
	if updated.Status != entity.FlowCommandSucceeded {
		t.Fatalf("expected command succeeded, got %s", updated.Status)
	}
}

func TestExecuteReplaySameCommandReturnsSuccessWithoutDuplicateLogOrNotification(t *testing.T) {
	executor, db, _ := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")

	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})

	if err := executor.Execute(context.Background(), cmd, testFence()); err != nil {
		t.Fatalf("first Execute: %v", err)
	}

	var cmd2 entity.FlowCommand
	if err := db.First(&cmd2, cmd.ID).Error; err != nil {
		t.Fatalf("load command: %v", err)
	}
	if err := executor.Execute(context.Background(), &cmd2, testFence()); err != nil {
		t.Fatalf("replay Execute: %v", err)
	}

	var logCount int64
	db.Model(&entity.DrillInstanceLog{}).Where("task_instance_id = ?", stepID).Count(&logCount)
	if logCount != 1 {
		t.Fatalf("expected 1 log after replay, got %d", logCount)
	}

	var notifCount int64
	db.Model(&entity.Notification{}).Where("step_id = ?", stepID).Count(&notifCount)
	if notifCount != 1 {
		t.Fatalf("expected 1 notification after replay, got %d", notifCount)
	}
}

func TestExecuteCompleteStepOnPendingStepFailsWithInvalidStatus(t *testing.T) {
	executor, db, _ := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "pending")

	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})
	err := executor.Execute(context.Background(), cmd, testFence())
	if err == nil {
		t.Fatal("expected error for completing pending step")
	}

	var updated entity.FlowCommand
	db.First(&updated, cmd.ID)
	if updated.Status != entity.FlowCommandFailed {
		t.Fatalf("expected command failed, got %s", updated.Status)
	}
	if updated.ErrorCode == nil || *updated.ErrorCode != "invalid_status" {
		t.Fatalf("expected error_code invalid_status, got %v", updated.ErrorCode)
	}

	var step entity.StepInstance
	db.First(&step, stepID)
	if step.Status != "pending" {
		t.Fatalf("expected step still pending, got %s", step.Status)
	}
}

func TestExecuteDecodesTypedPayloads(t *testing.T) {
	tests := []struct {
		name       string
		cmdType    string
		payload    any
		setupState string
		check      func(t *testing.T, db *gorm.DB, stepID uint64)
	}{
		{
			name:       "complete_step decodes remark",
			cmdType:    "complete_step",
			payload:    CompleteStepPayload{Remark: "finished work"},
			setupState: "running",
			check: func(t *testing.T, db *gorm.DB, stepID uint64) {
				var step entity.StepInstance
				db.First(&step, stepID)
				if step.Remark != "finished work" {
					t.Fatalf("expected remark 'finished work', got %q", step.Remark)
				}
				if step.Status != "completed" {
					t.Fatalf("expected status completed, got %s", step.Status)
				}
			},
		},
		{
			name:       "report_issue decodes issue_desc",
			cmdType:    "report_issue",
			payload:    ReportIssuePayload{IssueDesc: "network down"},
			setupState: "running",
			check: func(t *testing.T, db *gorm.DB, stepID uint64) {
				var step entity.StepInstance
				db.First(&step, stepID)
				if step.IssueDesc != "network down" {
					t.Fatalf("expected issue_desc 'network down', got %q", step.IssueDesc)
				}
				if step.Status != "issue" {
					t.Fatalf("expected status issue, got %s", step.Status)
				}
			},
		},
		{
			name:       "assign_step decodes assignee_ids",
			cmdType:    "assign_step",
			payload:    AssignStepPayload{AssigneeIDs: []uint64{7, 8}},
			setupState: "pending",
			check: func(t *testing.T, db *gorm.DB, stepID uint64) {
				var step entity.StepInstance
				db.First(&step, stepID)
				if step.AssigneeIDs != "[7,8]" {
					t.Fatalf("expected assignee_ids '[7,8]', got %q", step.AssigneeIDs)
				}
			},
		},
		{
			name:       "update_step_info decodes remark",
			cmdType:    "update_step_info",
			payload:    UpdateStepInfoPayload{Remark: "updated info"},
			setupState: "pending",
			check: func(t *testing.T, db *gorm.DB, stepID uint64) {
				var step entity.StepInstance
				db.First(&step, stepID)
				if step.Remark != "updated info" {
					t.Fatalf("expected remark 'updated info', got %q", step.Remark)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, db, _ := newExecutorForTest(t)
			drillID, stepID := seedDrillAndStepForExecutorTest(t, db, tt.setupState)

			cmd := createExecutorCommand(t, db, tt.cmdType, drillID, &stepID, tt.payload)
			_ = executor.Execute(context.Background(), cmd, testFence())

			tt.check(t, db, stepID)
		})
	}
}

func TestExecutePublishesEventsAfterCommit(t *testing.T) {
	executor, db, publisher := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")

	var publishedBeforeCommit bool
	publisher.onEvent = func(_ events.WSMessage) {
		var step entity.StepInstance
		if err := db.First(&step, stepID).Error; err == nil {
			if step.Status != "completed" {
				publishedBeforeCommit = true
			}
		}
	}

	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})
	if err := executor.Execute(context.Background(), cmd, testFence()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if publishedBeforeCommit {
		t.Fatal("event was published before the transaction committed")
	}

	published := publisher.Events()
	if len(published) == 0 {
		t.Fatal("expected at least one event to be published after commit")
	}

	var step entity.StepInstance
	db.First(&step, stepID)
	if step.Status != "completed" {
		t.Fatalf("expected step committed, got status %s", step.Status)
	}

	var updated entity.FlowCommand
	db.First(&updated, cmd.ID)
	if updated.Status != entity.FlowCommandSucceeded {
		t.Fatalf("expected command committed before events, got status %s", updated.Status)
	}
}

// TestExecuteRejectsStaleFence verifies that a stale ExecutionFence (one whose
// epoch/token do not match the command's stamped ownership) cannot mutate the
// command or the step. The fenced WHERE clause in markCommandSucceededInTx and
// the repo's MarkSucceededFenced/MarkFailedFenced must match zero rows, causing
// the transaction to roll back and the command to stay non-terminal.
func TestExecuteRejectsStaleFence(t *testing.T) {
	executor, db, _ := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")

	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})

	// Override the command's ownership to a different epoch/token than the
	// test fence carries. This simulates a stale worker whose epoch has been
	// superseded by a newer worker.
	future := time.Now().Add(time.Hour)
	if err := db.Model(&entity.FlowCommand{}).Where("id = ?", cmd.ID).Updates(map[string]any{
		"worker_epoch": 99,
		"lease_token":  "real-token",
		"lease_until":  future,
	}).Error; err != nil {
		t.Fatalf("stamp ownership: %v", err)
	}

	// testFence() carries epoch=1, token="test-lease-token" — both stale.
	_ = executor.Execute(context.Background(), cmd, testFence())

	var updated entity.FlowCommand
	db.First(&updated, cmd.ID)
	if updated.Status == entity.FlowCommandSucceeded {
		t.Fatalf("stale fence must not mark command succeeded; status = %s", updated.Status)
	}
	if updated.Status == entity.FlowCommandFailed {
		t.Fatalf("stale fence must not mark command failed; status = %s", updated.Status)
	}

	var step entity.StepInstance
	db.First(&step, stepID)
	if step.Status != "running" {
		t.Fatalf("stale fence must not change step status; got %s", step.Status)
	}
}

// TestExecuteStartDrillUsesTransactionPath verifies that executeStartDrill
// routes through the DB transaction path. There is no production bypass
// branch: the executor unconditionally updates drill status inside a fenced
// transaction. This test serves as a regression guard for the unified path.
func TestExecuteStartDrillUsesTransactionPath(t *testing.T) {
	db := setupExecutorTestDB(t)
	repo := newFlowCommandRepoForExecutorTest(db)
	publisher := &fakePublisher{}
	executor := NewFlowCommandExecutor(db, repo, publisher)

	drill := entity.DrillInstance{
		ID: 1, TemplateID: 1, Name: "test-drill", Status: "pending", CreatedBy: 100,
	}
	if err := db.Create(&drill).Error; err != nil {
		t.Fatalf("create drill: %v", err)
	}

	cmd := createExecutorCommand(t, db, "start_drill", drill.ID, nil, nil)
	if err := executor.Execute(context.Background(), cmd, testFence()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var updated entity.DrillInstance
	if err := db.First(&updated, drill.ID).Error; err != nil {
		t.Fatalf("load drill: %v", err)
	}
	if updated.Status != "running" {
		t.Fatalf("expected drill status 'running' (DB path), got %s", updated.Status)
	}

	var cmdUpdated entity.FlowCommand
	db.First(&cmdUpdated, cmd.ID)
	if cmdUpdated.Status != entity.FlowCommandSucceeded {
		t.Fatalf("expected command succeeded (DB path), got %s", cmdUpdated.Status)
	}
}

// TestExecuteCompleteStepUsesTransactionPath verifies that executeCompleteStep
// routes through the DB transaction path. There is no production bypass
// branch: the executor unconditionally updates step status inside a fenced
// transaction. This test serves as a regression guard for the unified path.
func TestExecuteCompleteStepUsesTransactionPath(t *testing.T) {
	db := setupExecutorTestDB(t)
	repo := newFlowCommandRepoForExecutorTest(db)
	publisher := &fakePublisher{}
	executor := NewFlowCommandExecutor(db, repo, publisher)

	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")
	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})

	if err := executor.Execute(context.Background(), cmd, testFence()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var step entity.StepInstance
	if err := db.First(&step, stepID).Error; err != nil {
		t.Fatalf("load step: %v", err)
	}
	if step.Status != "completed" {
		t.Fatalf("expected step status 'completed' (DB path), got %s", step.Status)
	}

	var updated entity.FlowCommand
	db.First(&updated, cmd.ID)
	if updated.Status != entity.FlowCommandSucceeded {
		t.Fatalf("expected command succeeded (DB path), got %s", updated.Status)
	}
}

// TestNamedLockAcquiredOnMutation verifies that acquireNamedLock is invoked
// exactly once with the drill ID and a 5-second timeout when a mutation
// transaction runs. The stub records calls so we can assert the lock was
// acquired before any business logic ran.
func TestNamedLockAcquiredOnMutation(t *testing.T) {
	type call struct {
		drillID  uint64
		timeout  int
	}
	var calls []call
	original := acquireNamedLock
	acquireNamedLock = func(tx *gorm.DB, drillID uint64, timeoutSeconds int) error {
		calls = append(calls, call{drillID: drillID, timeout: timeoutSeconds})
		return nil
	}
	defer func() { acquireNamedLock = original }()

	executor, db, _ := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")
	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})

	if err := executor.Execute(context.Background(), cmd, testFence()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected named lock acquired once, got %d calls: %v", len(calls), calls)
	}
	if calls[0].drillID != drillID {
		t.Fatalf("expected lock for drill %d, got %d", drillID, calls[0].drillID)
	}
	if calls[0].timeout != 5 {
		t.Fatalf("expected 5s timeout, got %d", calls[0].timeout)
	}
}

// TestNamedLockTimeoutIsRetryable verifies that when acquireNamedLock returns
// a lock_timeout error, the command is NOT marked failed (so it can be
// requeued when the lease expires) and no partial state is persisted.
func TestNamedLockTimeoutIsRetryable(t *testing.T) {
	original := acquireNamedLock
	acquireNamedLock = func(tx *gorm.DB, drillID uint64, timeoutSeconds int) error {
		return &commandError{Code: "lock_timeout", Message: "could not acquire lock"}
	}
	defer func() { acquireNamedLock = original }()

	executor, db, _ := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")
	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})

	err := executor.Execute(context.Background(), cmd, testFence())
	if err == nil {
		t.Fatal("expected lock timeout error")
	}

	// Command must not be marked failed (retryable) nor succeeded.
	var updated entity.FlowCommand
	db.First(&updated, cmd.ID)
	if updated.Status == entity.FlowCommandFailed {
		t.Fatalf("lock timeout must be retryable, not failed; status = %s", updated.Status)
	}
	if updated.Status == entity.FlowCommandSucceeded {
		t.Fatalf("lock timeout must not mark command succeeded; status = %s", updated.Status)
	}

	// No partial state: step must remain running.
	var step entity.StepInstance
	db.First(&step, stepID)
	if step.Status != "running" {
		t.Fatalf("expected step still running after lock timeout, got %s", step.Status)
	}
}

// TestRollbackLeavesNoPartialState verifies that when a mid-transaction failure
// occurs (e.g. notification insert fails because the table is missing), the
// entire transaction rolls back: step status, log entries, and command status
// all revert to their pre-transaction state.
func TestRollbackLeavesNoPartialState(t *testing.T) {
	executor, db, _ := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")

	// Drop the notification table to force a failure after the step status
	// update and log insert succeed but before the transaction commits.
	if err := db.Exec("DROP TABLE notification").Error; err != nil {
		t.Fatalf("drop notification table: %v", err)
	}

	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})
	_ = executor.Execute(context.Background(), cmd, testFence())

	var step entity.StepInstance
	if err := db.First(&step, stepID).Error; err != nil {
		t.Fatalf("load step: %v", err)
	}
	if step.Status != "running" {
		t.Fatalf("expected step status 'running' after rollback, got %s", step.Status)
	}

	var logCount int64
	db.Model(&entity.DrillInstanceLog{}).Where("task_instance_id = ?", stepID).Count(&logCount)
	if logCount != 0 {
		t.Fatalf("expected 0 logs after rollback, got %d", logCount)
	}

	var updated entity.FlowCommand
	db.First(&updated, cmd.ID)
	if updated.Status == entity.FlowCommandSucceeded {
		t.Fatalf("expected command NOT marked succeeded after rollback, got %s", updated.Status)
	}
}
