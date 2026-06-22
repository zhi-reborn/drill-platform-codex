package drill

import (
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/repository"
	"drill-platform/internal/service"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	drillService *service.DrillService
	userService  *service.AuthService
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
	drills, total, err := h.drillService.GetList(q.Page, q.PageSize, status)
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

	// 兜底: 重新从数据库拉取每个 step 的 action_params 并合并到 attributes
	if len(drill.Steps) > 0 {
		ids := make([]uint64, 0, len(drill.Steps))
		for _, s := range drill.Steps {
			ids = append(ids, s.ID)
		}
		type row struct {
			ID          uint64
			ActionParam string
		}
		var rows []row
		repository.DB.Table("drill_instance_step").
			Select("id, action_params").
			Where("id IN ?", ids).
			Scan(&rows)
		m := make(map[uint64]string, len(rows))
		for _, r := range rows {
			m[r.ID] = r.ActionParam
		}
		for i := range drill.Steps {
			if v, ok := m[drill.Steps[i].ID]; ok && v != "" && v != "null" {
				drill.Steps[i].JSONAttributes = v
			}
		}
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	if err := h.drillService.Start(id); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	response.SuccessWithMessage(c, "演练已启动", nil)
}

func (h *Handler) Pause(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	if err := h.drillService.Pause(id); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	response.SuccessWithMessage(c, "演练已暂停", nil)
}

func (h *Handler) Resume(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	if err := h.drillService.Resume(id); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	response.SuccessWithMessage(c, "演练已恢复", nil)
}

func (h *Handler) Terminate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	if err := h.drillService.Terminate(id); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	response.SuccessWithMessage(c, "演练已终止", nil)
}

