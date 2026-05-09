package display

import (
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	displayService *service.DisplayService
}

func NewHandler(displayService *service.DisplayService) *Handler {
	return &Handler{displayService: displayService}
}

func (h *Handler) GetDrillData(c *gin.Context) {
	drillID, err := strconv.ParseUint(c.Param("drillId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练ID")
		return
	}

	data, err := h.displayService.GetDrillData(drillID)
	if err != nil {
		response.NotFound(c, "演练不存在")
		return
	}

	response.Success(c, data)
}
