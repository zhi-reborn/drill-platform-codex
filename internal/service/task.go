package service

import (
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/repository"
)

type TaskService struct {
	stepRepo              *repository.StepRepo
	wsManager             *websocket.Manager
	notificationService   *NotificationService
}

func NewTaskService(stepRepo *repository.StepRepo) *TaskService {
	return &TaskService{stepRepo: stepRepo}
}

func (s *TaskService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

func (s *TaskService) SetNotificationService(ns *NotificationService) {
	s.notificationService = ns
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

	prevStatus := step.Status
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

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.StepChangePayload{
			DrillID:        uint(step.DrillInstanceID),
			StepID:         uint(stepID),
			StepName:       step.Name,
			PreviousStatus: prevStatus,
			NewStatus:      "completed",
		}
		s.wsManager.SendStepChange(uint(step.DrillInstanceID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   operatorID,
			Type:     "step_complete",
			Title:    "步骤已完成",
			Content:  step.Name + " 已完成",
			DrillID:  &step.DrillInstanceID,
			StepID:   &stepID,
			IsRead:   false,
		})
	}

	return nil
}

func (s *TaskService) ReportIssue(stepID uint64, operatorID uint64, issueDesc string) error {
	step, err := s.stepRepo.FindByID(stepID)
	if err != nil {
		return err
	}

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

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.StepChangePayload{
			DrillID:        uint(step.DrillInstanceID),
			StepID:         uint(stepID),
			StepName:       step.Name,
			PreviousStatus: step.Status,
			NewStatus:      "issue",
		}
		s.wsManager.SendStepChange(uint(step.DrillInstanceID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   step.DrillInstance.CreatedBy,
			Type:     "step_timeout",
			Title:    "步骤异常上报",
			Content:  step.Name + " 上报异常：" + issueDesc,
			DrillID:  &step.DrillInstanceID,
			StepID:   &stepID,
			IsRead:   false,
		})
	}

	return nil
}
