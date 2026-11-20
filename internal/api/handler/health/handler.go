// Package health exposes /live and /ready endpoints that report process
// liveness and dependency readiness without mutating dependency state.
// It also serves GET /api/v1/system/health (admin only), which aggregates
// every backend node's heartbeat snapshot from the Redis node registry.
package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"sort"
	"sync/atomic"
	"time"

	"drill-platform/internal/infrastructure/events"
	"drill-platform/internal/infrastructure/nodereg"
	"drill-platform/internal/pkg/response"
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

// ComponentStatus is one dependency check result plus its ping latency.
type ComponentStatus struct {
	Status    string `json:"status"` // ok | fail
	LatencyMs int64  `json:"latency_ms"`
}

// DBPoolStats is the useful subset of sql.DBStats for the health page.
type DBPoolStats struct {
	OpenConnections int   `json:"open"`
	InUse           int   `json:"in_use"`
	Idle            int   `json:"idle"`
	WaitCount       int64 `json:"wait_count"`
	WaitDurationMs  int64 `json:"wait_duration_ms"`
}

// Snapshot is a point-in-time health report for one node. It is both the
// heartbeat payload stored in Redis and the per-node body in SystemHealthResp.
type Snapshot struct {
	InstanceID     string                     `json:"instance_id"`
	Role           string                     `json:"role"`
	Ready          bool                       `json:"ready"`
	WorkerStatus   string                     `json:"worker_status"`
	Components     map[string]ComponentStatus `json:"components"`
	WsConnections  map[string]int             `json:"ws_connections,omitempty"`
	Goroutines     int                        `json:"goroutines"`
	HeapAllocBytes uint64                     `json:"heap_alloc_bytes"`
	SysBytes       uint64                     `json:"sys_bytes"`
	UptimeSeconds  int64                      `json:"uptime_seconds"`
	StartedAt      int64                      `json:"started_at"`
	DBPool         *DBPoolStats               `json:"db_pool,omitempty"`
}

// NodeHealth is one node's snapshot plus its registry state.
type NodeHealth struct {
	Snapshot
	HeartbeatAt int64  `json:"heartbeat_at"` // unix seconds on the Redis clock
	State       string `json:"state"`        // online | offline
}

// SystemHealthResp is the aggregated view for the admin health page.
type SystemHealthResp struct {
	Overall   string       `json:"overall"` // healthy | degraded
	UpdatedAt int64        `json:"updated_at"`
	Warning   string       `json:"warning,omitempty"`
	Nodes     []NodeHealth `json:"nodes"`
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
	publisher  HealthChecker
	subscriber events.Subscriber
	worker     WorkerStatus
	role       string
	instanceID string
	ready      *atomic.Bool
	registry   *nodereg.Registry
	wsStats    func() map[string]interface{}
	dbStats    func() *sql.DBStats
	startedAt  time.Time
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
		startedAt:  time.Now(),
	}
}

// SetPublisher wires the event publisher readiness check. In HA deployments a
// node must be drained if it cannot publish Redis events, even if it can still
// answer ordinary HTTP requests.
func (h *Handler) SetPublisher(p HealthChecker) {
	h.publisher = p
}

// SetRegistry wires the Redis node registry for multi-node aggregation.
func (h *Handler) SetRegistry(r *nodereg.Registry) {
	h.registry = r
}

// SetWSStats wires a callback returning per-channel WebSocket client counts.
func (h *Handler) SetWSStats(fn func() map[string]interface{}) {
	h.wsStats = fn
}

// SetDBStats wires a callback returning the MySQL connection pool stats.
func (h *Handler) SetDBStats(fn func() *sql.DBStats) {
	h.dbStats = fn
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

// checkComponents pings each configured dependency with the given context and
// records status plus latency. It never mutates dependency state.
func (h *Handler) checkComponents(ctx context.Context) (map[string]ComponentStatus, string) {
	components := map[string]ComponentStatus{}
	reason := ""

	ping := func(name, reasonText string, checker HealthChecker) {
		if checker == nil {
			return
		}
		start := time.Now()
		err := checker.Ping(ctx)
		st := ComponentStatus{Status: "ok", LatencyMs: time.Since(start).Milliseconds()}
		if err != nil {
			st.Status = "fail"
			reason = firstReason(reason, reasonText)
		}
		components[name] = st
	}

	ping("db", "database unreachable", h.db)
	ping("redis", "redis unreachable", h.redis)
	ping("publisher", "publisher unreachable", h.publisher)

	if h.subscriber != nil {
		st := ComponentStatus{Status: "ok"}
		if !h.subscriber.Healthy() {
			st.Status = "fail"
			reason = firstReason(reason, "subscriber unhealthy")
		}
		components["subscriber"] = st
	}
	return components, reason
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

	statuses, reason := h.checkComponents(ctx)
	components := make(map[string]string, len(statuses))
	for name, st := range statuses {
		components[name] = st.Status
	}

	if reason != "" {
		h.respondUnready(c, reason, components)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "ready",
		"role":          h.role,
		"instance_id":   h.instanceID,
		"worker_status": h.workerStatus(),
		"components":    components,
	})
}

