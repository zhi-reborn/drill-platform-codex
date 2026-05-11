package notification

import (
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	notificationService *service.NotificationService
}

func NewHandler(notificationService *service.NotificationService) *Handler {
	return &Handler{
		notificationService: notificationService,
	}
}

func (h *Handler) List(c *gin.Context) {
	var query dto.NotificationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	result, err := h.notificationService.GetList(userID, &query)
	if err != nil {
		response.InternalError(c, "获取通知列表失败："+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的通知 ID")
		return
	}

	userID := middleware.GetUserID(c)
	notification, err := h.notificationService.MarkAsRead(userID, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, notification)
}

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.notificationService.MarkAllAsRead(userID); err != nil {
		response.InternalError(c, "标记全部已读失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的通知 ID")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.notificationService.Delete(userID, id); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, nil)
}