func (h *Handler) GetLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	logs, err := h.drillService.GetLogs(id)
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	if h.drillService.Engine() == nil {
		response.InternalError(c, "流程引擎未初始化")
		return
	}

	target, err := resolveStepOperationTarget(id, req.StepID)
	if err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}

	stepDefID := int64(target.stepDefID)
	syncEngineStepsFromDB(h.drillService.Engine(), id)

	err = h.drillService.Engine().ManualStartStep(int64(id), stepDefID)
	if err == flowengine.ErrInstanceNotFound {
		if recErr := h.drillService.Recover(id); recErr != nil {
			response.InternalError(c, "恢复演练状态失败")
			return
		}
		syncEngineStepsFromDB(h.drillService.Engine(), id)
		if err = h.drillService.Engine().ManualStartStep(int64(id), stepDefID); err != nil {
			response.BadRequest(c, stepStartErrorMessage(err))
			return
		}
	} else if err != nil {
		response.BadRequest(c, stepStartErrorMessage(err))
		return
	}

	h.drillService.InvalidateStepCache(id)
	response.SuccessWithMessage(c, "步骤已开始", nil)
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	var req struct {
		StepID     uint64                 `json:"step_id" binding:"required"`
		Attributes map[string]interface{} `json:"attributes"` // 业务属性,合并到 action_params
		Remark     string                 `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	target, err := resolveStepOperationTarget(id, req.StepID)
	if err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if len(req.Attributes) > 0 {
		// 读取现有 action_params,合并新属性(用原生 SQL 避免依赖实体字段)
		var actionParams string
		if err := repository.DB.Model(&entity.StepInstance{}).
			Select("action_params").
			Where("id = ?", target.step.ID).
			Scan(&actionParams).Error; err != nil {
			response.InternalError(c, "读取步骤失败: "+err.Error())
			return
		}
		merged := map[string]interface{}{}
		if actionParams != "" && actionParams != "null" {
			_ = json.Unmarshal([]byte(actionParams), &merged)
		}
		for k, v := range req.Attributes {
			if v != nil && fmt.Sprintf("%v", v) != "" {
				merged[k] = v
			}
		}
		buf, _ := json.Marshal(merged)
		updates["action_params"] = string(buf)
	}
	if len(updates) == 0 {
		response.BadRequest(c, "没有需要更新的字段")
		return
	}

	if err := repository.DB.Model(&entity.StepInstance{}).
		Where("id = ?", target.step.ID).
		Updates(updates).Error; err != nil {
		response.InternalError(c, "保存失败: "+err.Error())
		return
	}
	h.drillService.InvalidateStepCache(id)

	response.SuccessWithMessage(c, "步骤信息已更新", nil)
}

func (h *Handler) CompleteStep(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	if h.drillService.Engine() == nil {
		response.InternalError(c, "流程引擎未初始化")
		return
	}

	target, err := resolveStepOperationTarget(id, req.StepID)
	if err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}
	if target.step.Status != string(flowengine.StepStatusRunning) {
		response.BadRequest(c, fmt.Sprintf("步骤「%s」当前状态为%s，不能重复完成", target.step.Name, stepStatusText(target.step.Status)))
		return
	}

	// 先同步引擎内存中的步骤状态（防止引擎状态与DB不一致）
	stepDefID := int64(target.stepDefID)
	if inst, _ := h.drillService.Engine().GetInstanceForMutate(int64(id)); inst != nil {
		if si, ok := inst.Steps[stepDefID]; ok {
			si.Status = flowengine.StepStatus(target.step.Status)
		}
	}

	err = h.drillService.Engine().DirectorCompleteStep(int64(id), stepDefID, int64(middleware.GetUserID(c)))
	if err == flowengine.ErrInstanceNotFound {
		if recErr := h.drillService.Recover(id); recErr != nil {
			response.InternalError(c, "恢复演练状态失败")
			return
		}
		if err = h.drillService.Engine().DirectorCompleteStep(int64(id), stepDefID, int64(middleware.GetUserID(c))); err != nil {
			if err == flowengine.ErrInvalidStatus || err == flowengine.ErrStepNotActive {
				response.BadRequest(c, fmt.Sprintf("步骤「%s」当前状态为%s，不能重复完成", target.step.Name, stepStatusText(target.step.Status)))
				return
			}
			response.InternalError(c, "完成步骤失败："+err.Error())
			return
		}
	} else if err != nil {
		if err == flowengine.ErrInvalidStatus || err == flowengine.ErrStepNotActive {
			response.BadRequest(c, fmt.Sprintf("步骤「%s」当前状态为%s，不能重复完成", target.step.Name, stepStatusText(target.step.Status)))
			return
		}
		response.InternalError(c, "完成步骤失败："+err.Error())
		return
	}

	h.drillService.InvalidateStepCache(id)
	response.SuccessWithMessage(c, "步骤已完成", nil)
}

func (h *Handler) SkipStep(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	if h.drillService.Engine() == nil {
		response.InternalError(c, "流程引擎未初始化")
		return
	}

	contentSkip := req.Remark
	if contentSkip == "" {
		contentSkip = "指挥员手动操作跳过"
	}

	target, err := resolveStepOperationTarget(id, req.StepID)
	if err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}

	// 先同步引擎内存中的步骤状态（防止引擎状态与DB不一致）
	stepDefID := int64(target.stepDefID)
	if inst, _ := h.drillService.Engine().GetInstanceForMutate(int64(id)); inst != nil {
		if si, ok := inst.Steps[stepDefID]; ok {
			si.Status = flowengine.StepStatus(target.step.Status)
		}
	}

	nowS := time.Now()
	err = h.drillService.Engine().DirectorSkipStep(int64(id), stepDefID, int64(middleware.GetUserID(c)))
	if err == flowengine.ErrInstanceNotFound {
		if recErr := h.drillService.Recover(id); recErr != nil {
			response.InternalError(c, "恢复演练状态失败")
			return
		}
		if err = h.drillService.Engine().DirectorSkipStep(int64(id), stepDefID, int64(middleware.GetUserID(c))); err != nil {
			response.InternalError(c, "内部错误")
			return
		}
	} else if err != nil {
		response.InternalError(c, "内部错误")
		return
	}

	repository.DB.Model(&entity.StepInstance{}).
		Where("id = ?", target.step.ID).
		Updates(map[string]interface{}{
			"status":   "skipped",
			"end_time": &nowS,
			"remark":   contentSkip,
		})

	h.drillService.InvalidateStepCache(id)
	response.SuccessWithMessage(c, "步骤已跳过", nil)
}

func (h *Handler) ForceCompleteStep(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	if h.drillService.Engine() == nil {
		response.InternalError(c, "流程引擎未初始化")
		return
	}

	target, err := resolveStepOperationTarget(id, req.StepID)
	if err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}

	// 先同步引擎内存中的步骤状态（防止引擎状态与DB不一致）
	stepDefID := int64(target.stepDefID)
	if inst, _ := h.drillService.Engine().GetInstanceForMutate(int64(id)); inst != nil {
		if si, ok := inst.Steps[stepDefID]; ok {
			si.Status = flowengine.StepStatus(target.step.Status)
		}
	}

	err = h.drillService.Engine().DirectorForceComplete(int64(id), stepDefID, int64(middleware.GetUserID(c)))
	if err == flowengine.ErrInstanceNotFound {
		if recErr := h.drillService.Recover(id); recErr != nil {
			response.InternalError(c, "恢复演练状态失败")
			return
		}
		if err = h.drillService.Engine().DirectorForceComplete(int64(id), stepDefID, int64(middleware.GetUserID(c))); err != nil {
			response.InternalError(c, "内部错误")
			return
		}
	} else if err != nil {
		response.InternalError(c, "内部错误")
		return
	}

	h.drillService.InvalidateStepCache(id)
	response.SuccessWithMessage(c, "步骤已强制完成", nil)
}

func (h *Handler) AssignStep(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}
	var req struct {
		StepID  uint64   `json:"step_id" binding:"required"`
		UserIDs []uint64 `json:"user_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", middleware.GetUserID(c)).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}
	if bytes, err := json.Marshal(req.UserIDs); err == nil {
		repository.DB.Model(&entity.StepInstance{}).
			Where("drill_instance_id = ? AND template_step_id = ?", id, req.StepID).
			Update("assignee_ids", string(bytes))

		now := time.Now()
		repository.DB.Create(&entity.DrillInstanceLog{
			DrillInstanceID: id,
			Action:          "assign",
			OperatorID:      middleware.GetUserID(c),
			OperatorName:    operatorName,
			Content:         fmt.Sprintf("重新指派执行人: %v", req.UserIDs),
			CreatedAt:       now,
		})
	}

	h.drillService.InvalidateStepCache(id)
	response.SuccessWithMessage(c, "执行人已更新", nil)
}

