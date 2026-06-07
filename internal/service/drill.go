package service

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"

	"gorm.io/gorm"
)

type DrillService struct {
	drillRepo           *repository.DrillRepo
	templateRepo        *repository.TemplateRepo
	stepRepo            *repository.StepRepo
	userRepo            *repository.UserRepo
	engine              *flowengine.Engine
	adapter             *DrillFlowAdapter
	wsManager           *websocket.Manager
	notificationService *NotificationService
	redis               RedisClient
}

func NewDrillService(drillRepo *repository.DrillRepo, templateRepo *repository.TemplateRepo, stepRepo *repository.StepRepo, userRepo *repository.UserRepo) *DrillService {
	return &DrillService{
		drillRepo:    drillRepo,
		templateRepo: templateRepo,
		stepRepo:     stepRepo,
		userRepo:     userRepo,
	}
}

func (s *DrillService) SetRedis(redis RedisClient) {
	s.redis = redis
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

	steps, err := s.stepRepo.FindStepsByDrillID(id)
	if err != nil {
		return err
	}

	if err := s.backfillMissingStepTemplateIDs(drill.ID, template.Steps, steps); err != nil {
		return err
	}

	// 同步内存中步骤实例的 ID 到数据库 ID
	s.adapter.SyncStepInstanceIDs(int64(drill.ID))

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

func (s *DrillService) backfillMissingStepTemplateIDs(drillID uint64, templateSteps []entity.StepTemplate, steps []entity.StepInstance) error {
	bySeqName := make(map[string]uint64, len(templateSteps))
	bySeq := make(map[int]uint64, len(templateSteps))
	templateIDs := make(map[uint64]struct{}, len(templateSteps))
	for _, step := range templateSteps {
		bySeqName[stepTemplateKey(step.Seq, step.Name)] = step.ID
		if _, exists := bySeq[step.Seq]; !exists {
			bySeq[step.Seq] = step.ID
		}
		templateIDs[step.ID] = struct{}{}
	}

	for i := range steps {
		// 已存在且指向有效模板步骤则跳过；否则按 seq+name 重新匹配
		if _, ok := templateIDs[steps[i].StepTemplateID]; ok {
			continue
		}
		stepTemplateID, exists := bySeqName[stepTemplateKey(steps[i].Seq, steps[i].Name)]
		if !exists {
			stepTemplateID, exists = bySeq[steps[i].Seq]
		}
		if !exists {
			return errors.New("步骤模板映射不存在")
		}
		steps[i].StepTemplateID = stepTemplateID
		if err := repository.DB.Model(&entity.StepInstance{}).
			Where("drill_instance_id = ? AND id = ?", drillID, steps[i].ID).
			Update("template_step_id", stepTemplateID).Error; err != nil {
			return err
		}
	}

	return nil
}

func stepTemplateKey(seq int, name string) string {
	return strconv.Itoa(seq) + "\x00" + name
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

func (s *DrillService) GetUsersByIDs(ids []uint64) (map[uint64]*entity.User, error) {
	return s.userRepo.FindByIDs(ids)
}

func (s *DrillService) GetDetail(id uint64) (*entity.DrillInstance, error) {
	return s.drillRepo.FindByID(id)
}

func (s *DrillService) GetSteps(id uint64) ([]entity.StepInstance, error) {
	if steps, ok := GetCachedSteps(s.redis, id); ok {
		return steps, nil
	}
	steps, err := s.stepRepo.FindStepsByDrillID(id)
	if err != nil {
		return nil, err
	}
	SetCachedSteps(s.redis, id, steps)
	return steps, nil
}

func (s *DrillService) InvalidateStepCache(drillID uint64) {
	InvalidateStepCache(s.redis, drillID)
}

func (s *DrillService) EnrichStepsWithAssigneeNames(steps []entity.StepInstance) []entity.StepInstance {
	allIDs := make(map[uint64]bool)
	for i := range steps {
		if steps[i].AssigneeIDs == "" || steps[i].AssigneeIDs == "[]" {
			continue
		}
		var ids []uint64
		if json.Unmarshal([]byte(steps[i].AssigneeIDs), &ids) == nil {
			for _, id := range ids {
				allIDs[id] = true
			}
		}
	}
	if len(allIDs) == 0 {
		return steps
	}

	ids := make([]uint64, 0, len(allIDs))
	for id := range allIDs {
		ids = append(ids, id)
	}

	nameMap := GetCachedUserNames(s.redis, ids)
	if nameMap == nil {
		var users []entity.User
		repository.DB.Where("id IN ?", ids).Find(&users)
		nameMap = make(map[uint64]string, len(users))
		for _, u := range users {
			nameMap[u.ID] = u.RealName
		}
		SetCachedUserNames(s.redis, users)
	}

	for i := range steps {
		if steps[i].AssigneeIDs == "" || steps[i].AssigneeIDs == "[]" {
			continue
		}
		var ids []uint64
		if json.Unmarshal([]byte(steps[i].AssigneeIDs), &ids) == nil && len(ids) > 0 {
			var names []string
			for _, id := range ids {
				if n, ok := nameMap[id]; ok {
					names = append(names, n)
				}
			}
			steps[i].AssigneeNames = namesJoin(names)
		}
	}
	return steps
}

func namesJoin(names []string) string {
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}

func (s *DrillService) Create(req *dto.CreateDrillRequest, createdBy uint64) (*entity.DrillInstance, error) {
	template, err := s.templateRepo.FindByID(req.TemplateID)
	if err != nil {
		return nil, errors.New("模板不存在")
	}

	if template.Status == 0 {
		return nil, errors.New("模板已禁用，无法创建演练")
	}

	departments := make(map[string]bool)
	for _, stepTpl := range template.Steps {
		if _, ok := req.Assignees[stepTpl.ID]; ok {
			continue
		}
		if stepTpl.ExecutorTeam != "" {
			departments[stepTpl.ExecutorTeam] = true
		}
	}

	var deptUsers map[string][]entity.User
	if len(departments) > 0 {
		deptList := make([]string, 0, len(departments))
		for d := range departments {
			deptList = append(deptList, d)
		}
		deptUsers, _ = s.userRepo.FindByDepartments(deptList)
	}

	drill := &entity.DrillInstance{
		TemplateID:  req.TemplateID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "pending",
		CreatedBy:   createdBy,
	}

	if req.PlannedStart != "" {
		t, err := time.Parse(time.RFC3339, req.PlannedStart)
		if err == nil {
			drill.PlannedStart = &t
		}
	}

	err = repository.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(drill).Error; err != nil {
			return err
		}

		for _, stepTpl := range template.Steps {
			assigneeIDs := "[]"
			if userIDs, ok := req.Assignees[stepTpl.ID]; ok && len(userIDs) > 0 {
				if bytes, _ := json.Marshal(userIDs); bytes != nil {
					assigneeIDs = string(bytes)
				}
			} else if stepTpl.ExecutorTeam != "" {
				if users, ok := deptUsers[stepTpl.ExecutorTeam]; ok && len(users) > 0 {
					userIDs := make([]uint64, len(users))
					for i, u := range users {
						userIDs[i] = u.ID
					}
					if bytes, _ := json.Marshal(userIDs); bytes != nil {
						assigneeIDs = string(bytes)
					}
				}
			}

			step := entity.StepInstance{
				DrillInstanceID:          drill.ID,
				StepTemplateID:           stepTpl.ID,
				Name:                     stepTpl.Name,
				Seq:                      stepTpl.Seq,
				Status:                   "pending",
				AssigneeIDs:              assigneeIDs,
				StepType:                 stepTpl.StepType,
				TimeoutMinutes:           stepTpl.TimeoutMinutes,
				ParentStepID:             nil,
				PreStepIDs:               "[]",
				Phase:                    stepTpl.Phase,
				PhaseStep:                stepTpl.PhaseStep,
				DefaultAssigneeRole:      stepTpl.DefaultAssigneeRole,
				ExecutorTeam:             stepTpl.ExecutorTeam,
				EstimatedDurationMinutes: stepTpl.EstimatedDurationMinutes,
				EstimatedStartOffset:     stepTpl.EstimatedStartOffset,
				JSONAttributes:           stepTpl.JSONAttributes,
			}
			if err := tx.Create(&step).Error; err != nil {
				log.Printf("[ERROR] Failed to create step instance for template step %d: %v", stepTpl.ID, err)
				return err
			}
		}

		var instanceSteps []entity.StepInstance
		if err := tx.Where("drill_instance_id = ?", drill.ID).Order("seq ASC").Find(&instanceSteps).Error; err != nil {
			return err
		}

		tplIDtoInstID := make(map[uint64]uint64)
		for _, si := range instanceSteps {
			tplIDtoInstID[si.StepTemplateID] = si.ID
		}

		tplStepMap := make(map[uint64]entity.StepTemplate)
		for _, s := range template.Steps {
			tplStepMap[s.ID] = s
		}
		// 批量更新 parent_step_id，避免 N 次单条 UPDATE
		type parentUpdate struct {
			ID           uint64
			ParentStepID uint64
		}
		var parentUpdates []parentUpdate
		for i := range instanceSteps {
			tpl := tplStepMap[instanceSteps[i].StepTemplateID]
			if tpl.ParentStepID != nil && *tpl.ParentStepID > 0 {
				if instParentID, ok := tplIDtoInstID[*tpl.ParentStepID]; ok {
					instanceSteps[i].ParentStepID = &instParentID
					parentUpdates = append(parentUpdates, parentUpdate{
						ID:           instanceSteps[i].ID,
						ParentStepID: instParentID,
					})
				}
			}
		}
		if len(parentUpdates) > 0 {
			caseSQL := "UPDATE `drill_instance_step` SET `parent_step_id` = CASE `id` "
			var args []interface{}
			var allIDs []uint64
			for _, u := range parentUpdates {
				caseSQL += "WHEN ? THEN ? "
				args = append(args, u.ID, u.ParentStepID)
				allIDs = append(allIDs, u.ID)
			}
			caseSQL += "ELSE `parent_step_id` END WHERE `id` IN ?"
			args = append(args, allIDs)
			if err := tx.Exec(caseSQL, args...).Error; err != nil {
				return err
			}
		}

		s.computeInstancePreStepIDsTx(instanceSteps, template.Steps, tx)

		return nil
	})

	if err != nil {
		return nil, err
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

	s.syncPreStepIDsToEngine(int64(drill.ID))

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
			Operator:       operatorName,
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知（自己操作不通知自己）
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:  drill.CreatedBy,
			Type:    "drill_paused",
			Title:   "演练已暂停",
			Content: drill.Name + " 已暂停",
			DrillID: &drill.ID,
			IsRead:  false,
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
		Content:         "演练已恢复",
	})

	// WebSocket 广播
	if s.wsManager != nil {
		payload := websocket.DrillStatusPayload{
			DrillID:        uint(drill.ID),
			DrillName:      drill.Name,
			PreviousStatus: prevStatus,
			NewStatus:      "running",
			Operator:       operatorName,
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:  drill.CreatedBy,
			Type:    "drill_resumed",
			Title:   "演练已恢复",
			Content: drill.Name + " 已恢复执行",
			DrillID: &drill.ID,
			IsRead:  false,
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
		"status":   "terminated",
		"end_time": &now,
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
			Operator:       operatorName,
		}
		s.wsManager.SendDrillStatus(uint(drill.ID), payload)
	}

	// 创建通知
	if s.notificationService != nil {
		s.notificationService.CreateNotification(&entity.Notification{
			UserID:  drill.CreatedBy,
			Type:    "drill_terminated",
			Title:   "演练已终止",
			Content: drill.Name + " 已终止",
			DrillID: &drill.ID,
			IsRead:  false,
		}, drill.CreatedBy)
	}

	return nil
}

func (s *DrillService) GetLogs(id uint64) ([]entity.DrillInstanceLog, error) {
	return s.drillRepo.GetLogs(id, 0) // 0 → 默认 200 条
}

func (s *DrillService) Delete(id uint64) error {
	return s.drillRepo.Delete(id)
}
