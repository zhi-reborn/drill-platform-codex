package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"drill-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxUserID   = "user_id"
	CtxUsername = "username"
	CtxRole     = "role"
	CtxUserIDInt = "user_id_int"
)

type JWTConfig struct {
	Secret string
}

// extractAndValidateToken 从请求中提取并验证 JWT token
// 优先从 Authorization header 读取，WS 场景下回退到 ?token= 查询参数
func extractAndValidateToken(c *gin.Context, cfg JWTConfig) (*jwt.Token, error) {
	tokenString := ""

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	}

	// WebSocket 场景：前端无法设置 Authorization header，通过 ?token= 传递
	if tokenString == "" {
		tokenString = c.Query("token")
	}

	if tokenString == "" {
		return nil, fmt.Errorf("未提供认证令牌")
	}

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 必须验证签名算法，防止 alg:none 攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})
}

// setUserContext 将用户信息注入 gin.Context
func setUserContext(c *gin.Context, claims jwt.MapClaims) error {
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return fmt.Errorf("user_id 类型断言失败")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return fmt.Errorf("username 类型断言失败")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return fmt.Errorf("role 类型断言失败")
	}

	userID := uint64(userIDFloat)

	c.Set(CtxUserID, strconv.FormatUint(userID, 10))
	c.Set(CtxUserIDInt, userID)
	c.Set(CtxUsername, username)
	c.Set(CtxRole, role)
	return nil
}

func JWTAuth(cfg JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractAndValidateToken(c, cfg)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "认证令牌解析失败")
			c.Abort()
			return
		}

		if err := setUserContext(c, claims); err != nil {
			response.Unauthorized(c, "认证令牌数据格式错误")
			c.Abort()
			return
		}

		c.Next()
	}
}

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