func (h *Handler) ResumeTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练 ID")
		return
	}

	var req struct {
		StepID uint64 `json:"step_id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", middleware.GetUserID(c)).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	if h.drillService.Engine() == nil {
		response.InternalError(c, "流程引擎未初始化")
		return
	}

	target, err := resolveStepOperationTarget(id, req.StepID)
	if err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}
	stepDefID := int64(target.stepDefID)

	if inst, _ := h.drillService.Engine().GetInstanceForMutate(int64(id)); inst != nil {
		if si, ok := inst.Steps[stepDefID]; ok {
			si.Status = flowengine.StepStatus(target.step.Status)
		}
	}

	err = h.drillService.Engine().Intervene(int64(id), flowengine.ActionResumeTask, &stepDefID, int64(middleware.GetUserID(c)))
	if err == flowengine.ErrInstanceNotFound {
		if recErr := h.drillService.Recover(id); recErr != nil {
			response.InternalError(c, "恢复演练状态失败")
			return
		}
		if err = h.drillService.Engine().Intervene(int64(id), flowengine.ActionResumeTask, &stepDefID, int64(middleware.GetUserID(c))); err != nil {
			if err == flowengine.ErrInvalidStatus {
				response.BadRequest(c, "当前状态不支持重新派发")
				return
			}
			response.InternalError(c, "内部错误")
			return
		}
	} else if err != nil {
		if err == flowengine.ErrInvalidStatus {
			response.BadRequest(c, "当前状态不支持重新派发")
			return
		}
		response.InternalError(c, "内部错误")
		return
	}

	// 持久化重新派发状态到DB
	contentRT := req.Remark
	if contentRT == "" {
		contentRT = "指挥员手动重新派发任务"
	}
	nowRT := time.Now()
	var timeoutAtRT *time.Time
	if target.step.TimeoutMinutes > 0 {
		t := nowRT.Add(time.Duration(target.step.TimeoutMinutes) * time.Minute)
		timeoutAtRT = &t
	}
	repository.DB.Model(&entity.StepInstance{}).
		Where("id = ?", target.step.ID).
		Updates(map[string]interface{}{
			"status":          "running",
			"actual_operator": nil,
			"start_time":      nowRT,
			"end_time":        nil,
			"timeout_at":      timeoutAtRT,
			"remark":          "",
			"issue_desc":      "",
		})

	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "resume_task",
		OperatorID:      middleware.GetUserID(c),
		OperatorName:    operatorName,
		Content:         contentRT,
		CreatedAt:       nowRT,
	})

	h.drillService.InvalidateStepCache(id)
	response.SuccessWithMessage(c, "任务已重新派发", nil)
}
