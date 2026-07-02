package service

import (
	"testing"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTaskTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE user (
			id integer PRIMARY KEY,
			username text NOT NULL,
			real_name text NOT NULL,
			password_hash text,
			email text,
			role text NOT NULL,
			department text,
			phone text,
			status integer,
			created_at datetime
		);
		CREATE TABLE drill_instance (
			id integer PRIMARY KEY,
			template_id integer NOT NULL,
			name text NOT NULL,
			description text,
			status text NOT NULL,
			start_time datetime NULL,
			end_time datetime NULL,
			planned_start datetime NULL,
			current_task_id integer NULL,
			progress_pct integer DEFAULT 0,
			created_by integer NOT NULL,
			created_at datetime,
			updated_at datetime
		);
		CREATE TABLE drill_instance_step (
			id integer PRIMARY KEY,
			drill_instance_id integer NOT NULL,
			parent_step_id integer NULL,
			template_step_id integer NOT NULL,
			name text NOT NULL,
			seq integer NOT NULL,
			status text NOT NULL,
			assignee_ids text NOT NULL,
			actual_operator integer NULL,
			start_time datetime NULL,
			end_time datetime NULL,
			timeout_at datetime NULL,
			remark text,
			issue_desc text,
			step_type text,
			timeout_minutes integer,
			default_assignee_role text,
			executor_team text,
			phase text,
			phase_step text,
			pre_step_ids text,
			estimated_duration_minutes integer NULL,
			estimated_start_offset integer NULL,
			action_params text,
			created_at datetime
		);
		CREATE TABLE drill_instance_step_log (
			id integer PRIMARY KEY,
			drill_instance_id integer NOT NULL,
			task_instance_id integer NULL,
				action text NOT NULL,
				operator_id integer NULL,
				operator_name text,
				command_id integer NULL,
				content text,
				created_at datetime
			);
	`).Error; err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func withTaskTestDB(t *testing.T, fn func(*gorm.DB)) {
	t.Helper()
	db := setupTaskTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()
	fn(db)
}

func insertTaskTestUser(t *testing.T, db *gorm.DB, user entity.User) {
	t.Helper()
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func insertTaskTestDrill(t *testing.T, db *gorm.DB, drill entity.DrillInstance) {
	t.Helper()
	if err := db.Create(&drill).Error; err != nil {
		t.Fatalf("create drill: %v", err)
	}
}

func insertTaskTestStep(t *testing.T, db *gorm.DB, step entity.StepInstance) {
	t.Helper()
	if err := db.Create(&step).Error; err != nil {
		t.Fatalf("create step: %v", err)
	}
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func TestGetMyTasksReturnsOnlyUserAssignedTasks(t *testing.T) {
	withTaskTestDB(t, func(db *gorm.DB) {
		insertTaskTestUser(t, db, entity.User{ID: 7, Username: "executor", RealName: "执行员", Role: "executor", Department: "研发部"})
		insertTaskTestUser(t, db, entity.User{ID: 8, Username: "other-executor", RealName: "其他执行员", Role: "executor", Department: "研发部"})
		insertTaskTestUser(t, db, entity.User{ID: 9, Username: "admin", RealName: "管理员", Role: "admin", Department: "技术部"})
		insertTaskTestDrill(t, db, entity.DrillInstance{ID: 10, TemplateID: 1, Name: "活跃演练", Status: "running", CreatedBy: 1})

		insertTaskTestStep(t, db, entity.StepInstance{ID: 1, DrillInstanceID: 10, StepTemplateID: 101, Name: "精确分配", Seq: 1, Status: "running", AssigneeIDs: "[7]", DefaultAssigneeRole: "executor"})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 2, DrillInstanceID: 10, StepTemplateID: 102, Name: "部门分配", Seq: 2, Status: "pending", AssigneeIDs: "[]", DefaultAssigneeRole: "executor", ExecutorTeam: "研发部"})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 3, DrillInstanceID: 10, StepTemplateID: 103, Name: "仅角色匹配", Seq: 3, Status: "pending", AssigneeIDs: "[]", DefaultAssigneeRole: "executor"})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 4, DrillInstanceID: 10, StepTemplateID: 104, Name: "其他部门", Seq: 4, Status: "pending", AssigneeIDs: "[]", DefaultAssigneeRole: "executor", ExecutorTeam: "网络部"})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 5, DrillInstanceID: 10, StepTemplateID: 105, Name: "同部门其他人", Seq: 5, Status: "pending", AssigneeIDs: "[8]", DefaultAssigneeRole: "executor", ExecutorTeam: "研发部"})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 6, DrillInstanceID: 10, StepTemplateID: 106, Name: "显式分配给管理员", Seq: 6, Status: "running", AssigneeIDs: "[9]", DefaultAssigneeRole: "executor", ExecutorTeam: "技术部"})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 7, DrillInstanceID: 10, StepTemplateID: 107, Name: "操作人属性匹配", Seq: 7, Status: "completed", AssigneeIDs: "[9]", DefaultAssigneeRole: "executor", ExecutorTeam: "技术部", JSONAttributes: `{"operator":"执行员"}`})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 8, DrillInstanceID: 10, StepTemplateID: 108, Name: "操作人属性不匹配", Seq: 8, Status: "running", AssigneeIDs: "[]", DefaultAssigneeRole: "executor", JSONAttributes: `{"operator":"李四"}`})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 9, DrillInstanceID: 10, StepTemplateID: 109, Name: "实际操作人匹配", Seq: 9, Status: "completed", AssigneeIDs: "[9]", DefaultAssigneeRole: "executor", ExecutorTeam: "技术部", ActualOperator: uint64Ptr(7), JSONAttributes: `{"operator":"李四"}`})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 10, DrillInstanceID: 10, StepTemplateID: 110, Name: "混合属性匹配", Seq: 10, Status: "running", AssigneeIDs: "[]", DefaultAssigneeRole: "executor", JSONAttributes: `{"operator":"执行员","generated":true}`})

		svc := NewTaskService(repository.NewStepRepo())
		svc.SetUserRepo(repository.NewUserRepo())

		tasks, err := svc.GetMyTasks(7)
		if err != nil {
			t.Fatalf("GetMyTasks: %v", err)
		}
		if len(tasks) != 4 {
			t.Fatalf("expected 4 user-owned tasks, got %d: %#v", len(tasks), tasks)
		}
		got := map[uint64]bool{}
		for _, task := range tasks {
			got[task.ID] = true
		}
		if !got[1] || !got[7] || !got[9] || !got[10] {
			t.Fatalf("expected explicit, operator, mixed-attribute, and actual-operator tasks, got %#v", got)
		}
		if got[2] || got[3] || got[4] || got[5] || got[6] || got[8] {
			t.Fatalf("tasks not owned by current operator must not be returned: %#v", got)
		}
	})
}

func TestCompleteStepAllowsExplicitAssigneeTask(t *testing.T) {
	withTaskTestDB(t, func(db *gorm.DB) {
		insertTaskTestUser(t, db, entity.User{ID: 7, Username: "executor", RealName: "执行员", Role: "executor", Department: "研发部"})
		insertTaskTestDrill(t, db, entity.DrillInstance{ID: 10, TemplateID: 1, Name: "活跃演练", Status: "running", CreatedBy: 1})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 1, DrillInstanceID: 10, StepTemplateID: 101, Name: "精确分配", Seq: 1, Status: "running", AssigneeIDs: "[7]", DefaultAssigneeRole: "executor"})

		svc := NewTaskService(repository.NewStepRepo())
		svc.SetUserRepo(repository.NewUserRepo())

		if err := svc.CompleteStep(1, 7, "done"); err != nil {
			t.Fatalf("expected explicit-assignee task to be completed: %v", err)
		}
	})
}

func TestCompleteStepRejectsDepartmentOnlyTask(t *testing.T) {
	withTaskTestDB(t, func(db *gorm.DB) {
		insertTaskTestUser(t, db, entity.User{ID: 7, Username: "executor", RealName: "执行员", Role: "executor", Department: "研发部"})
		insertTaskTestDrill(t, db, entity.DrillInstance{ID: 10, TemplateID: 1, Name: "活跃演练", Status: "running", CreatedBy: 1})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 2, DrillInstanceID: 10, StepTemplateID: 102, Name: "部门分配", Seq: 2, Status: "running", AssigneeIDs: "[]", DefaultAssigneeRole: "executor", ExecutorTeam: "研发部"})

		svc := NewTaskService(repository.NewStepRepo())
		svc.SetUserRepo(repository.NewUserRepo())

		if err := svc.CompleteStep(2, 7, "done"); err == nil {
			t.Fatalf("expected department-only task to be rejected")
		}
	})
}

func TestCompleteStepRejectsRoleOnlyTask(t *testing.T) {
	withTaskTestDB(t, func(db *gorm.DB) {
		insertTaskTestUser(t, db, entity.User{ID: 7, Username: "executor", RealName: "执行员", Role: "executor", Department: "研发部"})
		insertTaskTestDrill(t, db, entity.DrillInstance{ID: 10, TemplateID: 1, Name: "活跃演练", Status: "running", CreatedBy: 1})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 3, DrillInstanceID: 10, StepTemplateID: 103, Name: "仅角色匹配", Seq: 3, Status: "running", AssigneeIDs: "[]", DefaultAssigneeRole: "executor"})

		svc := NewTaskService(repository.NewStepRepo())
		svc.SetUserRepo(repository.NewUserRepo())

		if err := svc.CompleteStep(3, 7, "done"); err == nil {
			t.Fatalf("expected role-only task to be rejected")
		}
	})
}

func TestCompleteStepRejectsTaskAssignedToAnotherUser(t *testing.T) {
	withTaskTestDB(t, func(db *gorm.DB) {
		insertTaskTestUser(t, db, entity.User{ID: 7, Username: "executor", RealName: "执行员", Role: "executor", Department: "研发部"})
		insertTaskTestUser(t, db, entity.User{ID: 8, Username: "other-executor", RealName: "其他执行员", Role: "executor", Department: "研发部"})
		insertTaskTestDrill(t, db, entity.DrillInstance{ID: 10, TemplateID: 1, Name: "活跃演练", Status: "running", CreatedBy: 1})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 5, DrillInstanceID: 10, StepTemplateID: 105, Name: "同部门其他人", Seq: 5, Status: "running", AssigneeIDs: "[8]", DefaultAssigneeRole: "executor", ExecutorTeam: "研发部"})

		svc := NewTaskService(repository.NewStepRepo())
		svc.SetUserRepo(repository.NewUserRepo())

		if err := svc.CompleteStep(5, 7, "done"); err == nil {
			t.Fatalf("expected task assigned to another user to be rejected")
		}
	})
}

func TestCompleteStepRejectsExplicitAssigneeEvenWhenRoleMatches(t *testing.T) {
	withTaskTestDB(t, func(db *gorm.DB) {
		insertTaskTestUser(t, db, entity.User{ID: 7, Username: "executor", RealName: "执行员", Role: "executor", Department: "研发部"})
		insertTaskTestUser(t, db, entity.User{ID: 9, Username: "admin", RealName: "管理员", Role: "admin", Department: "技术部"})
		insertTaskTestDrill(t, db, entity.DrillInstance{ID: 10, TemplateID: 1, Name: "活跃演练", Status: "running", CreatedBy: 1})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 6, DrillInstanceID: 10, StepTemplateID: 106, Name: "显式分配给管理员", Seq: 6, Status: "running", AssigneeIDs: "[9]", DefaultAssigneeRole: "executor", ExecutorTeam: "技术部"})

		svc := NewTaskService(repository.NewStepRepo())
		svc.SetUserRepo(repository.NewUserRepo())

		if err := svc.CompleteStep(6, 7, "done"); err == nil {
			t.Fatalf("expected explicit assignee mismatch to be rejected")
		}
	})
}

func TestCompleteStepRejectsTaskThatIsNotRunning(t *testing.T) {
	withTaskTestDB(t, func(db *gorm.DB) {
		insertTaskTestUser(t, db, entity.User{ID: 7, Username: "executor", RealName: "执行员", Role: "executor", Department: "研发部"})
		insertTaskTestDrill(t, db, entity.DrillInstance{ID: 10, TemplateID: 1, Name: "活跃演练", Status: "running", CreatedBy: 1})
		insertTaskTestStep(t, db, entity.StepInstance{ID: 1, DrillInstanceID: 10, StepTemplateID: 101, Name: "待执行任务", Seq: 1, Status: "pending", AssigneeIDs: "[7]", DefaultAssigneeRole: "executor"})

		svc := NewTaskService(repository.NewStepRepo())
		svc.SetUserRepo(repository.NewUserRepo())

		if err := svc.CompleteStep(1, 7, "done"); err == nil {
			t.Fatalf("expected pending task completion to be rejected")
		}
	})
}
