package task

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
	"drill-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeTaskCommandService struct {
	requests []service.SubmitCommandRequest
}

func (f *fakeTaskCommandService) SubmitAndWait(_ context.Context, req service.SubmitCommandRequest) (*service.SubmitCommandResult, error) {
	f.requests = append(f.requests, req)
	return &service.SubmitCommandResult{Command: &entity.FlowCommand{
		ID:              uint64(len(f.requests)),
		CommandType:     req.CommandType,
		DrillInstanceID: req.DrillInstanceID,
		StepInstanceID:  req.StepInstanceID,
		OperatorID:      req.OperatorID,
		IdempotencyKey:  req.IdempotencyKey,
		Status:          entity.FlowCommandSucceeded,
	}}, nil
}

func setupTaskCommandTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE drill_instance_step (
		id INTEGER PRIMARY KEY,
		drill_instance_id INTEGER NOT NULL,
		parent_step_id INTEGER NULL,
		template_step_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		seq INTEGER NOT NULL,
		status TEXT NOT NULL,
		assignee_ids TEXT NOT NULL
	)`).Error; err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Table("drill_instance_step").Create(map[string]interface{}{
		"id":                501,
		"drill_instance_id": 77,
		"template_step_id":  9001,
		"name":              "执行步骤",
		"seq":               1,
		"status":            "pending",
		"assignee_ids":      "[]",
	}).Error; err != nil {
		t.Fatalf("create step: %v", err)
	}
	return db
}

func performTaskCommandRequest(handler gin.HandlerFunc, method, path, body, key string) *httptest.ResponseRecorder {
	router := gin.New()
	router.Handle(method, "/tasks/:stepId/action", func(c *gin.Context) {
		c.Set(middleware.CtxUserIDInt, uint64(42))
		handler(c)
	})
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func TestCommandStartStepSubmitsDurableCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTaskCommandTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	commands := &fakeTaskCommandService{}
	handler := NewHandlerWithCommands(nil, commands)

	resp := performTaskCommandRequest(handler.StartStep, http.MethodPost, "/tasks/501/action", `{}`, "task-start-key")

	assertTaskCommandSubmitted(t, resp, commands, service.SubmitCommandRequest{
		CommandType:     "start_step",
		DrillInstanceID: 77,
		StepInstanceID:  uint64Ptr(501),
		OperatorID:      42,
		IdempotencyKey:  "task-start-key",
		Payload:         map[string]interface{}{},
	})
}

func TestCommandCompleteStepSubmitsDurableCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTaskCommandTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	commands := &fakeTaskCommandService{}
	handler := NewHandlerWithCommands(nil, commands)

	resp := performTaskCommandRequest(handler.CompleteStep, http.MethodPost, "/tasks/501/action", `{"remark":"done"}`, "task-complete-key")

	assertTaskCommandSubmitted(t, resp, commands, service.SubmitCommandRequest{
		CommandType:     "complete_step",
		DrillInstanceID: 77,
		StepInstanceID:  uint64Ptr(501),
		OperatorID:      42,
		IdempotencyKey:  "task-complete-key",
		Payload:         map[string]interface{}{"remark": "done"},
	})
}

func TestCommandReportIssueSubmitsDurableCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTaskCommandTestDB(t)
	origDB := repository.DB
	repository.DB = db
	defer func() { repository.DB = origDB }()

	commands := &fakeTaskCommandService{}
	handler := NewHandlerWithCommands(nil, commands)

	resp := performTaskCommandRequest(handler.ReportIssue, http.MethodPost, "/tasks/501/action", `{"issue_desc":"blocked"}`, "task-issue-key")

	assertTaskCommandSubmitted(t, resp, commands, service.SubmitCommandRequest{
		CommandType:     "report_issue",
		DrillInstanceID: 77,
		StepInstanceID:  uint64Ptr(501),
		OperatorID:      42,
		IdempotencyKey:  "task-issue-key",
		Payload:         map[string]interface{}{"issue_desc": "blocked"},
	})
}

func assertTaskCommandSubmitted(t *testing.T, resp *httptest.ResponseRecorder, commands *fakeTaskCommandService, want service.SubmitCommandRequest) {
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
	if got.StepInstanceID == nil || want.StepInstanceID == nil || *got.StepInstanceID != *want.StepInstanceID {
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
