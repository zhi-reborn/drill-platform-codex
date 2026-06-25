package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"drill-platform/internal/infrastructure/events"
	"drill-platform/internal/worker"

	"github.com/gin-gonic/gin"
)

// fakeChecker is a configurable HealthChecker stub.
type fakeChecker struct {
	err error
}

func (f fakeChecker) Ping(ctx context.Context) error {
	return f.err
}

// fakeSubscriber satisfies events.Subscriber for the health handler.
type fakeSubscriber struct {
	healthy bool
}

func (f fakeSubscriber) Subscribe(ctx context.Context, fn func(events.Event)) error {
	return nil
}

func (f fakeSubscriber) Healthy() bool {
	return f.healthy
}

// fakeWorker stands in for *worker.Worker so we can drive Status() without
// spinning up the real election loop.
type fakeWorker struct {
	status worker.Status
}

func (f *fakeWorker) Status() worker.Status { return f.status }

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.GET("/live", h.Live)
	r.GET("/ready", h.Ready)
	return r
}

func doRequest(t *testing.T, r *gin.Engine, path string) (*httptest.ResponseRecorder, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode body %q: %v", w.Body.String(), err)
		}
	}
	return w, body
}

func TestLive_Returns200(t *testing.T) {
	h := NewHandler(fakeChecker{}, fakeChecker{}, fakeSubscriber{healthy: true}, nil, "all", "node-a")
	r := newTestRouter(h)

	w, body := doRequest(t, r, "/live")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}
}

func TestReady_AllHealthyReturns200(t *testing.T) {
	h := NewHandler(
		fakeChecker{},
		fakeChecker{},
		fakeSubscriber{healthy: true},
		nil,
		"all",
		"node-a",
	)
	h.SetReady(true)
	r := newTestRouter(h)

	w, body := doRequest(t, r, "/ready")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if body["status"] != "ready" {
		t.Errorf("status = %v, want ready", body["status"])
	}
	if body["role"] != "all" {
		t.Errorf("role = %v, want all", body["role"])
	}
	if body["instance_id"] != "node-a" {
		t.Errorf("instance_id = %v, want node-a", body["instance_id"])
	}
}

func TestReady_RedisUnavailableReturns503(t *testing.T) {
	h := NewHandler(
		fakeChecker{},
		fakeChecker{err: context.DeadlineExceeded},
		fakeSubscriber{healthy: true},
		nil,
		"all",
		"node-a",
	)
	h.SetReady(true)
	r := newTestRouter(h)

	w, body := doRequest(t, r, "/ready")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body=%s", w.Code, w.Body.String())
	}
	if body["status"] != "unready" {
		t.Errorf("status = %v, want unready", body["status"])
	}
}

func TestReady_DBUnavailableReturns503(t *testing.T) {
	h := NewHandler(
		fakeChecker{err: context.DeadlineExceeded},
		fakeChecker{},
		fakeSubscriber{healthy: true},
		nil,
		"all",
		"node-a",
	)
	h.SetReady(true)
	r := newTestRouter(h)

	w, body := doRequest(t, r, "/ready")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body=%s", w.Code, w.Body.String())
	}
	if body["status"] != "unready" {
		t.Errorf("status = %v, want unready", body["status"])
	}
}

func TestReady_SubscriberUnhealthyReturns503(t *testing.T) {
	h := NewHandler(
		fakeChecker{},
		fakeChecker{},
		fakeSubscriber{healthy: false},
		nil,
		"all",
		"node-a",
	)
	h.SetReady(true)
	r := newTestRouter(h)

	w, _ := doRequest(t, r, "/ready")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body=%s", w.Code, w.Body.String())
	}
}

func TestReady_NotReadyFlagReturns503(t *testing.T) {
	h := NewHandler(
		fakeChecker{},
		fakeChecker{},
		fakeSubscriber{healthy: true},
		nil,
		"all",
		"node-a",
	)
	// SetReady(false) by default; readiness flag is off
	r := newTestRouter(h)

	w, _ := doRequest(t, r, "/ready")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body=%s", w.Code, w.Body.String())
	}
}

// TestReady_StandbyWorkerReturns200 verifies that a healthy Worker in standby
// state still returns 200 - standby is a valid state, not an error.
func TestReady_StandbyWorkerReturns200(t *testing.T) {
	fw := &fakeWorker{status: worker.StatusStandby}
	h := NewHandler(
		fakeChecker{},
		fakeChecker{},
		fakeSubscriber{healthy: true},
		fw,
		"worker",
		"node-a",
	)
	h.SetReady(true)
	r := newTestRouter(h)

	w, body := doRequest(t, r, "/ready")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if body["worker_status"] != "standby" {
		t.Errorf("worker_status = %v, want standby", body["worker_status"])
	}
}

func TestReady_WorkerStatusReported(t *testing.T) {
	fw := &fakeWorker{status: worker.StatusLeaderReady}
	h := NewHandler(
		fakeChecker{},
		fakeChecker{},
		fakeSubscriber{healthy: true},
		fw,
		"worker",
		"node-a",
	)
	h.SetReady(true)
	r := newTestRouter(h)

	_, body := doRequest(t, r, "/ready")
	if body["worker_status"] != "leader-ready" {
		t.Errorf("worker_status = %v, want leader-ready", body["worker_status"])
	}
}
