package middleware

import (
	"strconv"
	"strings"

	"drill-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxUserID  = "user_id"
	CtxUsername = "username"
	CtxRole    = "role"
)

type JWTConfig struct {
	Secret string
}

func JWTAuth(cfg JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})
		if err != nil || !token.Valid {
			response.Unauthorized(c, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "认证令牌解析失败")
			c.Abort()
			return
		}

		userID := uint64(claims["user_id"].(float64))
		username := claims["username"].(string)
		role := claims["role"].(string)

		c.Set(CtxUserID, strconv.FormatUint(userID, 10))
		c.Set(CtxUserIDInt, userID)
		c.Set(CtxUsername, username)
		c.Set(CtxRole, role)

		c.Next()
	}
}

const CtxUserIDInt = "user_id_int"

func GetUserID(c *gin.Context) uint64 {
	if v, ok := c.Get(CtxUserIDInt); ok {
		return v.(uint64)
	}
	if v := c.GetString(CtxUserID); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		return id
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	return c.GetString(CtxUsername)
}

func GetRole(c *gin.Context) string {
	return c.GetString(CtxRole)
}
