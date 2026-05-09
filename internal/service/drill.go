package service

import (
	"encoding/json"
	"errors"
	"time"

	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
)

type DrillService struct {
	drillRepo    *repository.DrillRepo
	templateRepo *repository.TemplateRepo
	stepRepo     *repository.StepRepo
}

func NewDrillService(drillRepo *repository.DrillRepo, templateRepo *repository.TemplateRepo, stepRepo *repository.StepRepo) *DrillService {
	return &DrillService{
		drillRepo:    drillRepo,
		templateRepo: templateRepo,
		stepRepo:     stepRepo,
	}
}

func (s *DrillService) GetList(page, pageSize int, status string) ([]entity.DrillInstance, int64, error) {
	return s.drillRepo.List(page, pageSize, status)
}

func (s *DrillService) GetDetail(id uint64) (*entity.DrillInstance, error) {
	return s.drillRepo.FindByID(id)
}

func (s *DrillService) Create(req *dto.CreateDrillRequest, createdBy uint64) (*entity.DrillInstance, error) {
	template, err := s.templateRepo.FindByID(req.TemplateID)
	if err != nil {
		return nil, errors.New("模板不存在")
	}

	if template.Status != 1 {
		return nil, errors.New("模板已禁用，无法创建演练")
	}

	drill := &entity.DrillInstance{
		TemplateID: req.TemplateID,
		Name:       req.Name,
		Status:     "pending",
		CreatedBy:  createdBy,
	}

	if req.PlannedStart != "" {
		t, err := time.Parse(time.RFC3339, req.PlannedStart)
		if err == nil {
			drill.PlannedStart = &t
		}
	}

	if err := s.drillRepo.Create(drill); err != nil {
		return nil, err
	}

	for _, stepTpl := range template.Steps {
		assigneeIDs := "[]"
		if userIDs, ok := req.Assignees[stepTpl.ID]; ok {
			if bytes, _ := json.Marshal(userIDs); bytes != nil {
				assigneeIDs = string(bytes)
			}
		}

		step := entity.StepInstance{
			DrillInstanceID: drill.ID,
			StepTemplateID:  stepTpl.ID,
			Name:            stepTpl.Name,
			Seq:             stepTpl.Seq,
			Status:          "pending",
			AssigneeIDs:     assigneeIDs,
		}
		repository.DB.Create(&step)
	}

	return s.drillRepo.FindByID(drill.ID)
}

func (s *DrillService) Start(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}
	if drill.Status != "pending" {
		return errors.New("只有待启动状态的演练才能开始")
	}

	now := time.Now()
	drill.Status = "running"
	drill.StartTime = &now
	if len(drill.Steps) > 0 {
		drill.CurrentStepID = &drill.Steps[0].ID
	}

	return s.drillRepo.Update(drill)
}

func (s *DrillService) Pause(id uint64) error {
	return s.drillRepo.UpdateStatus(id, "paused")
}

func (s *DrillService) Resume(id uint64) error {
	return s.drillRepo.UpdateStatus(id, "running")
}

func (s *DrillService) Terminate(id uint64) error {
	now := time.Now()
	repository.DB.Model(&entity.DrillInstance{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    "terminated",
		"end_time":  &now,
	})
	return nil
}
