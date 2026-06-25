package drill

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"
	drillservice "drill-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeStepCacheRedis struct {
	values  map[string]string
	deleted []string
}

func newFakeStepCacheRedis() *fakeStepCacheRedis {
	return &fakeStepCacheRedis{values: map[string]string{}}
}

func (r *fakeStepCacheRedis) Get(key string) (string, error) {
	return r.values[key], nil
}

func (r *fakeStepCacheRedis) Set(key string, value interface{}, _ time.Duration) error {
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

func (r *fakeStepCacheRedis) Delete(keys ...string) error {
	r.deleted = append(r.deleted, keys...)
	for _, key := range keys {
		delete(r.values, key)
	}
	return nil
}

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
			status TEXT NOT NULL,
			created_by INTEGER NOT NULL DEFAULT 1,
			progress_pct INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE drill_template (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			created_by INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE drill_instance_step (
			id INTEGER PRIMARY KEY,
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
			pre_step_ids TEXT,
			action_params TEXT
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

func TestGetDetailDoesNotReturnStepsPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStepTargetTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	if err := db.Table("drill_template").Create(map[string]interface{}{
		"id":   3,
		"name": "模板",
	}).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	if err := db.Table("drill_instance").Create(map[string]interface{}{
		"id":          2,
		"template_id": 3,
		"name":        "大步骤演练",
		"status":      "running",
	}).Error; err != nil {
		t.Fatalf("create drill: %v", err)
	}
	for _, step := range []map[string]interface{}{
		{"id": 10, "drill_instance_id": 2, "template_step_id": 100, "name": "步骤1", "seq": 1, "status": "completed", "assignee_ids": "[]"},
		{"id": 11, "drill_instance_id": 2, "template_step_id": 101, "name": "步骤2", "seq": 2, "status": "running", "assignee_ids": "[]"},
	} {
		if err := db.Table("drill_instance_step").Create(step).Error; err != nil {
			t.Fatalf("create step: %v", err)
		}
	}

	svc := drillservice.NewDrillService(
		repository.NewDrillRepo(),
		repository.NewTemplateRepo(),
		repository.NewStepRepo(),
		repository.NewUserRepo(),
	)
	handler := NewHandler(svc, nil)

	router := gin.New()
	router.GET("/drills/:id", handler.GetDetail)
	req := httptest.NewRequest(http.MethodGet, "/drills/2", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected response data object, got %T", body["data"])
	}
	if _, ok := data["steps"]; ok {
		t.Fatalf("expected detail payload to omit steps, got %s", resp.Body.String())
	}
}

func TestSyncEngineStepsFromDBAllowsStartAfterDBPredecessorCompleted(t *testing.T) {
	db := setupStepTargetTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	steps := []map[string]interface{}{
		{
			"id":                1,
			"drill_instance_id": 2,
			"template_step_id":  10,
			"name":              "前序任务",
			"seq":               1,
			"status":            "completed",
			"assignee_ids":      "[]",
			"pre_step_ids":      "[]",
		},
		{
			"id":                2,
			"drill_instance_id": 2,
			"template_step_id":  20,
			"name":              "目标任务",
			"seq":               2,
			"status":            "pending",
			"assignee_ids":      "[]",
			"pre_step_ids":      "[1]",
		},
	}
	for _, step := range steps {
		if err := db.Table("drill_instance_step").Create(step).Error; err != nil {
			t.Fatalf("create step: %v", err)
		}
	}

	engine := flowengine.NewEngine()
	inst, err := engine.CreateInstance(&flowengine.FlowDef{
		ID:   2,
		Name: "演练",
		Steps: []*flowengine.StepDef{
			{ID: 10, Name: "前序任务", Seq: 1, StepType: flowengine.StepTypeSerial},
			{ID: 20, Name: "目标任务", Seq: 2, StepType: flowengine.StepTypeSerial, PreStepIDs: []int64{10}},
		},
	}, nil, 1)
	if err != nil {
		t.Fatalf("create flow instance: %v", err)
	}
	inst.Status = flowengine.FlowStatusRunning

	if err := engine.ManualStartStep(2, 20); err != flowengine.ErrPreStepsNotDone {
		t.Fatalf("expected stale engine state to block start, got %v", err)
	}

	syncEngineStepsFromDB(engine, 2)

	if err := engine.ManualStartStep(2, 20); err != nil {
		t.Fatalf("expected synced engine state to allow start, got %v", err)
	}
}

func TestUpdateStepInfoSubmitsCommandWithoutInvalidatingStepCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStepTargetTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	step := map[string]interface{}{
		"id":                12,
		"drill_instance_id": 2,
		"template_step_id":  252,
		"name":              "操作1",
		"seq":               1,
		"status":            "running",
		"assignee_ids":      "[]",
		"action_params":     `{"operator":"旧操作人"}`,
	}
	if err := db.Table("drill_instance_step").Create(step).Error; err != nil {
		t.Fatalf("create instance step: %v", err)
	}

	redis := newFakeStepCacheRedis()
	drillservice.SetCachedSteps(redis, 2, []entity.StepInstance{{
		ID:              12,
		DrillInstanceID: 2,
		StepTemplateID:  252,
		Name:            "操作1",
		Seq:             1,
		Status:          "running",
		AssigneeIDs:     "[]",
		JSONAttributes:  `{"operator":"旧操作人"}`,
	}})

	commands := &fakeDrillCommandService{}
	handler := NewHandlerWithCommands(nil, nil, commands)

	router := gin.New()
	router.PUT("/drills/:id/steps/info", func(c *gin.Context) {
		c.Set(middleware.CtxUserIDInt, uint64(42))
		handler.UpdateStepInfo(c)
	})
	body := bytes.NewBufferString(`{"step_id":252,"attributes":{"operator":"新操作人"}}`)
	req := httptest.NewRequest(http.MethodPut, "/drills/2/steps/info", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "update-info-key")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assertDrillCommandSubmitted(t, resp, commands, drillservice.SubmitCommandRequest{
		CommandType:     "update_step_info",
		DrillInstanceID: 2,
		StepInstanceID:  uint64Ptr(12),
		OperatorID:      42,
		IdempotencyKey:  "update-info-key",
		Payload: map[string]interface{}{
			"step_id":    float64(252),
			"attributes": map[string]interface{}{"operator": "新操作人"},
		},
	})
	if _, ok := drillservice.GetCachedSteps(redis, 2); !ok {
		t.Fatalf("expected handler to leave step cache untouched until command execution")
	}
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
		"template_step_id":  0,
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
		"template_step_id":  99,
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

func TestResolveStepOperationTargetPrefersInstanceIDWhenIDsCollide(t *testing.T) {
	db := setupStepTargetTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	steps := []map[string]interface{}{
		{
			"id":                80,
			"drill_instance_id": 7,
			"template_step_id":  99,
			"name":              "模板 ID 冲突步骤",
			"seq":               1,
			"status":            "completed",
			"assignee_ids":      "[]",
		},
		{
			"id":                99,
			"drill_instance_id": 7,
			"template_step_id":  100,
			"name":              "真正要操作的实例步骤",
			"seq":               2,
			"status":            "running",
			"assignee_ids":      "[]",
		},
	}
	for _, step := range steps {
		if err := db.Table("drill_instance_step").Create(step).Error; err != nil {
			t.Fatalf("create instance step: %v", err)
		}
	}

	target, err := resolveStepOperationTarget(7, 99)
	if err != nil {
		t.Fatalf("resolve target: %v", err)
	}
	if target.step.ID != 99 || target.stepDefID != 100 {
		t.Fatalf("expected instance step 99/template step 100, got step=%d def=%d", target.step.ID, target.stepDefID)
	}
}

type fakeDrillCommandService struct {
	requests []drillservice.SubmitCommandRequest
}

func (f *fakeDrillCommandService) SubmitAndWait(_ context.Context, req drillservice.SubmitCommandRequest) (*drillservice.SubmitCommandResult, error) {
	f.requests = append(f.requests, req)
	return &drillservice.SubmitCommandResult{Command: &entity.FlowCommand{
		ID:              uint64(len(f.requests)),
		CommandType:     req.CommandType,
		DrillInstanceID: req.DrillInstanceID,
		StepInstanceID:  req.StepInstanceID,
		OperatorID:      req.OperatorID,
		IdempotencyKey:  req.IdempotencyKey,
		Status:          entity.FlowCommandSucceeded,
	}}, nil
}

func TestCommandDrillLifecycleSubmitsDurableCommands(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		commandType string
		handler     func(*Handler) gin.HandlerFunc
	}{
		{name: "start", commandType: "start_drill", handler: func(h *Handler) gin.HandlerFunc { return h.Start }},
		{name: "pause", commandType: "pause_drill", handler: func(h *Handler) gin.HandlerFunc { return h.Pause }},
		{name: "resume", commandType: "resume_drill", handler: func(h *Handler) gin.HandlerFunc { return h.Resume }},
		{name: "terminate", commandType: "terminate_drill", handler: func(h *Handler) gin.HandlerFunc { return h.Terminate }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := &fakeDrillCommandService{}
			handler := NewHandlerWithCommands(nil, nil, commands)

			resp := performDrillCommandRequest(tt.handler(handler), http.MethodPost, "/drills/88/action", `{}`, "drill-"+tt.name+"-key")

			assertDrillCommandSubmitted(t, resp, commands, drillservice.SubmitCommandRequest{
				CommandType:     tt.commandType,
				DrillInstanceID: 88,
				OperatorID:      42,
				IdempotencyKey:  "drill-" + tt.name + "-key",
				Payload:         map[string]interface{}{},
			})
		})
	}
}

