package middleware

import (
	"drill-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

var roleHierarchy = map[string]int{
	"viewer":   1,
	"executor": 2,
	"director": 3,
	"admin":    4,
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetRole(c)
		if userRole == "" {
			response.Forbidden(c, "无权访问")
			c.Abort()
			return
		}

		userLevel := roleHierarchy[userRole]
		for _, role := range roles {
			if roleLevel, ok := roleHierarchy[role]; ok && userLevel >= roleLevel {
				c.Next()
				return
			}
		}

		if userRole == "admin" {
			c.Next()
			return
		}

		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

func RequireDirectorOrAbove() gin.HandlerFunc {
	return RequireRole("director")
}