// BuildSnapshot collects a fresh health snapshot for this node: dependency
// checks with latency, Go runtime metrics, and the injected WebSocket/DB pool
// stats. Dependency pings use a bounded timeout and never mutate state.
func (h *Handler) BuildSnapshot(ctx context.Context) Snapshot {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	components, _ := h.checkComponents(ctx)

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	snap := Snapshot{
		InstanceID:     h.instanceID,
		Role:           h.role,
		Ready:          h.ready.Load(),
		WorkerStatus:   h.workerStatus(),
		Components:     components,
		Goroutines:     runtime.NumGoroutine(),
		HeapAllocBytes: ms.HeapAlloc,
		SysBytes:       ms.Sys,
		UptimeSeconds:  int64(time.Since(h.startedAt) / time.Second),
		StartedAt:     h.startedAt.Unix(),
	}

	if h.wsStats != nil {
		ws := make(map[string]int)
		for ch, v := range h.wsStats() {
			if n, ok := toInt(v); ok {
				ws[ch] = n
			}
		}
		if len(ws) > 0 {
			snap.WsConnections = ws
		}
	}
	if h.dbStats != nil {
		if s := h.dbStats(); s != nil {
			snap.DBPool = &DBPoolStats{
				OpenConnections: s.OpenConnections,
				InUse:           s.InUse,
				Idle:            s.Idle,
				WaitCount:       s.WaitCount,
				WaitDurationMs:  s.WaitDuration.Milliseconds(),
			}
		}
	}
	return snap
}

// SystemHealth serves GET /api/v1/system/health (admin only). It returns the
// local node's fresh snapshot aggregated with every node registered in Redis.
// When Redis (and thus the registry) is unavailable it degrades to a
// single-node report instead of failing: the health page is the first place
// operators look when Redis breaks.
func (h *Handler) SystemHealth(c *gin.Context) {
	local := h.BuildSnapshot(c.Request.Context())
	resp := SystemHealthResp{
		UpdatedAt: time.Now().Unix(),
		Nodes:     []NodeHealth{{Snapshot: local, State: "online"}},
	}

	if h.registry == nil {
		resp.Warning = "节点注册中心未启用（Redis 未配置），仅展示当前节点"
		resp.Overall = overallOf(resp.Nodes)
		response.Success(c, resp)
		return
	}

	records, redisNow, err := h.registry.List(c.Request.Context())
	if err != nil {
		resp.Warning = "节点注册中心不可用：" + err.Error() + "，仅展示当前节点"
		resp.Overall = overallOf(resp.Nodes)
		response.Success(c, resp)
		return
	}

	nodes := make([]NodeHealth, 0, len(records)+1)
	seenLocal := false
	for _, rec := range records {
		var snap Snapshot
		if err := json.Unmarshal(rec.Data, &snap); err != nil || snap.InstanceID == "" {
			continue
		}
		if rec.InstanceID == h.instanceID {
			seenLocal = true
			nodes = append(nodes, NodeHealth{Snapshot: local, HeartbeatAt: rec.HeartbeatAt, State: "online"})
			continue
		}
		node := NodeHealth{Snapshot: snap, HeartbeatAt: rec.HeartbeatAt}
		if redisNow-rec.HeartbeatAt > int64(nodereg.OfflineAfter/time.Second) {
			node.State = "offline"
		} else {
			node.State = "online"
		}
		nodes = append(nodes, node)
	}
	if !seenLocal {
		nodes = append(nodes, NodeHealth{Snapshot: local, State: "online"})
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].State != nodes[j].State {
			return nodes[i].State == "online"
		}
		return nodes[i].InstanceID < nodes[j].InstanceID
	})

	resp.Nodes = nodes
	resp.Overall = overallOf(nodes)
	response.Success(c, resp)
}

// overallOf derives the aggregate status: any offline node, or any online node
// that is unready or has a failed component, degrades the system.
func overallOf(nodes []NodeHealth) string {
	for _, n := range nodes {
		if n.State != "online" {
			continue
		}
		if !n.Ready {
			return "degraded"
		}
		for _, comp := range n.Components {
			if comp.Status == "fail" {
				return "degraded"
			}
		}
	}
	for _, n := range nodes {
		if n.State == "offline" {
			return "degraded"
		}
	}
	return "healthy"
}

// workerStatus returns the current worker status string, or empty when no
// worker is configured (API-only roles).
func (h *Handler) workerStatus() string {
	if h.worker == nil {
		return ""
	}
	return string(h.worker.Status())
}

func (h *Handler) respondUnready(c *gin.Context, reason string, components ...map[string]string) {
	body := gin.H{
		"status":        "unready",
		"role":          h.role,
		"instance_id":   h.instanceID,
		"worker_status": h.workerStatus(),
		"reason":        reason,
	}
	if len(components) > 0 && len(components[0]) > 0 {
		body["components"] = components[0]
	}
	c.JSON(http.StatusServiceUnavailable, body)
}

func firstReason(current, next string) string {
	if current != "" {
		return current
	}
	return next
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case uint:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}