func TestCommandDrillStepMutationsSubmitDurableCommands(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupStepTargetTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	if err := db.Table("drill_instance_step").Create(map[string]interface{}{
		"id":                601,
		"drill_instance_id": 88,
		"template_step_id":  7001,
		"name":              "指挥步骤",
		"seq":               1,
		"status":            "running",
		"assignee_ids":      "[]",
		"action_params":     `{}`,
	}).Error; err != nil {
		t.Fatalf("create instance step: %v", err)
	}

	tests := []struct {
		name        string
		commandType string
		handler     func(*Handler) gin.HandlerFunc
		body        string
		payload     map[string]interface{}
	}{
		{name: "start", commandType: "start_step", handler: func(h *Handler) gin.HandlerFunc { return h.StartStep }, body: `{"step_id":601,"remark":"start it"}`, payload: map[string]interface{}{"step_id": float64(601), "remark": "start it"}},
		{name: "complete", commandType: "complete_step", handler: func(h *Handler) gin.HandlerFunc { return h.CompleteStep }, body: `{"step_id":601,"remark":"done"}`, payload: map[string]interface{}{"step_id": float64(601), "remark": "done"}},
		{name: "skip", commandType: "skip_step", handler: func(h *Handler) gin.HandlerFunc { return h.SkipStep }, body: `{"step_id":601,"remark":"skip it"}`, payload: map[string]interface{}{"step_id": float64(601), "remark": "skip it"}},
		{name: "force_complete", commandType: "force_complete_step", handler: func(h *Handler) gin.HandlerFunc { return h.ForceCompleteStep }, body: `{"step_id":601,"remark":"force"}`, payload: map[string]interface{}{"step_id": float64(601), "remark": "force"}},
		{name: "resume_task", commandType: "resume_task", handler: func(h *Handler) gin.HandlerFunc { return h.ResumeTask }, body: `{"step_id":601,"remark":"retry"}`, payload: map[string]interface{}{"step_id": float64(601), "remark": "retry"}},
		{name: "assign", commandType: "assign_step", handler: func(h *Handler) gin.HandlerFunc { return h.AssignStep }, body: `{"step_id":601,"user_ids":[11,12]}`, payload: map[string]interface{}{"step_id": float64(601), "user_ids": []interface{}{float64(11), float64(12)}}},
		{name: "update_info", commandType: "update_step_info", handler: func(h *Handler) gin.HandlerFunc { return h.UpdateStepInfo }, body: `{"step_id":601,"attributes":{"operator":"new"},"remark":"updated"}`, payload: map[string]interface{}{"step_id": float64(601), "attributes": map[string]interface{}{"operator": "new"}, "remark": "updated"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := &fakeDrillCommandService{}
			handler := NewHandlerWithCommands(nil, nil, commands)

			resp := performDrillCommandRequest(tt.handler(handler), http.MethodPost, "/drills/88/action", tt.body, "drill-step-"+tt.name+"-key")

			assertDrillCommandSubmitted(t, resp, commands, drillservice.SubmitCommandRequest{
				CommandType:     tt.commandType,
				DrillInstanceID: 88,
				StepInstanceID:  uint64Ptr(601),
				OperatorID:      42,
				IdempotencyKey:  "drill-step-" + tt.name + "-key",
				Payload:         tt.payload,
			})
		})
	}
}

