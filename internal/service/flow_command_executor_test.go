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

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// fakePublisher records events for assertion.
type fakePublisher struct {
	mu      sync.Mutex
	events  []events.Event
	onEvent func(events.Event)
}

func (p *fakePublisher) Publish(_ context.Context, event events.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.onEvent != nil {
		p.onEvent(event)
	}
	p.events = append(p.events, event)
	return nil
}

func (p *fakePublisher) Events() []events.Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	cp := make([]events.Event, len(p.events))
	copy(cp, p.events)
	return cp
}

// fakeLeaderGuard always reports as leader.
type fakeLeaderGuard struct {
	value string
}

func (g *fakeLeaderGuard) Acquire(_ context.Context) (bool, error) { return true, nil }
func (g *fakeLeaderGuard) Renew(_ context.Context) (bool, error)   { return true, nil }
func (g *fakeLeaderGuard) Release(_ context.Context) (bool, error) { return true, nil }
func (g *fakeLeaderGuard) Value() string                           { return g.value }

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
			lease_until DATETIME NULL,
			attempts INTEGER DEFAULT 0,
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
	leader := &fakeLeaderGuard{value: "test-worker"}
	executor := NewFlowCommandExecutor(db, repo, nil, nil, publisher, leader)
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

func (r *flowCommandRepoForTest) MarkSucceeded(id uint64, result any) error {
	return r.db.Model(&entity.FlowCommand{}).Where("id = ?", id).Updates(map[string]any{
		"status":      entity.FlowCommandSucceeded,
		"finished_at": time.Now(),
	}).Error
}

func (r *flowCommandRepoForTest) MarkFailed(id uint64, code, message string) error {
	return r.db.Model(&entity.FlowCommand{}).Where("id = ?", id).Updates(map[string]any{
		"status":        entity.FlowCommandFailed,
		"error_code":    code,
		"error_message": message,
		"finished_at":   time.Now(),
	}).Error
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
	cmd := &entity.FlowCommand{
		CommandType:     cmdType,
		DrillInstanceID: drillID,
		StepInstanceID:  stepID,
		OperatorID:      100,
		IdempotencyKey:  fmt.Sprintf("key-%s-%d", cmdType, time.Now().UnixNano()),
		Payload:         string(payloadBytes),
		Status:          entity.FlowCommandProcessing,
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
	if err := executor.Execute(context.Background(), cmd); err != nil {
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

	if err := executor.Execute(context.Background(), cmd); err != nil {
		t.Fatalf("first Execute: %v", err)
	}

	var cmd2 entity.FlowCommand
	if err := db.First(&cmd2, cmd.ID).Error; err != nil {
		t.Fatalf("load command: %v", err)
	}
	if err := executor.Execute(context.Background(), &cmd2); err != nil {
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
	err := executor.Execute(context.Background(), cmd)
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
			_ = executor.Execute(context.Background(), cmd)

			tt.check(t, db, stepID)
		})
	}
}

func TestExecutePublishesEventsAfterCommit(t *testing.T) {
	executor, db, publisher := newExecutorForTest(t)
	drillID, stepID := seedDrillAndStepForExecutorTest(t, db, "running")

	var publishedBeforeCommit bool
	publisher.onEvent = func(_ events.Event) {
		var step entity.StepInstance
		if err := db.First(&step, stepID).Error; err == nil {
			if step.Status != "completed" {
				publishedBeforeCommit = true
			}
		}
	}

	cmd := createExecutorCommand(t, db, "complete_step", drillID, &stepID, CompleteStepPayload{Remark: "done"})
	if err := executor.Execute(context.Background(), cmd); err != nil {
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
