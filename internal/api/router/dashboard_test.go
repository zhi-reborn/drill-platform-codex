package router

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type dashboardTestRepo struct {
	err    error
	userID uint64
}

func (r *dashboardTestRepo) CountTeam(_ context.Context, ids []uint64) (int64, int64, error) {
	return 9, int64(len(ids)), r.err
}
func (r *dashboardTestRepo) CountMyTemplates(_ context.Context, id uint64) (int64, error) {
	r.userID = id
	return 0, r.err
}
func (r *dashboardTestRepo) AverageStepDuration(context.Context) (*int64, error) { return nil, r.err }

type dashboardTestPresence struct {
	err     error
	touches int
}

func (p *dashboardTestPresence) Touch(context.Context, uint64) error { p.touches++; return p.err }
func (p *dashboardTestPresence) OnlineIDs(context.Context) ([]uint64, error) {
	return []uint64{7}, p.err
}

func dashboardRouter(t *testing.T) (*gin.Engine, *dashboardTestRepo, *dashboardTestPresence) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ws := websocket.NewManager()
	s := service.NewServices(ws, nil)
	r, p := &dashboardTestRepo{}, &dashboardTestPresence{}
	s.DashboardService = service.NewDashboardService(r, p)
	return SetupRouter(s, ws, "dashboard-test-secret", nil, nil), r, p
}

func dashboardRequest(t *testing.T, r *gin.Engine, method, path, role string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if role != "" {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": 7, "username": "test-user", "role": role, "exp": time.Now().Add(time.Hour).Unix(),
		}).SignedString([]byte("dashboard-test-secret"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestDashboardRoutesRoleAccess(t *testing.T) {
	r, _, p := dashboardRouter(t)
	for _, tc := range []struct {
		path, role string
		status     int
	}{
		{"team", "admin", 200}, {"team", "director", 403}, {"team", "viewer", 403}, {"team", "executor", 403},
		{"my-templates", "admin", 200}, {"my-templates", "director", 200}, {"my-templates", "viewer", 403}, {"my-templates", "executor", 403},
		{"step-duration", "admin", 200}, {"step-duration", "director", 200}, {"step-duration", "viewer", 200}, {"step-duration", "executor", 403},
		{"team", "", 401}, {"my-templates", "", 401}, {"step-duration", "", 401},
		{"step-duration", "unknown", 403},
	} {
		t.Run(tc.path+"/"+tc.role, func(t *testing.T) {
			before := p.touches
			w := dashboardRequest(t, r, "GET", "/api/v1/dashboard/"+tc.path, tc.role)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.status, w.Body.String())
			}
			if tc.role == "" && p.touches != before {
				t.Fatal("unauthenticated request recorded")
			}
		})
	}
}

func TestDashboardRoutesJSONAndErrors(t *testing.T) {
	r, repo, p := dashboardRouter(t)
	assertData := func(path, role, field string, want interface{}) {
		t.Helper()
		w := dashboardRequest(t, r, "GET", "/api/v1/dashboard/"+path, role)
		if w.Code != 200 {
			t.Fatalf("status %d: %s", w.Code, w.Body.String())
		}
		var body struct{ Data map[string]interface{} }
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		got, ok := body.Data[field]
		if !ok || got != want {
			t.Fatalf("%s = %v (present %v), want %v", field, got, ok, want)
		}
	}
	assertData("team", "admin", "team_online_count", float64(1))
	assertData("team", "admin", "team_total_count", float64(9))
	assertData("my-templates?user_id=999", "director", "my_template_count", float64(0))
	if repo.userID != 7 {
		t.Fatalf("templates used client user ID: %d", repo.userID)
	}
	assertData("step-duration", "viewer", "avg_step_duration_seconds", nil)
	p.err = errors.New("Redis down")
	assertData("team", "admin", "team_online_count", nil)
	assertData("team", "admin", "team_total_count", float64(9))
	repo.err = errors.New("database down")
	for _, path := range []string{"team", "my-templates", "step-duration"} {
		w := dashboardRequest(t, r, "GET", "/api/v1/dashboard/"+path, "admin")
		if w.Code != 500 {
			t.Fatalf("%s database error = %d", path, w.Code)
		}
	}
}

func TestDashboardHeartbeatAuthenticationAndFailure(t *testing.T) {
	r, _, p := dashboardRouter(t)
	for _, role := range []string{"admin", "director", "viewer", "executor"} {
		before := p.touches
		w := dashboardRequest(t, r, "POST", "/api/v1/auth/heartbeat", role)
		if w.Code != 200 || p.touches != before+1 {
			t.Fatalf("heartbeat %s = %d, touches %d", role, w.Code, p.touches-before)
		}
	}
	before := p.touches
	w := dashboardRequest(t, r, "POST", "/api/v1/auth/heartbeat", "")
	if w.Code != 401 || p.touches != before {
		t.Fatalf("anonymous heartbeat = %d", w.Code)
	}
	p.err = errors.New("Redis down")
	w = dashboardRequest(t, r, "POST", "/api/v1/auth/heartbeat", "admin")
	if w.Code != 503 {
		t.Fatalf("failed heartbeat = %d: %s", w.Code, w.Body.String())
	}
}
