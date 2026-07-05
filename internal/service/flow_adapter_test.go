package service

import (
	"encoding/json"
	"testing"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type flowAdapterTestRedis struct {
	values map[string]string
}

func newFlowAdapterTestRedis() *flowAdapterTestRedis {
	return &flowAdapterTestRedis{values: map[string]string{}}
}

func (r *flowAdapterTestRedis) Get(key string) (string, error) {
	return r.values[key], nil
}

func (r *flowAdapterTestRedis) Set(key string, value interface{}, _ time.Duration) error {
	switch v := value.(type) {
	case []byte:
		r.values[key] = string(v)
	case string:
		r.values[key] = v
	default:
		buf, _ := json.Marshal(v)
		r.values[key] = string(buf)
	}
	return nil
}

func (r *flowAdapterTestRedis) Delete(keys ...string) error {
	for _, key := range keys {
		delete(r.values, key)
	}
	return nil
}

func setupFlowAdapterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE drill_instance_step (
		id INTEGER PRIMARY KEY,
		drill_instance_id INTEGER NOT NULL,
		template_step_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		seq INTEGER NOT NULL,
		status TEXT NOT NULL,
		assignee_ids TEXT NOT NULL,
		start_time DATETIME NULL,
		timeout_at DATETIME NULL
	)`).Error; err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func setupFlowAdapterAutoCompleteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	statements := []string{
		`CREATE TABLE drill_template (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			category TEXT NOT NULL,
			description TEXT,
			status INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			phase_order TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE drill_template_step (
			id INTEGER PRIMARY KEY,
			drill_template_id INTEGER NOT NULL,
			parent_step_id INTEGER NULL,
			name TEXT NOT NULL,
			seq INTEGER NOT NULL,
			step_type TEXT NOT NULL,
			timeout_minutes INTEGER NOT NULL,
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
		)`,
		`CREATE TABLE drill_instance_step (
			id INTEGER PRIMARY KEY,
			drill_instance_id INTEGER NOT NULL,
			template_step_id INTEGER NOT NULL,
			parent_step_id INTEGER NULL,
			name TEXT NOT NULL,
			seq INTEGER NOT NULL,
			status TEXT NOT NULL,
			step_type TEXT NOT NULL,
			assignee_ids TEXT NOT NULL,
			actual_operator INTEGER NULL,
			start_time DATETIME NULL,
			end_time DATETIME NULL,
			timeout_at DATETIME NULL,
			remark TEXT,
			issue_desc TEXT,
			timeout_minutes INTEGER NOT NULL DEFAULT 120,
			default_assignee_role TEXT,
			executor_team TEXT,
			phase TEXT,
			phase_step TEXT,
			pre_step_ids TEXT,
			estimated_duration_minutes INTEGER,
			estimated_start_offset INTEGER,
			action_params TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE drill_instance_step_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			drill_instance_id INTEGER NOT NULL,
			task_instance_id INTEGER NULL,
			command_id INTEGER NULL,
			action TEXT NOT NULL,
			operator_id INTEGER NOT NULL DEFAULT 0,
			operator_name TEXT NOT NULL DEFAULT '',
			content TEXT,
			created_at DATETIME
		)`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}
	return db
}

func TestOnStepStartedInvalidatesStepCache(t *testing.T) {
	db := setupFlowAdapterTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	step := entity.StepInstance{
		ID:              12,
		DrillInstanceID: 33,
		StepTemplateID:  302130,
		Name:            "检查网络设备",
		Seq:             7,
		Status:          "pending",
		AssigneeIDs:     "[]",
	}
	if err := db.Table("drill_instance_step").Create(map[string]interface{}{
		"id":                step.ID,
		"drill_instance_id": step.DrillInstanceID,
		"template_step_id":  step.StepTemplateID,
		"name":              step.Name,
		"seq":               step.Seq,
		"status":            step.Status,
		"assignee_ids":      step.AssigneeIDs,
	}).Error; err != nil {
		t.Fatalf("create step: %v", err)
	}

	redis := newFlowAdapterTestRedis()
	SetCachedSteps(redis, 33, []entity.StepInstance{step})

	adapter := NewDrillFlowAdapter(nil, nil, nil, nil, nil, nil, nil)
	adapter.SetRedis(redis)
	adapter.OnStepStarted(12, time.Now().Add(time.Hour))

	if _, ok := GetCachedSteps(redis, 33); ok {
		t.Fatalf("expected step cache to be invalidated after step start")
	}
}

func TestAutoCompleteParentStepRecursesToAncestorsAndStartsSamePhaseNextStep(t *testing.T) {
	db := setupFlowAdapterAutoCompleteTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	template := entity.DrillTemplate{
		ID:         10,
		Name:       "test template",
		Category:   "test",
		Status:     1,
		CreatedBy:  1,
		PhaseOrder: "[]",
		Steps: []entity.StepTemplate{
			{ID: 100, DrillTemplateID: 10, Name: "phase root", Seq: 1, StepType: "serial", TimeoutMinutes: 5, PreStepIDs: "[]", Phase: "phase-a"},
			{ID: 110, DrillTemplateID: 10, ParentStepID: flowAdapterUint64Ptr(100), Name: "section", Seq: 2, StepType: "serial", TimeoutMinutes: 5, PreStepIDs: "[]", Phase: "phase-a"},
			{ID: 111, DrillTemplateID: 10, ParentStepID: flowAdapterUint64Ptr(110), Name: "child 1", Seq: 3, StepType: "serial", TimeoutMinutes: 5, PreStepIDs: "[]", Phase: "phase-a"},
			{ID: 112, DrillTemplateID: 10, ParentStepID: flowAdapterUint64Ptr(110), Name: "child 2", Seq: 4, StepType: "serial", TimeoutMinutes: 5, PreStepIDs: "[111]", Phase: "phase-a"},
			{ID: 200, DrillTemplateID: 10, Name: "next section", Seq: 5, StepType: "serial", TimeoutMinutes: 5, PreStepIDs: "[100]", Phase: "phase-a"},
			{ID: 210, DrillTemplateID: 10, ParentStepID: flowAdapterUint64Ptr(200), Name: "next child", Seq: 6, StepType: "serial", TimeoutMinutes: 5, PreStepIDs: "[100]", Phase: "phase-a"},
		},
	}
	if err := db.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}

	steps := []entity.StepInstance{
		{ID: 1000, DrillInstanceID: 1, StepTemplateID: 100, Name: "phase root", Seq: 1, Status: "running", StepType: "serial", AssigneeIDs: "[]"},
		{ID: 1100, DrillInstanceID: 1, StepTemplateID: 110, ParentStepID: flowAdapterUint64Ptr(1000), Name: "section", Seq: 2, Status: "running", StepType: "serial", AssigneeIDs: "[]"},
		{ID: 1110, DrillInstanceID: 1, StepTemplateID: 111, ParentStepID: flowAdapterUint64Ptr(1100), Name: "child 1", Seq: 3, Status: "completed", StepType: "serial", AssigneeIDs: "[]"},
		{ID: 1120, DrillInstanceID: 1, StepTemplateID: 112, ParentStepID: flowAdapterUint64Ptr(1100), Name: "child 2", Seq: 4, Status: "completed", StepType: "serial", AssigneeIDs: "[]"},
		{ID: 2000, DrillInstanceID: 1, StepTemplateID: 200, Name: "next section", Seq: 5, Status: "pending", StepType: "serial", AssigneeIDs: "[]"},
		{ID: 2100, DrillInstanceID: 1, StepTemplateID: 210, ParentStepID: flowAdapterUint64Ptr(2000), Name: "next child", Seq: 6, Status: "pending", StepType: "serial", AssigneeIDs: "[]"},
	}
	if err := db.Create(&steps).Error; err != nil {
		t.Fatalf("create steps: %v", err)
	}

	engine := flowengine.NewEngine()
	adapter := NewDrillFlowAdapter(repository.NewTemplateRepo(), nil, nil, nil, nil, nil, nil)
	adapter.engine = engine
	adapter.RegisterDrillContext(1, drillContext{ID: 1, Name: "drill", Status: "running", TemplateID: 10})
	engine.SetCallbacks(adapter)
	engine.SetStepLoader(adapter)

	flowDef := &flowengine.FlowDef{ID: 1, Name: "flow"}
	for _, st := range template.Steps {
		parentID := int64(0)
		if st.ParentStepID != nil {
			parentID = int64(*st.ParentStepID)
		}
		var preIDs []int64
		_ = json.Unmarshal([]byte(st.PreStepIDs), &preIDs)
		flowDef.Steps = append(flowDef.Steps, &flowengine.StepDef{
			ID:           int64(st.ID),
			Name:         st.Name,
			Seq:          st.Seq,
			StepType:     flowengine.StepType(st.StepType),
			ParentStepID: parentID,
			PreStepIDs:   preIDs,
			Phase:        st.Phase,
		})
	}
	inst, err := engine.CreateInstance(flowDef, nil, 1)
	if err != nil {
		t.Fatalf("CreateInstance: %v", err)
	}
	inst.Status = flowengine.FlowStatusRunning
	for _, step := range steps {
		si := inst.Steps[int64(step.StepTemplateID)]
		si.ID = int64(step.ID)
		si.Status = flowengine.StepStatus(step.Status)
	}
	inst.CurrentStepIDs = []int64{100, 110}

	adapter.handleSubtaskCompletion(1, 112)

	if inst.Steps[110].Status != flowengine.StepStatusCompleted {
		t.Fatalf("expected direct parent completed, got %s", inst.Steps[110].Status)
	}
	if inst.Steps[100].Status != flowengine.StepStatusCompleted {
		t.Fatalf("expected ancestor parent completed, got %s", inst.Steps[100].Status)
	}
	if inst.Steps[200].Status != flowengine.StepStatusRunning {
		t.Fatalf("expected same-phase next section running, got %s", inst.Steps[200].Status)
	}
	if inst.Steps[210].Status != flowengine.StepStatusRunning {
		t.Fatalf("expected same-phase next child running, got %s", inst.Steps[210].Status)
	}
}

func flowAdapterUint64Ptr(v uint64) *uint64 {
	return &v
}
