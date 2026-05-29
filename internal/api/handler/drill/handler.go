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

func NewHandler(drillService *service.DrillService, authService *service.AuthService) *Handler {
	return &Handler{
		drillService: drillService,
		userService:  authService,
	}
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

	now := time.Now()
	repository.DB.Model(&entity.StepInstance{}).
		Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).
		Updates(map[string]interface{}{
			"status":          "completed",
			"actual_operator": middleware.GetUserID(c),
			"end_time":        &now,
			"remark":          req.Remark,
		})

	// 通知引擎该步骤已完成
	if h.drillService.Engine() != nil {
		err = h.drillService.Engine().DirectorCompleteStep(int64(id), int64(req.StepID), int64(middleware.GetUserID(c)))
		if err == flowengine.ErrInstanceNotFound {
			if recErr := h.drillService.Recover(id); recErr != nil {
				response.InternalError(c, "恢复演练状态失败")
				return
			}
			if err = h.drillService.Engine().DirectorCompleteStep(int64(id), int64(req.StepID), int64(middleware.GetUserID(c))); err != nil {
				response.InternalError(c, "内部错误")
				return
			}
		} else if err != nil {
			response.InternalError(c, "内部错误")
			return
		}
	}

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

	var stepInst entity.StepInstance
	if err := repository.DB.Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).First(&stepInst).Error; err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}

	nowS := time.Now()
	err = h.drillService.Engine().DirectorSkipStep(int64(id), int64(req.StepID), int64(middleware.GetUserID(c)))
	if err == flowengine.ErrInstanceNotFound {
		if recErr := h.drillService.Recover(id); recErr != nil {
			response.InternalError(c, "恢复演练状态失败")
			return
		}
		if err = h.drillService.Engine().DirectorSkipStep(int64(id), int64(req.StepID), int64(middleware.GetUserID(c))); err != nil {
			response.InternalError(c, "内部错误")
			return
		}
	} else if err != nil {
		response.InternalError(c, "内部错误")
		return
	}

	repository.DB.Model(&entity.StepInstance{}).
		Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).
		Updates(map[string]interface{}{
			"status":   "skipped",
			"end_time": &nowS,
			"remark":   contentSkip,
		})

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

	var stepInst entity.StepInstance
	if err := repository.DB.Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).First(&stepInst).Error; err != nil {
		response.InternalError(c, "步骤实例不存在")
		return
	}

	nowFC := time.Now()
	err = h.drillService.Engine().DirectorForceComplete(int64(id), int64(req.StepID), int64(middleware.GetUserID(c)))
	if err == flowengine.ErrInstanceNotFound {
		if recErr := h.drillService.Recover(id); recErr != nil {
			response.InternalError(c, "恢复演练状态失败")
			return
		}
		if err = h.drillService.Engine().DirectorForceComplete(int64(id), int64(req.StepID), int64(middleware.GetUserID(c))); err != nil {
			response.InternalError(c, "内部错误")
			return
		}
	} else if err != nil {
		response.InternalError(c, "内部错误")
		return
	}

	repository.DB.Model(&entity.StepInstance{}).
		Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).
		Updates(map[string]interface{}{
			"status":          "completed",
			"actual_operator": middleware.GetUserID(c),
			"end_time":        &nowFC,
			"remark":          req.Remark,
		})

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
			Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).
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

	stepDefID := int64(req.StepID)

	var stepDB entity.StepInstance
	repository.DB.Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).First(&stepDB)
	if inst, _ := h.drillService.Engine().GetInstanceForMutate(int64(id)); inst != nil {
		if si, ok := inst.Steps[stepDefID]; ok {
			si.Status = flowengine.StepStatus(stepDB.Status)
		}
	}

	err = h.drillService.Engine().Intervene(int64(id), flowengine.ActionResumeTask, &stepDefID, int64(middleware.GetUserID(c)))
	if err == flowengine.ErrInstanceNotFound {
		if recErr := h.drillService.Recover(id); recErr != nil {
			response.InternalError(c, "恢复演练状态失败")
			return
		}
		if err = h.drillService.Engine().Intervene(int64(id), flowengine.ActionResumeTask, &stepDefID, int64(middleware.GetUserID(c))); err != nil {
			response.InternalError(c, "内部错误")
			return
		}
	} else if err != nil {
		response.InternalError(c, "内部错误")
		return
	}

	// 持久化重新派发状态到DB
	contentRT := req.Remark
	if contentRT == "" {
		contentRT = "指挥员手动重新派发任务"
	}
	repository.DB.Model(&entity.StepInstance{}).
		Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).
		Updates(map[string]interface{}{
			"status":          "running",
			"start_time":      time.Now(),
			"end_time":        nil,
			"timeout_at":      nil,
			"actual_operator": nil,
		})

	nowRT := time.Now()
	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "resume_task",
		OperatorID:      middleware.GetUserID(c),
		OperatorName:    operatorName,
		Content:         contentRT,
		CreatedAt:       nowRT,
	})

	response.SuccessWithMessage(c, "任务已重新派发", nil)
}
