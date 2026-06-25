package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupFlowRecoveryTestDB creates an in-memory SQLite database with the
// tables required by FlowRecovery tests.
func setupFlowRecoveryTestDB(t *testing.T) *gorm.DB {
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
		CREATE TABLE drill_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT '',
			description TEXT,
			status INTEGER NOT NULL DEFAULT 1,
			created_by INTEGER NOT NULL DEFAULT 0,
			phase_order TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
		CREATE TABLE drill_template_step (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			drill_template_id INTEGER NOT NULL,
			parent_step_id INTEGER,
			name TEXT NOT NULL,
			seq INTEGER NOT NULL,
			step_type TEXT NOT NULL,
			timeout_minutes INTEGER NOT NULL DEFAULT 120,
			pre_step_ids TEXT,
			guide_content TEXT,
			is_blocking INTEGER NOT NULL DEFAULT 1,
			default_assignee_role TEXT,
			executor_team TEXT,
			phase TEXT,
			phase_step TEXT,
			estimated_duration_minutes INTEGER,
			estimated_start_offset INTEGER,
			attributes TEXT,
			created_at DATETIME
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
			parent_step_id INTEGER,
			template_step_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			seq INTEGER NOT NULL,
			status TEXT NOT NULL,
			assignee_ids TEXT NOT NULL,
			actual_operator INTEGER,
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
			estimated_duration_minutes INTEGER,
			estimated_start_offset INTEGER,
			action_params TEXT,
			created_at DATETIME
		);
		CREATE TABLE drill_flow_command (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command_type TEXT NOT NULL,
			drill_instance_id INTEGER NOT NULL,
			step_instance_id INTEGER,
			operator_id INTEGER NOT NULL,
			idempotency_key TEXT NOT NULL UNIQUE,
			payload TEXT NOT NULL,
			status TEXT NOT NULL,
			worker_id TEXT,
			lease_until DATETIME,
			attempts INTEGER DEFAULT 0,
			result TEXT,
			error_code TEXT,
			error_message TEXT,
			created_at DATETIME,
			started_at DATETIME,
			finished_at DATETIME,
			updated_at DATETIME
		);
		CREATE TABLE drill_instance_step_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			drill_instance_id INTEGER NOT NULL,
			task_instance_id INTEGER,
			command_id INTEGER,
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
			command_id INTEGER,
			title TEXT NOT NULL,
			content TEXT,
			drill_id INTEGER,
			drill_name TEXT,
			step_id INTEGER,
			step_name TEXT,
			is_read INTEGER DEFAULT 0,
			created_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seedRecoveryTestData persists a running drill with:
//   - one running step with a future timeout_at (step 1, template 101)
//   - one running step with an expired timeout_at (step 2, template 102)
//   - one completed predecessor step (step 3, template 103)
//   - one pending successor step (step 4, template 104, pre_step_ids=[3])
//
// It returns the drill ID and the expired step's timeout_at for idempotency
// key assertions.
func seedRecoveryTestData(t *testing.T, db *gorm.DB) (drillID uint64, expiredTimeoutAt time.Time) {
	t.Helper()
	now := time.Now()
	futureTimeout := now.Add(1 * time.Hour)
	expiredTimeout := now.Add(-1 * time.Hour)

	template := entity.DrillTemplate{
		ID:       1,
		Name:     "recovery-test-template",
		Category: "test",
		Status:   1,
	}
	if err := db.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}

	templateSteps := []entity.StepTemplate{
		{ID: 101, DrillTemplateID: 1, Name: "future-timeout-step", Seq: 1, StepType: "serial", TimeoutMinutes: 120, PreStepIDs: "[]"},
		{ID: 102, DrillTemplateID: 1, Name: "expired-timeout-step", Seq: 2, StepType: "serial", TimeoutMinutes: 120, PreStepIDs: "[]"},
		{ID: 103, DrillTemplateID: 1, Name: "completed-predecessor", Seq: 3, StepType: "serial", TimeoutMinutes: 120, PreStepIDs: "[]"},
		{ID: 104, DrillTemplateID: 1, Name: "pending-successor", Seq: 4, StepType: "serial", TimeoutMinutes: 120, PreStepIDs: "[]"},
	}
	for i := range templateSteps {
		if err := db.Create(&templateSteps[i]).Error; err != nil {
			t.Fatalf("create template step %d: %v", templateSteps[i].ID, err)
		}
	}

	drill := entity.DrillInstance{
		ID:         1,
		TemplateID: 1,
		Name:       "recovery-test-drill",
		Status:     "running",
		CreatedBy:  100,
	}
	if err := db.Create(&drill).Error; err != nil {
		t.Fatalf("create drill: %v", err)
	}

	stepInstances := []entity.StepInstance{
		{ID: 1, DrillInstanceID: 1, StepTemplateID: 101, Name: "future-timeout-step", Seq: 1, Status: "running", AssigneeIDs: "[]", TimeoutAt: &futureTimeout, StepType: "serial"},
		{ID: 2, DrillInstanceID: 1, StepTemplateID: 102, Name: "expired-timeout-step", Seq: 2, Status: "running", AssigneeIDs: "[]", TimeoutAt: &expiredTimeout, StepType: "serial"},
		{ID: 3, DrillInstanceID: 1, StepTemplateID: 103, Name: "completed-predecessor", Seq: 3, Status: "completed", AssigneeIDs: "[]", StepType: "serial"},
		{ID: 4, DrillInstanceID: 1, StepTemplateID: 104, Name: "pending-successor", Seq: 4, Status: "pending", AssigneeIDs: "[]", PreStepIDs: "[3]", StepType: "serial"},
	}
	for i := range stepInstances {
		if err := db.Create(&stepInstances[i]).Error; err != nil {
			t.Fatalf("create step instance %d: %v", stepInstances[i].ID, err)
		}
	}

	return 1, expiredTimeout
}

