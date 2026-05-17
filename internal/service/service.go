package service

import (
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/repository"
)

type Services struct {
	AuthService       *AuthService
	TemplateService   *TemplateService
	DrillService      *DrillService
	TaskService       *TaskService
	DisplayService    *DisplayService
	ReportService     *ReportService
	NotificationService *NotificationService
	wsManager         *websocket.Manager
}

func NewServices(wsManager *websocket.Manager) *Services {
	userRepo := repository.NewUserRepo()
	templateRepo := repository.NewTemplateRepo()
	drillRepo := repository.NewDrillRepo()
	stepRepo := repository.NewStepRepo()
	notificationRepo := repository.NewNotificationRepo()

	s := &Services{
		AuthService:       NewAuthService(userRepo),
		TemplateService:   NewTemplateService(templateRepo),
		DrillService:      NewDrillService(drillRepo, templateRepo, stepRepo, userRepo),
		TaskService:       NewTaskService(stepRepo),
		DisplayService:    NewDisplayService(drillRepo, stepRepo),
		ReportService:     NewReportService(drillRepo, stepRepo),
		NotificationService: NewNotificationService(notificationRepo),
		wsManager:         wsManager,
	}

	// 注入 WebSocket Manager 到需要广播的服务
	s.DrillService.SetWebSocketManager(wsManager)
	s.TaskService.SetWebSocketManager(wsManager)
	s.DrillService.SetNotificationService(s.NotificationService)
	s.TaskService.SetNotificationService(s.NotificationService)

	return s
}
