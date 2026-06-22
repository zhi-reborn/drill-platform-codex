package service

import "drill-platform/internal/repository"

type DisplayService struct {
	drillRepo *repository.DrillRepo
	stepRepo  *repository.StepRepo
}

func NewDisplayService(drillRepo *repository.DrillRepo, stepRepo *repository.StepRepo) *DisplayService {
	return &DisplayService{drillRepo: drillRepo, stepRepo: stepRepo}
}

func (s *DisplayService) GetDrillData(drillID uint64) (interface{}, error) {
	drill, err := s.drillRepo.FindByIDWithSteps(drillID)
	if err != nil {
		return nil, err
	}
	return drill, nil
}