// newFlowRecoveryForTest wires up a FlowRecovery backed by a SQLite test DB.
// It sets repository.DB so the existing repos (which use the package-level DB)
// read from the test database.
func newFlowRecoveryForTest(t *testing.T) (*FlowRecovery, *gorm.DB, *flowengine.Engine, uint64, time.Time) {
	t.Helper()
	db := setupFlowRecoveryTestDB(t)

	origDB := repository.DB
	repository.DB = db
	t.Cleanup(func() { repository.DB = origDB })

	drillID, expiredTimeoutAt := seedRecoveryTestData(t, db)

	engine := flowengine.NewEngine()
	templateRepo := repository.NewTemplateRepo()
	drillRepo := repository.NewDrillRepo()
	stepRepo := repository.NewStepRepo()
	userRepo := repository.NewUserRepo()
	notificationRepo := repository.NewNotificationRepo()
	flowCommandRepo := repository.NewFlowCommandRepo(db)

	adapter := NewDrillFlowAdapter(templateRepo, drillRepo, stepRepo, notificationRepo, userRepo, nil, nil)
	adapter.engine = engine

	engine.SetCallbacks(adapter)
	engine.SetStepLoader(adapter)

	drillService := NewDrillService(drillRepo, templateRepo, stepRepo, userRepo)
	drillService.SetEngine(engine, adapter)

	recovery := NewFlowRecovery(drillService, drillRepo, stepRepo, flowCommandRepo)
	return recovery, db, engine, drillID, expiredTimeoutAt
}

func TestFlowRecovery_RecoverAll_RebuildsEngineAndRegistersFutureTimeout(t *testing.T) {
	recovery, db, engine, drillID, _ := newFlowRecoveryForTest(t)

	if err := recovery.RecoverAll(context.Background()); err != nil {
		t.Fatalf("RecoverAll: %v", err)
	}

	// Engine should have the drill instance rebuilt.
	inst, ok := engine.GetInstance(int64(drillID))
	if !ok {
		t.Fatal("expected engine to have the drill instance after recovery")
	}
	if inst.Status != flowengine.FlowStatusRunning {
		t.Fatalf("expected drill status running, got %s", inst.Status)
	}

	// Future timeout (step 1, template 101) should be registered.
	scheduler := engine.TimeoutScheduler()
	if !scheduler.IsRegistered(int64(drillID), 101) {
		t.Fatal("expected future timeout for step template 101 to be registered")
	}

	// Expired timeout (step 2, template 102) should NOT be registered.
	if scheduler.IsRegistered(int64(drillID), 102) {
		t.Fatal("expected expired timeout for step template 102 to NOT be registered")
	}

	// No logs should be created by recovery itself.
	var logCount int64
	db.Model(&entity.DrillInstanceLog{}).Where("drill_instance_id = ?", drillID).Count(&logCount)
	if logCount != 0 {
		t.Fatalf("expected 0 logs after recovery, got %d", logCount)
	}

	// No notifications should be created by recovery itself.
	var notifCount int64
	db.Model(&entity.Notification{}).Where("drill_id = ?", drillID).Count(&notifCount)
	if notifCount != 0 {
		t.Fatalf("expected 0 notifications after recovery, got %d", notifCount)
	}
}

