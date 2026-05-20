package service

import (
	"encoding/json"
	"errors"
	"time"

	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"
)

type DrillService struct {
	drillRepo             *repository.DrillRepo
	templateRepo          *repository.TemplateRepo
	stepRepo              *repository.StepRepo
	userRepo              *repository.UserRepo
	engine                *flowengine.Engine
	adapter               *DrillFlowAdapter
	wsManager             *websocket.Manager
	notificationService   *NotificationService
}

func NewDrillService(drillRepo *repository.DrillRepo, templateRepo *repository.TemplateRepo, stepRepo *repository.StepRepo, userRepo *repository.UserRepo) *DrillService {
	return &DrillService{
		drillRepo:    drillRepo,
		templateRepo: templateRepo,
		stepRepo:     stepRepo,
		userRepo:     userRepo,
	}
}

func (s *DrillService) SetEngine(engine *flowengine.Engine, adapter *DrillFlowAdapter) {
	s.engine = engine
	s.adapter = adapter
}

func (s *DrillService) Engine() *flowengine.Engine {
	return s.engine
}

func (s *DrillService) Recover(id uint64) error {
	drill, err := s.drillRepo.FindByID(id)
	if err != nil {
		return err
	}

	template, err := s.templateRepo.FindByID(drill.TemplateID)
	if err != nil || template == nil {
		return errors.New("关联模板不存在")
	}

	flowDef := s.adapter.BuildFlowDef(template)
	flowDef.ID = int64(drill.ID)
	assignees := s.adapter.BuildAssignees(drill.ID)

	inst, err := s.engine.CreateInstance(flowDef, assignees, int64(drill.CreatedBy))
	if err != nil {
		return err
	}
	inst.Status = flowengine.FlowStatus(drill.Status)

	// 同步内存中步骤实例的 ID 到数据库 ID
	s.adapter.SyncStepInstanceIDs(int64(drill.ID))

	steps, err := s.stepRepo.FindStepsByDrillID(id)
	if err != nil {
		return err
	}

	for _, step := range steps {
		si, exists := inst.Steps[int64(step.StepTemplateID)]
		if exists {
			si.Status = flowengine.StepStatus(step.Status)
			si.StartTime = step.StartTime
			si.EndTime = step.EndTime
			si.TimeoutAt = step.TimeoutAt
			si.Remark = step.Remark
			si.IssueDesc = step.IssueDesc
			if step.ActualOperator != nil {
				op := int64(*step.ActualOperator)
				si.ActualOperator = &op
			}

			if step.Status == "running" && step.TimeoutAt != nil {
				s.engine.TimeoutScheduler().Register(int64(drill.ID), int64(step.StepTemplateID), int64(step.ID), *step.TimeoutAt)
			}
		}
	}

	return nil
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

func (s *DrillService) GetUserByID(id uint64) (*entity.User, error) {
	return s.userRepo.FindByID(id)
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

	if template.Status == 0 {
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
		// 优先使用手动指定的 assignees
		if userIDs, ok := req.Assignees[stepTpl.ID]; ok && len(userIDs) > 0 {
			if bytes, _ := json.Marshal(userIDs); bytes != nil {
				assigneeIDs = string(bytes)
			}
		} else if stepTpl.ExecutorTeam != "" {
			// 按执行组部门自动分配：查找该部门下所有活跃用户
			var deptUsers []entity.User
			repository.DB.Where("department = ? AND status = 1", stepTpl.ExecutorTeam).Find(&deptUsers)
			if len(deptUsers) > 0 {
				userIDs := make([]uint64, len(deptUsers))
				for i, u := range deptUsers {
					userIDs[i] = u.ID
				}
				if bytes, _ := json.Marshal(userIDs); bytes != nil {
					assigneeIDs = string(bytes)
				}
			}
		}

		step := entity.StepInstance{
			DrillInstanceID:     drill.ID,
			StepTemplateID:      stepTpl.ID,
			Name:                stepTpl.Name,
			Seq:                 stepTpl.Seq,
			Status:              "pending",
			AssigneeIDs:         assigneeIDs,
			StepType:            stepTpl.StepType,
			TimeoutMinutes:      stepTpl.TimeoutMinutes,
			DefaultAssigneeRole: stepTpl.DefaultAssigneeRole,
			ExecutorTeam:        stepTpl.ExecutorTeam,
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

	if s.engine == nil {
		return errors.New("流程引擎未初始化")
	}

	template, err := s.templateRepo.FindByID(drill.TemplateID)
	if err != nil || template == nil {
		return errors.New("关联模板不存在")
	}

	if len(template.Steps) == 0 {
		return errors.New("模板没有步骤，无法启动")
	}

	s.adapter.RegisterDrillContext(int64(drill.ID), drillContext{
		ID:         drill.ID,
		Name:       drill.Name,
		Status:     "running",
		TemplateID: drill.TemplateID,
	})

	flowDef := s.adapter.BuildFlowDef(template)
	flowDef.ID = int64(drill.ID)
	assignees := s.adapter.BuildAssignees(drill.ID)

	_, err = s.engine.CreateInstance(flowDef, assignees, int64(drill.CreatedBy))
	if err != nil {
		return err
	}

	// 同步内存中步骤实例的 ID 到数据库 ID
	s.adapter.SyncStepInstanceIDs(int64(drill.ID))

	// 查询操作人姓名
	var user entity.User
	operatorName := ""
	repository.DB.Model(&entity.User{}).Where("id = ?", drill.CreatedBy).First(&user)
	if user.ID > 0 {
		operatorName = user.RealName
	}

	// 创建演练日志
	s.drillRepo.CreateLog(&entity.DrillInstanceLog{
		DrillInstanceID: drill.ID,
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

	return s.engine.Start(int64(drill.ID))
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

	// 创建通知（自己操作不通知自己）
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:   drill.CreatedBy,
			Type:     "drill_paused",
			Title:    "演练已暂停",
			Content:  drill.Name + " 已暂停",
			DrillID:  &drill.ID,
			IsRead:   false,
		}, drill.CreatedBy)
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
		}, drill.CreatedBy)
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
		}, drill.CreatedBy)
	}

	return nil
}

func (s *DrillService) GetLogs(id uint64) ([]entity.DrillInstanceLog, error) {
	return s.drillRepo.GetLogs(id)
}

func (s *DrillService) Delete(id uint64) error {
	return s.drillRepo.Delete(id)
}
