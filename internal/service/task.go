package service

import (
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
)

type TaskService struct {
	stepRepo *repository.StepRepo
}

func NewTaskService(stepRepo *repository.StepRepo) *TaskService {
	return &TaskService{stepRepo: stepRepo}
}

func (s *TaskService) GetMyTasks(userID uint64) ([]entity.StepInstance, error) {
	var steps []entity.StepInstance
	err := repository.DB.
		Where("JSON_CONTAINS(assignee_ids, CAST(? AS JSON)) AND status IN (?, ?)",
			userID, "pending", "running").
		Find(&steps).Error
	return steps, err
}

func (s *TaskService) GetTaskDetail(stepID uint64) (*entity.StepInstance, error) {
	return s.stepRepo.FindByID(stepID)
}

func (s *TaskService) CompleteStep(stepID uint64, operatorID uint64, remark string) error {
	step, err := s.stepRepo.FindByID(stepID)
	if err != nil {
		return err
	}
	if step.Status != "running" {
		return nil
	}

	now := time.Now()
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepID).Updates(map[string]interface{}{
		"status":         "completed",
		"actual_operator": operatorID,
		"end_time":       &now,
		"remark":         remark,
	})

	repository.DB.Create(&entity.StepInstanceLog{
		StepInstanceID: stepID,
		Action:         "complete",
		OperatorID:     operatorID,
		Content:        remark,
	})

	return nil
}

func (s *TaskService) ReportIssue(stepID uint64, operatorID uint64, issueDesc string) error {
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepID).Updates(map[string]interface{}{
		"status":     "issue",
		"issue_desc": issueDesc,
	})

	repository.DB.Create(&entity.StepInstanceLog{
		StepInstanceID: stepID,
		Action:         "issue",
		OperatorID:     operatorID,
		Content:        issueDesc,
	})

	return nil
}
