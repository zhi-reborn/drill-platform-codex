package template

import (
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
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

	response.SuccessPage(c, templates, total, q.Page, q.PageSize)
}

func (h *Handler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
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
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	template, err := h.templateService.Create(&req, middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, "创建模板失败: "+err.Error())
		return
	}

	response.Success(c, template)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	var req dto.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
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
		response.BadRequest(c, "无效的模板ID")
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
		response.BadRequest(c, "无效的模板ID")
		return
	}

	template, err := h.templateService.Clone(id)
	if err != nil {
		response.InternalError(c, "复制模板失败")
		return
	}

	response.Success(c, template)
}
