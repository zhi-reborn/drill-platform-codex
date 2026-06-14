package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"
)

type TaskService struct {
	stepRepo            *repository.StepRepo
	drillService        *DrillService
	wsManager           *websocket.Manager
	notificationService *NotificationService
	userRepo            *repository.UserRepo
	redis               RedisClient
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

func assigneeIDsContain(assigneeIDs string, userID uint64) bool {
	ids, ok := parseAssigneeIDs(assigneeIDs)
	if !ok {
		return false
	}
	for _, id := range ids {
		if id == userID {
			return true
		}
	}
	return false
}

func parseAssigneeIDs(assigneeIDs string) ([]uint64, bool) {
	if assigneeIDs == "" || assigneeIDs == "[]" || assigneeIDs == "null" {
		return nil, false
	}
	var ids []uint64
	if err := json.Unmarshal([]byte(assigneeIDs), &ids); err != nil {
		return nil, false
	}
	return ids, len(ids) > 0
}

func implicitAssigneeMatches(step *entity.StepInstance, user *entity.User) bool {
	if step.ExecutorTeam != "" {
		return user.Department == step.ExecutorTeam
	}
	return false
}

func parseStepAttributes(attributes string) map[string]string {
	if attributes == "" || attributes == "{}" || attributes == "null" {
		return nil
	}
	values := map[string]string{}
	if err := json.Unmarshal([]byte(attributes), &values); err != nil {
		return nil
	}
	return values
}

func operatorAttributeMatches(step *entity.StepInstance, user *entity.User) (bool, bool) {
	attributes := parseStepAttributes(step.JSONAttributes)
	operator := attributes["operator"]
	if operator == "" {
		return false, false
	}
	return operator == user.RealName || operator == user.Username, true
}

func canExecutorOperateStep(step *entity.StepInstance, user *entity.User) bool {
	if step.ActualOperator != nil && *step.ActualOperator == user.ID {
		return true
	}
	if matched, hasOperator := operatorAttributeMatches(step, user); hasOperator {
		return matched
	}
	ids, hasExplicitAssignees := parseAssigneeIDs(step.AssigneeIDs)
	if hasExplicitAssignees {
		for _, id := range ids {
			if id == user.ID {
				return true
			}
		}
		return false
	}
	return implicitAssigneeMatches(step, user)
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

	if !canExecutorOperateStep(step, user) {
		return nil, fmt.Errorf("该任务未分配给您或您所在部门，无法操作")
	}

	return step, nil
}

func (s *TaskService) GetMyTasks(userID uint64) ([]entity.StepInstance, error) {
	var steps []entity.StepInstance
	var user *entity.User

	if s.userRepo != nil {
		u, err := s.userRepo.FindByID(userID)
		if err != nil {
			return nil, errors.New("用户不存在")
		}
		user = u
	}

	statuses := []string{"pending", "running", "issue", "completed", "skipped", "timeout"}
	err := repository.DB.
		Table("drill_instance_step").
		Select("drill_instance_step.*").
		Joins("JOIN drill_instance ON drill_instance.id = drill_instance_step.drill_instance_id").
		Where("drill_instance.status IN ?", []string{"running", "paused"}).
		Where("drill_instance_step.status IN ?", statuses).
		Order("drill_instance_step.drill_instance_id DESC, drill_instance_step.seq ASC").
		Find(&steps).Error
	if err != nil {
		return nil, err
	}

	// 只返回当前用户可操作的任务；流程排序由前端使用全量步骤骨架完成。
	filtered := steps[:0]
	for _, step := range steps {
		if assigneeIDsContain(step.AssigneeIDs, userID) || (user != nil && canExecutorOperateStep(&step, user)) {
			filtered = append(filtered, step)
		}
	}

	return filtered, nil
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

func (s *TaskService) StartStep(stepID uint64, operatorID uint64) error {
	step, err := s.checkExecutorPermission(stepID, operatorID)
	if err != nil {
		return err
	}
	if step.Status != "pending" {
		return fmt.Errorf("只有待执行任务可以开始")
	}

	var childCount int64
	if err := repository.DB.Model(&entity.StepInstance{}).
		Where("drill_instance_id = ? AND parent_step_id = ?", step.DrillInstanceID, step.ID).
		Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return fmt.Errorf("父任务由子任务推进，不能直接开始")
	}

	if s.drillService != nil && s.drillService.engine != nil {
		err := s.drillService.engine.ManualStartStep(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
		)
		if errors.Is(err, flowengine.ErrInstanceNotFound) {
			if recErr := s.drillService.Recover(step.DrillInstanceID); recErr != nil {
				return fmt.Errorf("恢复演练状态失败")
			}
			return s.drillService.engine.ManualStartStep(
				int64(step.DrillInstanceID),
				int64(step.StepTemplateID),
			)
		}
		return err
	}

	now := time.Now()
	return repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepID).Updates(map[string]interface{}{
		"status":     "running",
		"start_time": &now,
	}).Error
}

func (s *TaskService) CompleteStep(stepID uint64, operatorID uint64, remark string) error {
	step, err := s.checkExecutorPermission(stepID, operatorID)
	if err != nil {
		return err
	}
	if step.Status != "running" {
		return fmt.Errorf("只有执行中的任务可以完成")
	}

	if s.drillService != nil && s.drillService.engine != nil {
		err := s.drillService.engine.CompleteStep(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
			int64(operatorID),
			remark,
		)
		if errors.Is(err, flowengine.ErrInstanceNotFound) {
			if recErr := s.drillService.Recover(step.DrillInstanceID); recErr != nil {
				return fmt.Errorf("恢复演练状态失败")
			}
			return s.drillService.engine.CompleteStep(
				int64(step.DrillInstanceID),
				int64(step.StepTemplateID),
				int64(operatorID),
				remark,
			)
		}
		return err
	}

	now := time.Now()
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepID).Updates(map[string]interface{}{
		"status":          "completed",
		"actual_operator": operatorID,
		"end_time":        &now,
		"remark":          remark,
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
			UserID:  operatorID,
			Type:    entity.NotificationTypeStepComplete,
			Title:   "步骤已完成",
			Content: step.Name + " 已完成",
			DrillID: &step.DrillInstanceID,
			StepID:  &stepID,
			IsRead:  false,
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
		err := s.drillService.engine.ReportIssue(
			int64(step.DrillInstanceID),
			int64(step.StepTemplateID),
			int64(operatorID),
			issueDesc,
		)
		if errors.Is(err, flowengine.ErrInstanceNotFound) {
			if recErr := s.drillService.Recover(step.DrillInstanceID); recErr != nil {
				return fmt.Errorf("恢复演练状态失败")
			}
			return s.drillService.engine.ReportIssue(
				int64(step.DrillInstanceID),
				int64(step.StepTemplateID),
				int64(operatorID),
				issueDesc,
			)
		}
		return err
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
			UserID:  step.DrillInstance.CreatedBy,
			Type:    entity.NotificationTypeStepTimeout,
			Title:   "步骤异常上报",
			Content: step.Name + " 上报异常：" + issueDesc,
			DrillID: &step.DrillInstanceID,
			StepID:  &stepID,
			IsRead:  false,
		})
	}

	return nil
}
