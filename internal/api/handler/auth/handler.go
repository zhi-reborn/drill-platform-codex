package auth

import (
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"
	"strconv"

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
	var query dto.UserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	query.Normalize()

	users, total, err := h.authService.ListUsersPaginated(query.Page, query.PageSize, query.Role)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"items": users,
		"total": total,
		"page": query.Page,
		"page_size": query.PageSize,
	})
}

func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	user, err := h.authService.GetUserByID(id)
	if err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.authService.CreateUser(&req)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			response.BadRequest(c, "用户名已存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.authService.UpdateUser(id, &req)
	if err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.authService.DeleteUser(id); err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}
