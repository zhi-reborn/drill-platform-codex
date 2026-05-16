package service

import (
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
	"errors"
)

type TemplateService struct {
	templateRepo *repository.TemplateRepo
}

func NewTemplateService(templateRepo *repository.TemplateRepo) *TemplateService {
	return &TemplateService{templateRepo: templateRepo}
}

func (s *TemplateService) GetList(page, pageSize int, category string) ([]entity.DrillTemplate, int64, error) {
	return s.templateRepo.List(page, pageSize, category)
}

func (s *TemplateService) GetDetail(id uint64) (*entity.DrillTemplate, error) {
	return s.templateRepo.FindByID(id)
}

func (s *TemplateService) Create(req *dto.CreateTemplateRequest, createdBy uint64) (*entity.DrillTemplate, error) {
	template := &entity.DrillTemplate{
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		CreatedBy:   createdBy,
		Status:      1,
	}

	for _, stepReq := range req.Steps {
		step := entity.StepTemplate{
			Name:                stepReq.Name,
			Seq:                 stepReq.Seq,
			StepType:            stepReq.StepType,
			TimeoutMinutes:      stepReq.TimeoutMinutes,
			GuideContent:        stepReq.GuideContent,
			IsBlocking:          stepReq.IsBlocking,
			DefaultAssigneeRole: stepReq.DefaultAssigneeRole,
		}
		if len(stepReq.PreStepIDs) > 0 {
			step.PreStepIDs = "[]"
		}
		template.Steps = append(template.Steps, step)
	}

	return template, s.templateRepo.Create(template)
}

func (s *TemplateService) Update(id uint64, req *dto.UpdateTemplateRequest) error {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return err
	}

	template.Name = req.Name
	template.Category = req.Category
	template.Description = req.Description

	return s.templateRepo.Update(template)
}

func (s *TemplateService) Delete(id uint64) error {
	return s.templateRepo.Delete(id)
}

func (s *TemplateService) Clone(id uint64) (*entity.DrillTemplate, error) {
	return s.templateRepo.Clone(id)
}

func (s *TemplateService) GetCategories() ([]entity.TemplateCategory, error) {
	return s.templateRepo.GetCategories()
}

func (s *TemplateService) SaveCategories(categories []entity.TemplateCategory) error {
	return s.templateRepo.SaveCategories(categories)
}

func (s *TemplateService) Publish(id uint64) error {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return err
	}

	if len(template.Steps) == 0 {
		return errors.New("模板必须包含至少一个步骤才能发布")
	}

	template.Status = 2
	return s.templateRepo.Update(template)
}

func (s *TemplateService) UpdateSteps(id uint64, steps []dto.StepTemplateRequest) error {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return err
	}

	newSteps := make([]entity.StepTemplate, 0, len(steps))
	for i, stepReq := range steps {
		step := entity.StepTemplate{
			Name:                stepReq.Name,
			Seq:                 i + 1,
			StepType:            stepReq.StepType,
			TimeoutMinutes:      stepReq.TimeoutMinutes,
			GuideContent:        stepReq.GuideContent,
			IsBlocking:          stepReq.IsBlocking,
			DefaultAssigneeRole: stepReq.DefaultAssigneeRole,
		}
		newSteps = append(newSteps, step)
	}

	return s.templateRepo.UpdateSteps(template.ID, newSteps)
}
