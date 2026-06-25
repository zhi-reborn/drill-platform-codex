// Package health exposes /live and /ready endpoints that report process
// liveness and dependency readiness without mutating dependency state.
package health

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"drill-platform/internal/infrastructure/events"
	"drill-platform/internal/worker"

	"github.com/gin-gonic/gin"
)

// HealthChecker is implemented by dependencies that can be pinged for
// readiness. Both *gorm.DB (via repository) and *redis.Client satisfy this
// through small adapters.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// WorkerStatus is the minimal subset of *worker.Worker the handler needs.
// Defining the interface here keeps the handler testable without spinning up
// the real election loop.
type WorkerStatus interface {
	Status() worker.Status
}

// Handler exposes /live and /ready endpoints.
//
// /live is a cheap liveness probe that returns 200 as long as the process is
// running. /ready pings MySQL and Redis, checks the events subscriber health
// flag, and reports the worker status. Readiness checks use a bounded context
// timeout and never mutate dependency state.
type Handler struct {
	db         HealthChecker
	redis      HealthChecker
	subscriber events.Subscriber
	worker     WorkerStatus
	role       string
	instanceID string
	ready      *atomic.Bool
}

// NewHandler constructs a health handler. db and redis must be non-nil for API
// roles; subscriber may be nil when the role does not run an events subscriber;
// worker may be nil for API-only roles. The readiness flag starts false so
// /ready returns 503 until SetReady(true) is called after startup completes.
func NewHandler(db, redis HealthChecker, subscriber events.Subscriber, w WorkerStatus, role, instanceID string) *Handler {
	ready := &atomic.Bool{}
	ready.Store(false)
	return &Handler{
		db:         db,
		redis:      redis,
		subscriber: subscriber,
		worker:     w,
		role:       role,
		instanceID: instanceID,
		ready:      ready,
	}
}

// SetReady toggles the readiness flag. main calls SetReady(true) once the
// HTTP server is listening and SetReady(false) before initiating shutdown so
// load balancers drain traffic during the graceful shutdown window.
func (h *Handler) SetReady(ok bool) {
	h.ready.Store(ok)
}

// Live responds 200 with a minimal body. It is the liveness probe.
func (h *Handler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready pings each dependency with a bounded timeout and returns 200 only when
// all of them are healthy and the readiness flag is true. It never mutates
// dependency state.
func (h *Handler) Ready(c *gin.Context) {
	if !h.ready.Load() {
		h.respondUnready(c, "shutting down")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			h.respondUnready(c, "database unreachable")
			return
		}
	}

	if h.redis != nil {
		if err := h.redis.Ping(ctx); err != nil {
			h.respondUnready(c, "redis unreachable")
			return
		}
	}

	if h.subscriber != nil && !h.subscriber.Healthy() {
		h.respondUnready(c, "subscriber unhealthy")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "ready",
		"role":          h.role,
		"instance_id":   h.instanceID,
		"worker_status": h.workerStatus(),
	})
}

// workerStatus returns the current worker status string, or empty when no
// worker is configured (API-only roles).
func (h *Handler) workerStatus() string {
	if h.worker == nil {
		return ""
	}
	return string(h.worker.Status())
}

func (h *Handler) respondUnready(c *gin.Context, reason string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"status":        "unready",
		"role":          h.role,
		"instance_id":   h.instanceID,
		"worker_status": h.workerStatus(),
		"reason":        reason,
	})
}
