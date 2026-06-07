package service

import (
	"errors"
	"fmt"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/repository"
)

type TaskService struct {
	stepRepo              *repository.StepRepo
	drillService          *DrillService
	wsManager             *websocket.Manager
	notificationService   *NotificationService
	userRepo              *repository.UserRepo
	redis                 RedisClient
}

func NewTaskService(stepRepo *repository.StepRepo) *TaskService {
	return &TaskService{stepRepo: stepRepo}
}

func (s *TaskService) SetUserRepo(repo *repository.UserRepo) {
	s.userRepo = repo
}

func (s *TaskService) SetDrillService(ds *DrillService) {
	s.drillService = ds
}

func (s *TaskService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

func (s *TaskService) SetNotificationService(ns *NotificationService) {
	s.notificationService = ns
}

func (s *TaskService) SetRedis(redis RedisClient) {
	s.redis = redis
}

func (s *TaskService) checkExecutorPermission(stepID uint64, userID uint64) (*entity.StepInstance, error) {
	step, err := s.stepRepo.FindByID(stepID)
	if err != nil {
		return nil, errors.New("任务不存在")
	}

	if s.userRepo == nil {
		return step, nil
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if user.Role == "admin" || user.Role == "director" {
		return step, nil
	}

	if step.ExecutorTeam != "" && user.Department != step.ExecutorTeam {
		return nil, fmt.Errorf("您所在部门 [%s] 不在该任务的执行组 [%s] 内，无法操作", user.Department, step.ExecutorTeam)
	}

	return step, nil
}

func (s *TaskService) GetMyTasks(userID uint64) ([]entity.StepInstance, error) {
	var steps []entity.StepInstance
	err := repository.DB.
		Where("JSON_CONTAINS(assignee_ids, CAST(? AS JSON)) AND status = ?",
			userID, "running").
		Find(&steps).Error
	return steps, err
}

func (s *TaskService) EnrichStepsWithAssigneeNames(steps []entity.StepInstance) []entity.StepInstance {
	if s.drillService != nil {
		return s.drillService.EnrichStepsWithAssigneeNames(steps)
	}
	return steps
}

func (s *TaskService) GetTaskDetail(stepID uint64) (*entity.StepInstance, error) {
	return s.stepRepo.FindByID(stepID)
}

func (s *TaskService) CompleteStep(stepID uint64, operatorID uint64, remark string) error {
	step, err := s.checkExecutorPermission(stepID, operatorID)
	if err != nil {
		return err
	}
	if step.Status != "running" {
		return nil
	}

	if s.drillService != nil && s.drillService.engine != nil {
		return s.drillService.engine.CompleteStep(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
			int64(operatorID),
			remark,
		)
	}

	now := time.Now()
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepID).Updates(map[string]interface{}{
		"status":         "completed",
		"actual_operator": operatorID,
		"end_time":       &now,
		"remark":         remark,
	})

	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", operatorID).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: step.DrillInstanceID,
		StepInstanceID:  &stepID,
		Action:          "complete",
		OperatorID:      operatorID,
		OperatorName:    operatorName,
		Content:         remark,
	})

	if s.wsManager != nil {
		startTimeStr := ""
		if step.StartTime != nil {
			startTimeStr = step.StartTime.Format(time.RFC3339)
		}
		endTimeStr := now.Format(time.RFC3339)
		s.wsManager.SendStepChange(uint(step.DrillInstanceID), websocket.StepChangePayload{
			DrillID:        uint(step.DrillInstanceID),
			StepID:         uint(stepID),
			StepName:       step.Name,
			PreviousStatus: "running",
			NewStatus:      "completed",
			Executor:       operatorName,
			StartTime:      &startTimeStr,
			EndTime:        &endTimeStr,
			Remark:         remark,
			AssigneeNames:  step.AssigneeNames,
		})
		PatchCachedStep(s.redis, step.DrillInstanceID, uint(stepID), map[string]interface{}{
			"status":         "completed",
			"start_time":     &startTimeStr,
			"end_time":       &endTimeStr,
			"remark":         remark,
			"assignee_names": step.AssigneeNames,
		})
	}

	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   operatorID,
			Type:     entity.NotificationTypeStepComplete,
			Title:    "步骤已完成",
			Content:  step.Name + " 已完成",
			DrillID:  &step.DrillInstanceID,
			StepID:   &stepID,
			IsRead:   false,
		}, operatorID)
	}

	return nil
}

func (s *TaskService) ReportIssue(stepID uint64, operatorID uint64, issueDesc string) error {
	step, err := s.checkExecutorPermission(stepID, operatorID)
	if err != nil {
		return err
	}

	if s.drillService != nil && s.drillService.engine != nil {
		return s.drillService.engine.ReportIssue(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
			int64(operatorID),
			issueDesc,
		)
	}

	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepID).Updates(map[string]interface{}{
		"status":     "issue",
		"issue_desc": issueDesc,
	})

	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", operatorID).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: step.DrillInstanceID,
		StepInstanceID:  &stepID,
		Action:          "issue",
		OperatorID:      operatorID,
		OperatorName:    operatorName,
		Content:         issueDesc,
	})

	if s.wsManager != nil {
		startTimeStr := ""
		if step.StartTime != nil {
			startTimeStr = step.StartTime.Format(time.RFC3339)
		}
		nowStr := time.Now().Format(time.RFC3339)
		s.wsManager.SendStepChange(uint(step.DrillInstanceID), websocket.StepChangePayload{
			DrillID:        uint(step.DrillInstanceID),
			StepID:         uint(stepID),
			StepName:       step.Name,
			PreviousStatus: step.Status,
			NewStatus:      "issue",
			Executor:       operatorName,
			StartTime:      &startTimeStr,
			EndTime:        &nowStr,
			IssueDesc:      issueDesc,
			AssigneeNames:  step.AssigneeNames,
		})
		PatchCachedStep(s.redis, step.DrillInstanceID, uint(stepID), map[string]interface{}{
			"status":         "issue",
			"start_time":     &startTimeStr,
			"end_time":       &nowStr,
			"issue_desc":     issueDesc,
			"assignee_names": step.AssigneeNames,
		})
	}

	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   step.DrillInstance.CreatedBy,
			Type:     entity.NotificationTypeStepTimeout,
			Title:    "步骤异常上报",
			Content:  step.Name + " 上报异常：" + issueDesc,
			DrillID:  &step.DrillInstanceID,
			StepID:   &stepID,
			IsRead:   false,
		})
	}

	return nil
}