func performDrillCommandRequest(handler gin.HandlerFunc, method, path, body, key string) *httptest.ResponseRecorder {
	router := gin.New()
	router.Handle(method, "/drills/:id/action", func(c *gin.Context) {
		c.Set(middleware.CtxUserIDInt, uint64(42))
		handler(c)
	})
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func assertDrillCommandSubmitted(t *testing.T, resp *httptest.ResponseRecorder, commands *fakeDrillCommandService, want drillservice.SubmitCommandRequest) {
	t.Helper()
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
	}
	if resp.Header().Get("Idempotency-Key") != want.IdempotencyKey {
		t.Fatalf("expected response idempotency key %q, got %q", want.IdempotencyKey, resp.Header().Get("Idempotency-Key"))
	}
	if len(commands.requests) != 1 {
		t.Fatalf("expected one submitted command, got %d", len(commands.requests))
	}
	got := commands.requests[0]
	if got.CommandType != want.CommandType || got.DrillInstanceID != want.DrillInstanceID || got.OperatorID != want.OperatorID || got.IdempotencyKey != want.IdempotencyKey {
		t.Fatalf("unexpected command request: %#v", got)
	}
	if want.StepInstanceID == nil {
		if got.StepInstanceID != nil {
			t.Fatalf("expected nil step instance id, got %v", *got.StepInstanceID)
		}
	} else if got.StepInstanceID == nil || *got.StepInstanceID != *want.StepInstanceID {
		t.Fatalf("expected step instance id %v, got %v", want.StepInstanceID, got.StepInstanceID)
	}
	gotPayload, _ := json.Marshal(got.Payload)
	wantPayload, _ := json.Marshal(want.Payload)
	if string(gotPayload) != string(wantPayload) {
		t.Fatalf("expected payload %s, got %s", wantPayload, gotPayload)
	}
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}
