package dashboard

import (
	"context"
	"net/http"

	"drill-platform/internal/api/middleware"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type StatisticsService interface {
	TeamStats(context.Context) (service.TeamStats, error)
	MyTemplateCount(context.Context, uint64) (int64, error)
	AverageStepDuration(context.Context) (*int64, error)
	Touch(context.Context, uint64) error
}

type Handler struct{ service StatisticsService }

func NewHandler(s StatisticsService) *Handler { return &Handler{service: s} }

func (h *Handler) Team(c *gin.Context) {
	if !authorize(c, "admin") {
		return
	}
	stats, err := h.service.TeamStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, "获取团队统计失败")
		return
	}
	response.Success(c, stats)
}

func (h *Handler) MyTemplates(c *gin.Context) {
	if !authorize(c, "admin", "director") {
		return
	}
	count, err := h.service.MyTemplateCount(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, "获取模板统计失败")
		return
	}
	response.Success(c, gin.H{"my_template_count": count})
}

func (h *Handler) StepDuration(c *gin.Context) {
	if !authorize(c, "admin", "director", "viewer") {
		return
	}
	average, err := h.service.AverageStepDuration(c.Request.Context())
	if err != nil {
		response.InternalError(c, "获取步骤耗时失败")
		return
	}
	response.Success(c, gin.H{"avg_step_duration_seconds": average})
}

func (h *Handler) Heartbeat(c *gin.Context) {
	if middleware.GetUserID(c) == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	if err := h.service.Touch(c.Request.Context(), middleware.GetUserID(c)); err != nil {
		response.Error(c, http.StatusServiceUnavailable, http.StatusServiceUnavailable, "在线状态暂不可用")
		return
	}
	response.Success(c, gin.H{"recorded": true})
}

// These endpoints use explicit roles, not the general role hierarchy (executor > viewer).
func authorize(c *gin.Context, roles ...string) bool {
	if middleware.GetUserID(c) == 0 {
		response.Unauthorized(c, "未登录")
		return false
	}
	for _, role := range roles {
		if middleware.GetRole(c) == role {
			return true
		}
	}
	response.Forbidden(c, "无权访问统计数据")
	return false
}
