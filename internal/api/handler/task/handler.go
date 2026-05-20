package task

import (
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	taskService *service.TaskService
}

func NewHandler(taskService *service.TaskService) *Handler {
	return &Handler{taskService: taskService}
}

func (h *Handler) GetMyTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tasks, err := h.taskService.GetMyTasks(userID)
	if err != nil {
		response.InternalError(c, "获取任务列表失败")
		return
	}

	response.Success(c, tasks)
}

func (h *Handler) GetDetail(c *gin.Context) {
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的步骤ID")
		return
	}

	task, err := h.taskService.GetTaskDetail(stepID)
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}

	response.Success(c, task)
}

func (h *Handler) CompleteStep(c *gin.Context) {
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的步骤ID")
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Remark = ""
	}

	if err := h.taskService.CompleteStep(stepID, middleware.GetUserID(c), req.Remark); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "任务已完成", nil)
}

func (h *Handler) ReportIssue(c *gin.Context) {
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的步骤ID")
		return
	}

	var req struct {
		IssueDesc string `json:"issue_desc" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "问题描述不能为空")
		return
	}

	if err := h.taskService.ReportIssue(stepID, middleware.GetUserID(c), req.IssueDesc); err != nil {
		response.InternalError(c, "上报问题失败")
		return
	}

	response.SuccessWithMessage(c, "问题已上报", nil)
}