func TestFlowRecovery_RecoverAll_ExpiredTimeoutEnqueuesCommand(t *testing.T) {
	recovery, db, _, drillID, expiredTimeoutAt := newFlowRecoveryForTest(t)

	if err := recovery.RecoverAll(context.Background()); err != nil {
		t.Fatalf("RecoverAll: %v", err)
	}

	// The expired step (ID=2) should produce exactly one step_timeout command.
	var commands []entity.FlowCommand
	if err := db.Where("drill_instance_id = ? AND command_type = ?", drillID, "step_timeout").Find(&commands).Error; err != nil {
		t.Fatalf("query commands: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected 1 step_timeout command, got %d", len(commands))
	}

	cmd := commands[0]
	expectedKey := fmt.Sprintf("timeout:%d:%d:%d", drillID, 2, expiredTimeoutAt.Unix())
	if cmd.IdempotencyKey != expectedKey {
		t.Fatalf("expected idempotency key %q, got %q", expectedKey, cmd.IdempotencyKey)
	}
	if cmd.StepInstanceID == nil || *cmd.StepInstanceID != 2 {
		t.Fatalf("expected step_instance_id=2, got %v", cmd.StepInstanceID)
	}
	if cmd.OperatorID != 0 {
		t.Fatalf("expected operator_id=0, got %d", cmd.OperatorID)
	}
	if cmd.Payload != "{}" {
		t.Fatalf("expected payload '{}', got %q", cmd.Payload)
	}
	if cmd.Status != entity.FlowCommandPending {
		t.Fatalf("expected status pending, got %s", cmd.Status)
	}
}

func TestFlowRecovery_RecoverAll_IdempotentCommandSubmission(t *testing.T) {
	recovery, db, _, drillID, _ := newFlowRecoveryForTest(t)

	if err := recovery.RecoverAll(context.Background()); err != nil {
		t.Fatalf("first RecoverAll: %v", err)
	}
	if err := recovery.RecoverAll(context.Background()); err != nil {
		t.Fatalf("second RecoverAll: %v", err)
	}

	// Calling RecoverAll twice should not produce duplicate timeout commands.
	var count int64
	db.Model(&entity.FlowCommand{}).
		Where("drill_instance_id = ? AND command_type = ?", drillID, "step_timeout").
		Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 step_timeout command after double recovery, got %d", count)
	}
}

func TestFlowRecovery_Recover_DelegatesToRecoverAll(t *testing.T) {
	recovery, db, _, drillID, _ := newFlowRecoveryForTest(t)

	// Recover implements worker.Recoverer interface and should delegate to RecoverAll.
	if err := recovery.Recover(context.Background()); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	// Verify the side effect: a timeout command should exist for the expired step.
	var cmdCount int64
	db.Model(&entity.FlowCommand{}).
		Where("drill_instance_id = ? AND command_type = ?", drillID, "step_timeout").
		Count(&cmdCount)
	if cmdCount != 1 {
		t.Fatalf("expected 1 step_timeout command after Recover, got %d", cmdCount)
	}
}

func TestFlowRecovery_IdempotencyKeyFormat(t *testing.T) {
	recovery, db, _, drillID, expiredTimeoutAt := newFlowRecoveryForTest(t)

	if err := recovery.RecoverAll(context.Background()); err != nil {
		t.Fatalf("RecoverAll: %v", err)
	}

	expectedKey := fmt.Sprintf("timeout:%d:%d:%d", drillID, 2, expiredTimeoutAt.Unix())
	var cmd entity.FlowCommand
	if err := db.Where("idempotency_key = ?", expectedKey).First(&cmd).Error; err != nil {
		t.Fatalf("expected command with key %q: %v", expectedKey, err)
	}
}
