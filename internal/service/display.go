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
	// 用步骤实时状态重算完成率，与大屏 ScreenView 的 progressPercent 同源
	drill.ProgressPct = ComputeProgressPct(drill.Steps)
	return drill, nil
}
