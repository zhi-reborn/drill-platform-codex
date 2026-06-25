package flowcommand

import (
	"errors"
	"strconv"
	"time"

	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	flowCommandService *service.FlowCommandService
}

type CommandStatusResponse struct {
	ID              uint64                   `json:"id"`
	CommandType     string                   `json:"command_type"`
	DrillInstanceID uint64                   `json:"drill_instance_id"`
	StepInstanceID  *uint64                  `json:"step_instance_id,omitempty"`
	Status          entity.FlowCommandStatus `json:"status"`
	Result          *string                  `json:"result,omitempty"`
	ErrorCode       *string                  `json:"error_code,omitempty"`
	ErrorMessage    *string                  `json:"error_message,omitempty"`
	CreatedAt       string                   `json:"created_at"`
	FinishedAt      *string                  `json:"finished_at,omitempty"`
}

func NewHandler(flowCommandService *service.FlowCommandService) *Handler {
	return &Handler{flowCommandService: flowCommandService}
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的命令ID")
		return
	}

	var cmd *entity.FlowCommand
	role := middleware.GetRole(c)
	if role == "director" || role == "admin" {
		cmd, err = h.flowCommandService.Get(id)
	} else {
		cmd, err = h.flowCommandService.GetForOperator(id, middleware.GetUserID(c))
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "命令不存在")
			return
		}
		response.InternalError(c, "获取命令状态失败")
		return
	}

	data := toCommandStatusResponse(cmd)
	if cmd.IsTerminal() {
		response.Success(c, data)
		return
	}
	response.Accepted(c, "命令处理中", data)
}

func toCommandStatusResponse(cmd *entity.FlowCommand) CommandStatusResponse {
	var finishedAt *string
	if cmd.FinishedAt != nil {
		formatted := cmd.FinishedAt.Format(time.RFC3339Nano)
		finishedAt = &formatted
	}
	return CommandStatusResponse{
		ID:              cmd.ID,
		CommandType:     cmd.CommandType,
		DrillInstanceID: cmd.DrillInstanceID,
		StepInstanceID:  cmd.StepInstanceID,
		Status:          cmd.Status,
		Result:          cmd.Result,
		ErrorCode:       cmd.ErrorCode,
		ErrorMessage:    cmd.ErrorMessage,
		CreatedAt:       cmd.CreatedAt.Format(time.RFC3339Nano),
		FinishedAt:      finishedAt,
	}
}
