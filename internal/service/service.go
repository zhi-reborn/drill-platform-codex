package service

import "drill-platform/internal/repository"

type Services struct {
	AuthService    *AuthService
	TemplateService *TemplateService
	DrillService   *DrillService
	TaskService    *TaskService
	DisplayService *DisplayService
	ReportService  *ReportService
}

func NewServices() *Services {
	userRepo := repository.NewUserRepo()
	templateRepo := repository.NewTemplateRepo()
	drillRepo := repository.NewDrillRepo()
	stepRepo := repository.NewStepRepo()

	return &Services{
		AuthService:    NewAuthService(userRepo),
		TemplateService: NewTemplateService(templateRepo),
		DrillService:   NewDrillService(drillRepo, templateRepo, stepRepo),
		TaskService:    NewTaskService(stepRepo),
		DisplayService: NewDisplayService(drillRepo, stepRepo),
		ReportService:  NewReportService(drillRepo, stepRepo),
	}
}
