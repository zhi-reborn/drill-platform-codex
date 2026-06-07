package drill

import (
	"testing"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStepTargetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	schema := []string{
		`CREATE TABLE drill_instance (
			id INTEGER PRIMARY KEY,
			template_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL
		)`,
		`CREATE TABLE drill_instance_step (
			id INTEGER PRIMARY KEY,
			drill_instance_id INTEGER NOT NULL,
			step_template_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			seq INTEGER NOT NULL,
			status TEXT NOT NULL,
			assignee_ids TEXT NOT NULL
		)`,
		`CREATE TABLE drill_template_step (
			id INTEGER PRIMARY KEY,
			drill_template_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			seq INTEGER NOT NULL,
			step_type TEXT NOT NULL
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}
	return db
}

func TestResolveStepOperationTargetBackfillsMissingTemplateID(t *testing.T) {
	db := setupStepTargetTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	drill := map[string]interface{}{"id": 2, "template_id": 3, "name": "容灾演练1", "status": "running"}
	if err := db.Table("drill_instance").Create(drill).Error; err != nil {
		t.Fatalf("create drill: %v", err)
	}
	stepName := "操作1.1.1.1：检查通知协调-指挥集结"
	stepTpl := map[string]interface{}{
		"id":                252,
		"drill_template_id": 3,
		"name":              stepName,
		"seq":               4,
		"step_type":         "serial",
	}
	if err := db.Table("drill_template_step").Create(stepTpl).Error; err != nil {
		t.Fatalf("create template step: %v", err)
	}
	step := map[string]interface{}{
		"id":                12,
		"drill_instance_id": 2,
		"step_template_id":  0,
		"name":              stepName,
		"seq":               4,
		"status":            "timeout",
		"assignee_ids":      "[]",
	}
	if err := db.Table("drill_instance_step").Create(step).Error; err != nil {
		t.Fatalf("create instance step: %v", err)
	}

	target, err := resolveStepOperationTarget(2, 12)
	if err != nil {
		t.Fatalf("resolve target: %v", err)
	}
	if target.step.ID != 12 {
		t.Fatalf("expected instance step 12, got %d", target.step.ID)
	}
	if target.stepDefID != 252 {
		t.Fatalf("expected inferred template step 252, got %d", target.stepDefID)
	}

	var updated entity.StepInstance
	if err := db.First(&updated, 12).Error; err != nil {
		t.Fatalf("reload step: %v", err)
	}
	if updated.StepTemplateID != 252 {
		t.Fatalf("expected backfilled template id 252, got %d", updated.StepTemplateID)
	}
}

func TestResolveStepOperationTargetKeepsTemplateIDLookup(t *testing.T) {
	db := setupStepTargetTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	step := map[string]interface{}{
		"id":                80,
		"drill_instance_id": 7,
		"step_template_id":  99,
		"name":              "步骤",
		"seq":               1,
		"status":            "running",
		"assignee_ids":      "[]",
	}
	if err := db.Table("drill_instance_step").Create(step).Error; err != nil {
		t.Fatalf("create instance step: %v", err)
	}

	target, err := resolveStepOperationTarget(7, 99)
	if err != nil {
		t.Fatalf("resolve target: %v", err)
	}
	if target.step.ID != 80 || target.stepDefID != 99 {
		t.Fatalf("unexpected target: step=%d def=%d", target.step.ID, target.stepDefID)
	}
}
