package router

import (
	"drill-platform/internal/api/handler/auth"
	"drill-platform/internal/api/handler/display"
	"drill-platform/internal/api/handler/drill"
	"drill-platform/internal/api/handler/notification"
	"drill-platform/internal/api/handler/report"
	"drill-platform/internal/api/handler/task"
	"drill-platform/internal/api/handler/template"
	"drill-platform/internal/api/middleware"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(services *service.Services, wsManager *websocket.Manager, jwtSecret string) *gin.Engine {
	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	authHandler := auth.NewHandler(services.AuthService)
	templateHandler := template.NewHandler(services.TemplateService)
	drillHandler := drill.NewHandler(services.DrillService)
	taskHandler := task.NewHandler(services.TaskService)
	displayHandler := display.NewHandler(services.DisplayService)
	reportHandler := report.NewHandler(services.ReportService)
	notificationHandler := notification.NewHandler(services.NotificationService)

	jwtAuth := middleware.JWTAuth(middleware.JWTConfig{Secret: jwtSecret})

v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authHandler.Login)
		v1.GET("/auth/dev-users", authHandler.ListUsers)

		v1.Use(jwtAuth)
		{
			v1.GET("/auth/me", authHandler.GetCurrentUser)

			v1.GET("/users", authHandler.ListUsers)
			v1.GET("/users/:id", authHandler.GetUser)
			v1.POST("/users", middleware.RequireAdmin(), authHandler.CreateUser)
			v1.PUT("/users/:id", middleware.RequireAdmin(), authHandler.UpdateUser)
			v1.DELETE("/users/:id", middleware.RequireAdmin(), authHandler.DeleteUser)

			v1.GET("/templates", templateHandler.List)
			v1.GET("/templates/:id", templateHandler.GetDetail)
			v1.POST("/templates", middleware.RequireAdmin(), templateHandler.Create)
			v1.PUT("/templates/:id", middleware.RequireAdmin(), templateHandler.Update)
			v1.DELETE("/templates/:id", middleware.RequireAdmin(), templateHandler.Delete)
			v1.POST("/templates/:id/clone", middleware.RequireAdmin(), templateHandler.Clone)

			v1.GET("/drills", drillHandler.List)
			v1.GET("/drills/:id", drillHandler.GetDetail)
			v1.POST("/drills", middleware.RequireDirectorOrAbove(), drillHandler.Create)
			v1.POST("/drills/:id/start", middleware.RequireDirectorOrAbove(), drillHandler.Start)
			v1.POST("/drills/:id/pause", middleware.RequireDirectorOrAbove(), drillHandler.Pause)
			v1.POST("/drills/:id/resume", middleware.RequireDirectorOrAbove(), drillHandler.Resume)
			v1.POST("/drills/:id/terminate", middleware.RequireDirectorOrAbove(), drillHandler.Terminate)

			v1.GET("/tasks/my", taskHandler.GetMyTasks)
			v1.GET("/tasks/:stepId", taskHandler.GetDetail)
			v1.POST("/tasks/:stepId/complete", taskHandler.CompleteStep)
			v1.POST("/tasks/:stepId/issue", taskHandler.ReportIssue)

			v1.GET("/display/:drillId", displayHandler.GetDrillData)

			v1.GET("/drills/:id/report", reportHandler.GetReport)
			v1.POST("/drills/:id/report/export", reportHandler.ExportPDF)

			v1.GET("/notifications", notificationHandler.List)
			v1.POST("/notifications/:id/read", notificationHandler.MarkAsRead)
			v1.POST("/notifications/read-all", notificationHandler.MarkAllAsRead)
			v1.DELETE("/notifications/:id", notificationHandler.Delete)
		}
	}

	ws := r.Group("/ws")
	{
		ws.GET("/display/:drillId", wsManager.HandleDisplay)
		ws.GET("/tasks", wsManager.HandleTasks)
		ws.GET("/control/:drillId", wsManager.HandleControl)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
