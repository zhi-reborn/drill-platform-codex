package service

import (
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/pkg/flowengine"
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
	notificationService := NewNotificationService(notificationRepo)

	engine := flowengine.NewEngine()

	adapter := NewDrillFlowAdapter(
		templateRepo,
		drillRepo,
		stepRepo,
		notificationRepo,
		userRepo,
		wsManager,
		notificationService,
	)

	engine.SetCallbacks(adapter)
	engine.SetStepLoader(adapter)
	adapter.SetupEventSubscriptions(engine)

	s := &Services{
		AuthService:       NewAuthService(userRepo),
		TemplateService:   NewTemplateService(templateRepo),
		DrillService:      NewDrillService(drillRepo, templateRepo, stepRepo, userRepo),
		TaskService:       NewTaskService(stepRepo),
		DisplayService:    NewDisplayService(drillRepo, stepRepo),
		ReportService:     NewReportService(drillRepo, stepRepo),
		NotificationService: notificationService,
		wsManager:         wsManager,
	}

	s.DrillService.SetEngine(engine, adapter)
	s.TaskService.SetDrillService(s.DrillService)
	s.TaskService.SetUserRepo(userRepo)

	s.DrillService.SetWebSocketManager(wsManager)
	s.TaskService.SetWebSocketManager(wsManager)
	s.DrillService.SetNotificationService(s.NotificationService)
	s.TaskService.SetNotificationService(s.NotificationService)

	engine.GetTimeoutScheduler().Start()

	return s
}
