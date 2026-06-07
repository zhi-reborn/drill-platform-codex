package auth

import (
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/pkg/response"
	"drill-platform/internal/service"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
		switch err.Error() {
		case "用户名不存在":
			response.Unauthorized(c, "用户名不存在")
		case "密码错误":
			response.Unauthorized(c, "密码错误")
		case "账户已被禁用":
			response.Forbidden(c, "账户已被禁用")
		default:
			response.InternalError(c, "内部错误")
		}
		return
	}

	response.Success(c, res)
}

func (h *Handler) CASLogin(c *gin.Context) {
	redirect := c.Query("redirect")
	serviceURL := h.authService.CASServiceURL(buildCASCallbackServiceURL(c), redirect)
	loginURL, err := h.authService.BuildCASLoginURL(serviceURL)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	c.Redirect(http.StatusFound, loginURL)
}

func (h *Handler) CASCallback(c *gin.Context) {
	ticket := c.Query("ticket")
	if ticket == "" {
		response.BadRequest(c, "CAS ticket 不能为空")
		return
	}
	redirect := c.Query("redirect")
	serviceURL := h.authService.CASServiceURL(buildCASCallbackServiceURL(c), redirect)
	res, err := h.authService.LoginWithCASTicket(ticket, serviceURL)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	if redirect == "" {
		response.Success(c, res)
		return
	}

	u, err := url.Parse(redirect)
	if err != nil {
		response.BadRequest(c, "无效的回跳地址")
		return
	}
	q := u.Query()
	q.Set("token", res.Token)
	q.Set("user_id", strconv.FormatUint(res.UserID, 10))
	q.Set("username", res.Username)
	q.Set("real_name", res.RealName)
	q.Set("role", res.Role)
	q.Set("department", res.Department)
	u.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, u.String())
}

func (h *Handler) Logout(c *gin.Context) {
	response.Success(c, nil)
}

func buildCASCallbackServiceURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := c.Request.Host
	if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   "/api/v1/auth/cas/callback",
	}
	return u.String()
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
		response.InternalError(c, "内部错误")
		return
	}
	response.Success(c, gin.H{
		"items":     users,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	})
}

func (h *Handler) GetDepartments(c *gin.Context) {
	departments, err := h.authService.GetDepartments()
	if err != nil {
		response.InternalError(c, "内部错误")
		return
	}
	response.Success(c, departments)
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
		response.InternalError(c, "内部错误")
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
		response.InternalError(c, "内部错误")
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
		response.InternalError(c, "内部错误")
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
		response.InternalError(c, "内部错误")
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "密码不能为空且至少 6 个字符")
		return
	}

	if err := h.authService.ResetPassword(id, req.Password); err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c, "内部错误")
		return
	}

	response.Success(c, nil)
}
