package auth

import (
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService *service.AuthService
}

func NewHandler(authService *service.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	res, err := h.authService.Login(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, res)
}

func (h *Handler) Logout(c *gin.Context) {
	response.Success(c, nil)
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	response.Success(c, gin.H{
		"user_id":  middleware.GetUserID(c),
		"username": middleware.GetUsername(c),
		"role":     middleware.GetRole(c),
	})
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, users)
}
