package template

import (
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	templateService *service.TemplateService
}

func NewHandler(templateService *service.TemplateService) *Handler {
	return &Handler{templateService: templateService}
}

func (h *Handler) List(c *gin.Context) {
	var q dto.PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 20
	}
	q.Normalize()

	category := c.Query("category")
	templates, total, err := h.templateService.GetList(q.Page, q.PageSize, category)
	if err != nil {
		response.InternalError(c, "获取模板列表失败")
		return
	}

	for i := range templates {
		templates[i].StatusLabel = statusToLabel(templates[i].Status)
	}

	response.SuccessPage(c, templates, total, q.Page, q.PageSize)
}

func statusToLabel(status int8) string {
	switch status {
	case 0:
		return "disabled"
	case 2:
		return "enabled"
	default:
		return "disabled"
	}
}

func (h *Handler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	template, err := h.templateService.GetDetail(id)
	if err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	response.Success(c, template)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	template, err := h.templateService.Create(&req, middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, "创建模板失败："+err.Error())
		return
	}

	response.Success(c, template)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req dto.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	if err := h.templateService.Update(id, &req); err != nil {
		response.InternalError(c, "更新模板失败")
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	if err := h.templateService.Delete(id); err != nil {
		response.InternalError(c, "删除模板失败")
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

func (h *Handler) Clone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	template, err := h.templateService.Clone(id)
	if err != nil {
		response.InternalError(c, "复制模板失败")
		return
	}

	response.Success(c, template)
}

func (h *Handler) GetCategories(c *gin.Context) {
	categories, err := h.templateService.GetCategories()
	if err != nil {
		response.InternalError(c, "获取分类失败")
		return
	}

	response.Success(c, categories)
}

func (h *Handler) SaveCategories(c *gin.Context) {
	var categories []entity.TemplateCategory
	if err := c.ShouldBindJSON(&categories); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	if err := h.templateService.SaveCategories(categories); err != nil {
		response.InternalError(c, "保存分类失败")
		return
	}

	response.SuccessWithMessage(c, "保存成功", nil)
}

func (h *Handler) ToggleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	if err := h.templateService.ToggleStatus(id); err != nil {
		response.InternalError(c, "切换状态失败："+err.Error())
		return
	}

	response.SuccessWithMessage(c, "状态已更新", nil)
}

type UpdateStepsRequest struct {
	Steps []struct {
		Name                string  `json:"name" binding:"required,max=200"`
		StepType            string  `json:"step_type" binding:"required,oneof=serial parallel any_of condition"`
		TimeoutMinutes      int     `json:"timeout_minutes"`
		GuideContent        string  `json:"guide_content"`
		IsBlocking          int8    `json:"is_blocking"`
		DefaultAssigneeRole string  `json:"default_assignee_role"`
		ExecutorTeam        string  `json:"executor_team"`
	} `json:"steps" binding:"required,dive"`
}

func (h *Handler) UpdateSteps(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板 ID")
		return
	}

	var req UpdateStepsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	steps := make([]dto.StepTemplateRequest, 0, len(req.Steps))
	for i, s := range req.Steps {
		steps = append(steps, dto.StepTemplateRequest{
			Name:                s.Name,
			Seq:                 i + 1,
			StepType:            s.StepType,
			TimeoutMinutes:      s.TimeoutMinutes,
			GuideContent:        s.GuideContent,
			IsBlocking:          s.IsBlocking,
			DefaultAssigneeRole: s.DefaultAssigneeRole,
			ExecutorTeam:        s.ExecutorTeam,
		})
	}

	if err := h.templateService.UpdateSteps(id, steps); err != nil {
		response.InternalError(c, "更新步骤失败："+err.Error())
		return
	}

	response.SuccessWithMessage(c, "步骤已保存", nil)
}
