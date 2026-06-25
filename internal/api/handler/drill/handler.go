package drill

import (
	"context"
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/repository"
	"drill-platform/internal/service"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommandService interface {
	SubmitAndWait(context.Context, service.SubmitCommandRequest) (*service.SubmitCommandResult, error)
}

type Handler struct {
	drillService   *service.DrillService
	userService    *service.AuthService
	commandService CommandService
}

type stepOperationTarget struct {
	step      entity.StepInstance
	stepDefID uint64
}

func NewHandler(drillService *service.DrillService, authService *service.AuthService) *Handler {
	return &Handler{
		drillService: drillService,
		userService:  authService,
	}
}

func NewHandlerWithCommands(drillService *service.DrillService, authService *service.AuthService, commandService CommandService) *Handler {
	return &Handler{
		drillService:   drillService,
		userService:    authService,
		commandService: commandService,
	}
}

func (h *Handler) SetCommandService(commandService CommandService) {
	h.commandService = commandService
}

func stepStatusText(status string) string {
	switch status {
	case string(flowengine.StepStatusPending):
		return "待执行"
	case string(flowengine.StepStatusRunning):
		return "进行中"
	case string(flowengine.StepStatusCompleted):
		return "已完成"
	case string(flowengine.StepStatusSkipped):
		return "已跳过"
	case string(flowengine.StepStatusTimeout):
		return "已超时"
	case string(flowengine.StepStatusIssue):
		return "异常"
	default:
		return status
	}
}

func resolveStepOperationTarget(drillID, requestedID uint64) (*stepOperationTarget, error) {
	var step entity.StepInstance
	if err := repository.DB.Where("drill_instance_id = ? AND id = ?", drillID, requestedID).First(&step).Error; err != nil {
		if err := repository.DB.Where("drill_instance_id = ? AND template_step_id = ?", drillID, requestedID).First(&step).Error; err != nil {
			return nil, err
		}
	}

	if step.StepTemplateID == 0 {
		stepTemplateID, err := inferStepTemplateID(drillID, step)
		if err != nil {
			return nil, err
		}
		step.StepTemplateID = stepTemplateID
		if err := repository.DB.Model(&entity.StepInstance{}).
			Where("id = ?", step.ID).
			Update("template_step_id", stepTemplateID).Error; err != nil {
			return nil, err
		}
	}

	return &stepOperationTarget{
		step:      step,
		stepDefID: step.StepTemplateID,
	}, nil
}

func inferStepTemplateID(drillID uint64, step entity.StepInstance) (uint64, error) {
	var drill entity.DrillInstance
	if err := repository.DB.Select("template_id").First(&drill, drillID).Error; err != nil {
		return 0, err
	}

	var stepTemplate entity.StepTemplate
	err := repository.DB.
		Where("drill_template_id = ? AND seq = ? AND name = ?", drill.TemplateID, step.Seq, step.Name).
		First(&stepTemplate).Error
	if err == nil {
		return stepTemplate.ID, nil
	}

	if err := repository.DB.
		Where("drill_template_id = ? AND seq = ?", drill.TemplateID, step.Seq).
		First(&stepTemplate).Error; err != nil {
		return 0, err
	}
	return stepTemplate.ID, nil
}

func (h *Handler) List(c *gin.Context) {
	var q dto.PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 20
	}
	q.Normalize()

	status := c.Query("status")
	drills, total, err := h.drillService.GetList(q.Page, q.PageSize, status, q.Keyword)
	if err != nil {
		response.InternalError(c, "获取演练列表失败")
		return
	}

	userIDs := make([]uint64, 0, len(drills))
	for _, d := range drills {
		if d.CreatedBy > 0 {
			userIDs = append(userIDs, d.CreatedBy)
		}
	}
	userMap, _ := h.drillService.GetUsersByIDs(userIDs)

	for i := range drills {
		if u, ok := userMap[drills[i].CreatedBy]; ok {
			drills[i].CreatedByName = u.RealName
		}
		if drills[i].Template.ID != 0 {
			drills[i].TemplateName = drills[i].Template.Name
		}
	}

	response.SuccessPage(c, drills, total, q.Page, q.PageSize)
}

func (h *Handler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	drill, err := h.drillService.GetDetail(id)
	if err != nil {
		response.NotFound(c, "演练不存在")
		return
	}

	response.Success(c, drill)
}

func (h *Handler) GetSteps(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	steps, err := h.drillService.GetSteps(id)
	if err != nil {
		response.InternalError(c, "获取步骤列表失败")
		return
	}

	steps = h.drillService.EnrichStepsWithAssigneeNames(steps)
	response.Success(c, steps)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateDrillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	drill, err := h.drillService.Create(&req, middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, "创建演练失败")
		return
	}

	response.Success(c, drill)
}

func (h *Handler) Start(c *gin.Context) {
	h.submitDrillLifecycleCommand(c, "start_drill")
}

func (h *Handler) Pause(c *gin.Context) {
	h.submitDrillLifecycleCommand(c, "pause_drill")
}

func (h *Handler) Resume(c *gin.Context) {
	h.submitDrillLifecycleCommand(c, "resume_drill")
}

func (h *Handler) Terminate(c *gin.Context) {
	h.submitDrillLifecycleCommand(c, "terminate_drill")
}

