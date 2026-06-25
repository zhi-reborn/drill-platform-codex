package flowcommand

import (
	"fmt"
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

func setupFlowCommandHandlerTestDB(t *testing.T) *gorm.DB {
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

func setupFlowCommandHandlerTestRouter(t *testing.T, db *gorm.DB, userID uint64, role string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := service.NewFlowCommandService(repository.NewFlowCommandRepo(db), 0)
	handler := NewHandler(svc)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.CtxUserIDInt, userID)
		c.Set(middleware.CtxRole, role)
		c.Next()
	})
	router.GET("/api/v1/flow-commands/:id", handler.Get)
	return router
}

func assertCommandResponseDoesNotExposeInternalFields(t *testing.T, body string) {
	t.Helper()
	for _, field := range []string{"idempotency_key", "payload", "worker_id", "lease_until", "attempts"} {
		if strings.Contains(body, field) {
			t.Fatalf("response exposed internal field %q: %s", field, body)
		}
	}
}

func createFlowCommandForHandlerTest(t *testing.T, db *gorm.DB, cmd entity.FlowCommand) entity.FlowCommand {
	t.Helper()
	if cmd.CommandType == "" {
		cmd.CommandType = "start_drill"
	}
	if cmd.DrillInstanceID == 0 {
		cmd.DrillInstanceID = 10
	}
	if cmd.IdempotencyKey == "" {
		cmd.IdempotencyKey = "key"
	}
	if cmd.Payload == "" {
		cmd.Payload = `{}`
	}
	if err := db.Create(&cmd).Error; err != nil {
		t.Fatalf("create command: %v", err)
	}
	return cmd
}

func TestGetCompletedCommandReturnsOK(t *testing.T) {
	db := setupFlowCommandHandlerTestDB(t)
	cmd := createFlowCommandForHandlerTest(t, db, entity.FlowCommand{
		OperatorID:     7,
		IdempotencyKey: "completed-key",
		Status:         entity.FlowCommandSucceeded,
	})
	router := setupFlowCommandHandlerTestRouter(t, db, 7, "executor")

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/flow-commands/%d", cmd.ID), nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	assertCommandResponseDoesNotExposeInternalFields(t, resp.Body.String())
}

func TestGetPendingCommandReturnsAccepted(t *testing.T) {
	db := setupFlowCommandHandlerTestDB(t)
	cmd := createFlowCommandForHandlerTest(t, db, entity.FlowCommand{
		OperatorID:     7,
		IdempotencyKey: "pending-key",
		Status:         entity.FlowCommandPending,
	})
	router := setupFlowCommandHandlerTestRouter(t, db, 7, "executor")

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/flow-commands/%d", cmd.ID), nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestGetProcessingCommandReturnsAccepted(t *testing.T) {
	db := setupFlowCommandHandlerTestDB(t)
	cmd := createFlowCommandForHandlerTest(t, db, entity.FlowCommand{
		OperatorID:     7,
		IdempotencyKey: "processing-key",
		Status:         entity.FlowCommandProcessing,
	})
	router := setupFlowCommandHandlerTestRouter(t, db, 7, "executor")

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/flow-commands/%d", cmd.ID), nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestGetAnotherUsersCommandReturnsNotFound(t *testing.T) {
	db := setupFlowCommandHandlerTestDB(t)
	cmd := createFlowCommandForHandlerTest(t, db, entity.FlowCommand{
		OperatorID:     8,
		IdempotencyKey: "other-key",
		Status:         entity.FlowCommandSucceeded,
	})
	router := setupFlowCommandHandlerTestRouter(t, db, 7, "executor")

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/flow-commands/%d", cmd.ID), nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestGetAnotherUsersCommandAllowsDirector(t *testing.T) {
	db := setupFlowCommandHandlerTestDB(t)
	cmd := createFlowCommandForHandlerTest(t, db, entity.FlowCommand{
		OperatorID:     8,
		IdempotencyKey: "director-visible-key",
		Status:         entity.FlowCommandSucceeded,
	})
	router := setupFlowCommandHandlerTestRouter(t, db, 7, "director")

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/flow-commands/%d", cmd.ID), nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
}
