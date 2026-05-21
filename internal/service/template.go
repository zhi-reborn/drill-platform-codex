package service

import (
	"encoding/json"
	"sort"

	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
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
			ParentStepID:        stepReq.ParentStepID,
			StepType:            stepReq.StepType,
			TimeoutMinutes:      stepReq.TimeoutMinutes,
			GuideContent:        stepReq.GuideContent,
			IsBlocking:          stepReq.IsBlocking,
			DefaultAssigneeRole: stepReq.DefaultAssigneeRole,
			ExecutorTeam:        stepReq.ExecutorTeam,
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

func (s *TemplateService) ToggleStatus(id uint64) error {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 切换状态：0=禁用，2=启用
	if template.Status == 2 {
		template.Status = 0
	} else {
		template.Status = 2
	}
	return s.templateRepo.Update(template)
}

func (s *TemplateService) UpdateSteps(id uint64, steps []dto.StepTemplateRequest) error {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Auto-chain PreStepIDs: root-level steps without explicit PreStepIDs
	// are linked in seq order within the same parent group.
	s.chainPreStepIDs(steps)

	newSteps := make([]entity.StepTemplate, 0, len(steps))
	for _, stepReq := range steps {
		est := 5
		if stepReq.EstimatedDurationMinutes != nil && *stepReq.EstimatedDurationMinutes > 0 {
			est = *stepReq.EstimatedDurationMinutes
		}
		if est < 5 {
			est = 5
		}
		step := entity.StepTemplate{
			Name:                     stepReq.Name,
			Seq:                      stepReq.Seq,
			ParentStepID:             stepReq.ParentStepID,
			StepType:                 stepReq.StepType,
			TimeoutMinutes:           est * 2,
			GuideContent:             stepReq.GuideContent,
			IsBlocking:               stepReq.IsBlocking,
			DefaultAssigneeRole:      stepReq.DefaultAssigneeRole,
			ExecutorTeam:             stepReq.ExecutorTeam,
			Phase:                    stepReq.Phase,
			PhaseStep:                stepReq.PhaseStep,
			ExecutionMode:            stepReq.ExecutionMode,
			EstimatedDurationMinutes: stepReq.EstimatedDurationMinutes,
			EstimatedStartOffset:     stepReq.EstimatedStartOffset,
			TaskName:                 stepReq.TaskName,
			SubTask:                  stepReq.SubTask,
			ResponsibleDepartment:    stepReq.ResponsibleDepartment,
			ResponsiblePerson:        stepReq.ResponsiblePerson,
			Executor:                 stepReq.Executor,
			Reviewer:                 stepReq.Reviewer,
		}
		if stepReq.ID != nil {
			step.ID = *stepReq.ID
		}
		if len(stepReq.PreStepIDs) > 0 {
			b, _ := json.Marshal(stepReq.PreStepIDs)
			step.PreStepIDs = string(b)
		}
		newSteps = append(newSteps, step)
	}

	return s.templateRepo.UpdateSteps(template.ID, newSteps)
}

func (s *TemplateService) UpdateStep(templateID uint64, stepID uint64, stepReq dto.StepTemplateRequest) error {
	template, err := s.templateRepo.FindByID(templateID)
	if err != nil {
		return err
	}

	est := 5
	if stepReq.EstimatedDurationMinutes != nil && *stepReq.EstimatedDurationMinutes > 0 {
		est = *stepReq.EstimatedDurationMinutes
	}
	if est < 5 {
		est = 5
	}
	step := entity.StepTemplate{
		ID:                       stepID,
		Name:                     stepReq.Name,
		Seq:                      stepReq.Seq,
		ParentStepID:             stepReq.ParentStepID,
		StepType:                 stepReq.StepType,
		TimeoutMinutes:           est * 2,
		GuideContent:             stepReq.GuideContent,
		IsBlocking:               stepReq.IsBlocking,
		DefaultAssigneeRole:      stepReq.DefaultAssigneeRole,
		ExecutorTeam:             stepReq.ExecutorTeam,
		Phase:                    stepReq.Phase,
		PhaseStep:                stepReq.PhaseStep,
		ExecutionMode:            stepReq.ExecutionMode,
		EstimatedDurationMinutes: stepReq.EstimatedDurationMinutes,
		EstimatedStartOffset:     stepReq.EstimatedStartOffset,
		TaskName:                 stepReq.TaskName,
		SubTask:                  stepReq.SubTask,
		ResponsibleDepartment:    stepReq.ResponsibleDepartment,
		ResponsiblePerson:        stepReq.ResponsiblePerson,
		Executor:                 stepReq.Executor,
		Reviewer:                 stepReq.Reviewer,
	}
	step.DrillTemplateID = template.ID

	return s.templateRepo.UpdateStep(&step)
}

func (s *TemplateService) chainPreStepIDs(steps []dto.StepTemplateRequest) {
	type groupKey struct {
		parentID uint64
		phase    string
	}
	groups := make(map[groupKey][]*dto.StepTemplateRequest)
	for i := range steps {
		pid := uint64(0)
		if steps[i].ParentStepID != nil {
			pid = *steps[i].ParentStepID
		}
		k := groupKey{parentID: pid, phase: steps[i].Phase}
		groups[k] = append(groups[k], &steps[i])
	}

	for _, g := range groups {
		sort.Slice(g, func(i, j int) bool { return g[i].Seq < g[j].Seq })
		for i := 1; i < len(g); i++ {
			if len(g[i].PreStepIDs) == 0 {
				prev := g[i-1]
				if prev.ID != nil {
					g[i].PreStepIDs = []int64{int64(*prev.ID)}
				}
			}
		}
		if len(g) > 0 {
			first := g[0]
			if first.ParentStepID != nil && *first.ParentStepID != 0 && len(first.PreStepIDs) == 0 {
				first.PreStepIDs = []int64{int64(*first.ParentStepID)}
			}
		}
	}
}
