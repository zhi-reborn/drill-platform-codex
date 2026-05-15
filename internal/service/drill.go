package service

import (
	"encoding/json"
	"errors"
	"time"

	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/repository"
)

type DrillService struct {
	drillRepo             *repository.DrillRepo
	templateRepo          *repository.TemplateRepo
	stepRepo              *repository.StepRepo
	wsManager             *websocket.Manager
	notificationService   *NotificationService
}

func NewDrillService(drillRepo *repository.DrillRepo, templateRepo *repository.TemplateRepo, stepRepo *repository.StepRepo) *DrillService {
	return &DrillService{
		drillRepo:    drillRepo,
		templateRepo: templateRepo,
		stepRepo:     stepRepo,
	}
}

func (s *DrillService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

func (s *DrillService) SetNotificationService(ns *NotificationService) {
	s.notificationService = ns
}

func (s *DrillService) GetList(page, pageSize int, status string) ([]entity.DrillInstance, int64, error) {
	return s.drillRepo.List(page, pageSize, status)
}

func (s *DrillService) GetDetail(id uint64) (*entity.DrillInstance, error) {
	return s.drillRepo.FindByID(id)
}

func (s *DrillService) GetSteps(id uint64) ([]entity.StepInstance, error) {
	return s.stepRepo.FindStepsByDrillID(id)
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

	if err := s.drillRepo.Update(drill); err != nil {
		return err
	}

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "start",
		OperatorID:      drill.CreatedBy,
		OperatorName:    operatorName,
		Content:         "演练已启动",
	})

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.DrillStatusPayload{
			DrillID:        uint(drill.ID),
			DrillName:      drill.Name,
			PreviousStatus: "pending",
			NewStatus:      "running",
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   drill.CreatedBy,
			Type:     "drill_started",
			Title:    "演练已启动",
			Content:  drill.Name + " 已开始执行",
			DrillID:  &drill.ID,
			IsRead:   false,
		})
	}

	return nil
}

func (s *DrillService) Pause(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}
	if drill.Status != "running" {
		return errors.New("只有运行中的演练才能暂停")
	}
	prevStatus := drill.Status
	
	if err := s.drillRepo.UpdateStatus(id, "paused"); err != nil {
		return err
	}

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "pause",
		OperatorID:      drill.CreatedBy,
		OperatorName:    operatorName,
		Content:         "演练已暂停",
	})

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.DrillStatusPayload{
			DrillID:        uint(drill.ID),
			DrillName:      drill.Name,
			PreviousStatus: prevStatus,
			NewStatus:      "paused",
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   drill.CreatedBy,
			Type:     "drill_paused",
			Title:    "演练已暂停",
			Content:  drill.Name + " 已暂停",
			DrillID:  &drill.ID,
			IsRead:   false,
		})
	}

	return nil
}

func (s *DrillService) Resume(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}
	if drill.Status != "paused" {
		return errors.New("只有已暂停的演练才能继续")
	}
	prevStatus := drill.Status
	
	if err := s.drillRepo.UpdateStatus(id, "running"); err != nil {
		return err
	}

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "resume",
		OperatorID:      drill.CreatedBy,
		OperatorName:    operatorName,
		Content:         "演练已恢复执行",
	})

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.DrillStatusPayload{
			DrillID:        uint(drill.ID),
			DrillName:      drill.Name,
			PreviousStatus: prevStatus,
			NewStatus:      "running",
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   drill.CreatedBy,
			Type:     "drill_resumed",
			Title:    "演练已恢复",
			Content:  drill.Name + " 已恢复执行",
			DrillID:  &drill.ID,
			IsRead:   false,
		})
	}

	return nil
}

func (s *DrillService) Terminate(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}
	if drill.Status == "completed" {
		return errors.New("已完成的演练不能终止")
	}
	if drill.Status == "terminated" {
		return errors.New("演练已终止，无法重复操作")
	}
	prevStatus := drill.Status
	
	now := time.Now()
	repository.DB.Model(&entity.DrillInstance{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    "terminated",
		"end_time":  &now,
	})

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: id,
		Action:          "terminate",
		OperatorID:      drill.CreatedBy,
		OperatorName:    operatorName,
		Content:         "演练已终止",
	})

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.DrillStatusPayload{
			DrillID:        uint(drill.ID),
			DrillName:      drill.Name,
			PreviousStatus: prevStatus,
			NewStatus:      "terminated",
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   drill.CreatedBy,
			Type:     "drill_terminated",
			Title:    "演练已终止",
			Content:  drill.Name + " 已终止",
			DrillID:  &drill.ID,
			IsRead:   false,
		})
	}

	return nil
}

func (s *DrillService) GetLogs(id uint64) ([]entity.DrillInstanceLog, error) {
	return s.drillRepo.GetLogs(id)
}
