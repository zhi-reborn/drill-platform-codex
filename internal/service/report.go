package service

import "drill-platform/internal/repository"

type ReportService struct {
	drillRepo *repository.DrillRepo
	stepRepo  *repository.StepRepo
}

func NewReportService(drillRepo *repository.DrillRepo, stepRepo *repository.StepRepo) *ReportService {
	return &ReportService{drillRepo: drillRepo, stepRepo: stepRepo}
}

func (s *ReportService) GetReport(drillID uint64) (interface{}, error) {
	drill, err := s.drillRepo.FindByID(drillID)
	if err != nil {
		return nil, err
	}
	return drill, nil
}

func (s *ReportService) ExportPDF(drillID uint64) ([]byte, error) {
	return nil, nil
}
