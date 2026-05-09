package report

import (
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	reportService *service.ReportService
}

func NewHandler(reportService *service.ReportService) *Handler {
	return &Handler{reportService: reportService}
}

func (h *Handler) GetReport(c *gin.Context) {
	drillID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练ID")
		return
	}

	report, err := h.reportService.GetReport(drillID)
	if err != nil {
		response.NotFound(c, "演练不存在")
		return
	}

	response.Success(c, report)
}

func (h *Handler) ExportPDF(c *gin.Context) {
	drillID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的演练ID")
		return
	}

	_, err = h.reportService.ExportPDF(drillID)
	if err != nil {
		response.InternalError(c, "导出失败")
		return
	}

	response.SuccessWithMessage(c, "PDF导出功能尚未实现", nil)
}
