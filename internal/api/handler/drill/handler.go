package drill

import (
	"encoding/json"
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/repository"
	"drill-platform/internal/service"
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

	for i := range drills {
		if user, err := h.drillService.GetUserByID(drills[i].CreatedBy); err == nil {
			drills[i].CreatedByName = user.RealName
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

	steps = repository.EnrichAssigneeNames(steps)
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
		response.InternalError(c, "创建演练失败："+err.Error())
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
		response.BadRequest(c, err.Error())
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
		response.BadRequest(c, err.Error())
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
		response.BadRequest(c, err.Error())
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
		response.BadRequest(c, err.Error())
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
		response.InternalError(c, "删除演练失败："+err.Error())
		return
	}

	response.SuccessWithMessage(c, "演练已删除", nil)
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

	stepDefID := int64(req.StepID)
	if err := h.drillService.Engine().Intervene(int64(id), flowengine.ActionSkip, &stepDefID, int64(middleware.GetUserID(c))); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", middleware.GetUserID(c)).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}
	now := time.Now()
	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "skip",
		OperatorID:      middleware.GetUserID(c),
		OperatorName:    operatorName,
		Content:         req.Remark,
		CreatedAt:       now,
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

	stepDefID := int64(req.StepID)
	if err := h.drillService.Engine().Intervene(int64(id), flowengine.ActionForceComplete, &stepDefID, int64(middleware.GetUserID(c))); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", middleware.GetUserID(c)).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}
	now := time.Now()
	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "force_complete",
		OperatorID:      middleware.GetUserID(c),
		OperatorName:    operatorName,
		Content:         req.Remark,
		CreatedAt:       now,
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

	if bytes, err := json.Marshal(req.UserIDs); err == nil {
		repository.DB.Model(&entity.StepInstance{}).
			Where("drill_instance_id = ? AND step_template_id = ?", id, req.StepID).
			Update("assignee_ids", string(bytes))
	}

	response.SuccessWithMessage(c, "执行人已更新", nil)
}