func (h *Handler) GetLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	limit := 0
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, parseErr := strconv.Atoi(rawLimit)
		if parseErr != nil || parsedLimit <= 0 {
			response.BadRequest(c, "无效的日志数量")
			return
		}
		if parsedLimit > 200 {
			parsedLimit = 200
		}
		limit = parsedLimit
	}

	logs, err := h.drillService.GetLogs(id, limit)
	if err != nil {
		response.InternalError(c, "获取日志失败")
		return
	}

	response.Success(c, logs)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	if err := h.drillService.Delete(id); err != nil {
		response.InternalError(c, "删除演练失败")
		return
	}

	response.SuccessWithMessage(c, "演练已删除", nil)
}

func (h *Handler) StartStep(c *gin.Context) {
	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	h.submitStepCommand(c, "start_step", req.StepID, map[string]interface{}{"step_id": req.StepID, "remark": req.Remark})
}

func syncEngineStepsFromDB(engine *flowengine.Engine, drillID uint64) {
	if engine == nil {
		return
	}
	inst, ok := engine.GetInstanceForMutate(int64(drillID))
	if !ok || inst == nil {
		return
	}

	var steps []entity.StepInstance
	if err := repository.DB.Where("drill_instance_id = ?", drillID).Find(&steps).Error; err != nil {
		return
	}

	instIDToDefID := make(map[uint64]int64, len(steps))
	for _, step := range steps {
		if step.StepTemplateID > 0 {
			instIDToDefID[step.ID] = int64(step.StepTemplateID)
		}
	}

	for _, step := range steps {
		si, exists := inst.Steps[int64(step.StepTemplateID)]
		if !exists {
			continue
		}
		si.ID = int64(step.ID)
		si.Status = flowengine.StepStatus(step.Status)
		si.StartTime = step.StartTime
		si.EndTime = step.EndTime
		si.TimeoutAt = step.TimeoutAt
		si.Remark = step.Remark
		si.IssueDesc = step.IssueDesc
		if step.ActualOperator != nil {
			op := int64(*step.ActualOperator)
			si.ActualOperator = &op
		} else {
			si.ActualOperator = nil
		}

		preDefIDs := make([]int64, 0)
		if step.PreStepIDs != "" && step.PreStepIDs != "[]" && step.PreStepIDs != "null" {
			var preInstIDs []uint64
			if err := json.Unmarshal([]byte(step.PreStepIDs), &preInstIDs); err == nil {
				for _, preInstID := range preInstIDs {
					if preDefID, ok := instIDToDefID[preInstID]; ok {
						preDefIDs = append(preDefIDs, preDefID)
					}
				}
			}
		}
		si.PreStepIDs = preDefIDs
	}
}

func stepStartErrorMessage(err error) string {
	switch err {
	case flowengine.ErrPreStepsNotDone:
		return "前序步骤尚未完成，无法开始"
	case flowengine.ErrInvalidStatus:
		return "步骤状态不允许开始"
	case flowengine.ErrInstanceNotRunning:
		return "演练未在执行中"
	case flowengine.ErrStepNotFound:
		return "步骤不存在于流程引擎"
	default:
		return "内部错误"
	}
}

// UpdateStepInfo 编辑步骤实例的可维护字段
// 注意:数据库实际列名为 action_params(代替 attributes)/ start_time/ end_time 等
func (h *Handler) UpdateStepInfo(c *gin.Context) {
	var req struct {
		StepID     uint64                 `json:"step_id" binding:"required"`
		Attributes map[string]interface{} `json:"attributes"`
		Remark     string                 `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	payload := map[string]interface{}{"step_id": req.StepID, "attributes": req.Attributes}
	if req.Remark != "" {
		payload["remark"] = req.Remark
	}
	h.submitStepCommand(c, "update_step_info", req.StepID, payload)
}

func (h *Handler) CompleteStep(c *gin.Context) {
	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	h.submitStepCommand(c, "complete_step", req.StepID, map[string]interface{}{"step_id": req.StepID, "remark": req.Remark})
}

func (h *Handler) SkipStep(c *gin.Context) {
	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	h.submitStepCommand(c, "skip_step", req.StepID, map[string]interface{}{"step_id": req.StepID, "remark": req.Remark})
}

func (h *Handler) ForceCompleteStep(c *gin.Context) {
	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	h.submitStepCommand(c, "force_complete_step", req.StepID, map[string]interface{}{"step_id": req.StepID, "remark": req.Remark})
}

func (h *Handler) AssignStep(c *gin.Context) {
	var req struct {
		StepID  uint64   `json:"step_id" binding:"required"`
		UserIDs []uint64 `json:"user_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	h.submitStepCommand(c, "assign_step", req.StepID, map[string]interface{}{"step_id": req.StepID, "user_ids": req.UserIDs})
}

func (h *Handler) ResumeTask(c *gin.Context) {
	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	h.submitStepCommand(c, "resume_task", req.StepID, map[string]interface{}{"step_id": req.StepID, "remark": req.Remark})
}

func (h *Handler) submitDrillLifecycleCommand(c *gin.Context, commandType string) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}
	h.submitCommand(c, commandType, id, nil, map[string]interface{}{})
}

func (h *Handler) submitStepCommand(c *gin.Context, commandType string, requestedStepID uint64, payload map[string]interface{}) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}
	target, err := resolveStepOperationTarget(id, requestedStepID)
	if err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}
	h.submitCommand(c, commandType, id, &target.step.ID, payload)
}

func (h *Handler) submitCommand(c *gin.Context, commandType string, drillID uint64, stepID *uint64, payload map[string]interface{}) {
	if h.commandService == nil {
		response.InternalError(c, "命令服务未初始化")
		return
	}
	result, err := h.commandService.SubmitAndWait(c.Request.Context(), service.SubmitCommandRequest{
		CommandType:     commandType,
		DrillInstanceID: drillID,
		StepInstanceID:  stepID,
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
