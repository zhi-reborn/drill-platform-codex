package task

import (
	"context"
	"strconv"

	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/repository"
	"drill-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type CommandService interface {
	SubmitAndWait(context.Context, service.SubmitCommandRequest) (*service.SubmitCommandResult, error)
}

type Handler struct {
	taskService    *service.TaskService
	commandService CommandService
}

func NewHandler(taskService *service.TaskService) *Handler {
	return &Handler{taskService: taskService}
}

func NewHandlerWithCommands(taskService *service.TaskService, commandService CommandService) *Handler {
	return &Handler{taskService: taskService, commandService: commandService}
}

func (h *Handler) SetCommandService(commandService CommandService) {
	h.commandService = commandService
}

func (h *Handler) GetMyTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tasks, err := h.taskService.GetMyTasks(userID)
	if err != nil {
		response.InternalError(c, "获取任务列表失败")
		return
	}

	tasks = h.taskService.EnrichStepsWithAssigneeNames(tasks)
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

	if task != nil {
		tasks := h.taskService.EnrichStepsWithAssigneeNames([]entity.StepInstance{*task})
		if len(tasks) > 0 {
			task = &tasks[0]
		}
	}
	response.Success(c, task)
}

func (h *Handler) StartStep(c *gin.Context) {
	stepID, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的步骤ID")
		return
	}
	step, err := findTaskStep(stepID)
	if err != nil {
		response.InternalError(c, "获取任务失败")
		return
	}

	h.submitCommand(c, "start_step", step.DrillInstanceID, stepID, map[string]interface{}{})
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
	step, err := findTaskStep(stepID)
	if err != nil {
		response.InternalError(c, "获取任务失败")
		return
	}

	h.submitCommand(c, "complete_step", step.DrillInstanceID, stepID, map[string]interface{}{"remark": req.Remark})
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
	step, err := findTaskStep(stepID)
	if err != nil {
		response.InternalError(c, "获取任务失败")
		return
	}

	h.submitCommand(c, "report_issue", step.DrillInstanceID, stepID, map[string]interface{}{"issue_desc": req.IssueDesc})
}

func findTaskStep(stepID uint64) (*entity.StepInstance, error) {
	var step entity.StepInstance
	if err := repository.DB.Select("id", "drill_instance_id").First(&step, stepID).Error; err != nil {
		return nil, err
	}
	return &step, nil
}

func (h *Handler) submitCommand(c *gin.Context, commandType string, drillID, stepID uint64, payload map[string]interface{}) {
	if h.commandService == nil {
		response.InternalError(c, "命令服务未初始化")
		return
	}
	result, err := h.commandService.SubmitAndWait(c.Request.Context(), service.SubmitCommandRequest{
		CommandType:     commandType,
		DrillInstanceID: drillID,
		StepInstanceID:  &stepID,
		OperatorID:      middleware.GetUserID(c),
		IdempotencyKey:  c.GetHeader("Idempotency-Key"),
		Payload:         payload,
	})
	respondCommandResult(c, result, err)
}

func respondCommandResult(c *gin.Context, result *service.SubmitCommandResult, err error) {
	if err != nil {
		response.InternalError(c, "提交命令失败")
		return
	}
	if result == nil || result.Command == nil {
		response.InternalError(c, "提交命令失败")
		return
	}
	cmd := result.Command
	c.Header("Idempotency-Key", cmd.IdempotencyKey)
	if cmd.Status == entity.FlowCommandSucceeded {
		response.Success(c, result)
		return
	}
	if cmd.Status == entity.FlowCommandFailed {
		if cmd.ErrorMessage != nil && *cmd.ErrorMessage != "" {
			response.BadRequest(c, *cmd.ErrorMessage)
			return
		}
		response.InternalError(c, "命令执行失败")
		return
	}
	response.Accepted(c, "命令处理中", result)
}
