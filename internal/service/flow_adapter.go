package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/infrastructure/websocket"
	"drill-platform/internal/pkg/flowengine"
	"drill-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

type drillContext struct {
	ID         uint64
	Name       string
	Status     string
	TemplateID uint64
}

type DrillFlowAdapter struct {
	templateRepo   *repository.TemplateRepo
	drillRepo      *repository.DrillRepo
	stepRepo       *repository.StepRepo
	notificationRepo *repository.NotificationRepo
	userRepo       *repository.UserRepo
	wsManager      *websocket.Manager
	notificationService *NotificationService
	engine         *flowengine.Engine

	mu       sync.RWMutex
	contexts map[int64]drillContext
}

func NewDrillFlowAdapter(
	templateRepo *repository.TemplateRepo,
	drillRepo *repository.DrillRepo,
	stepRepo *repository.StepRepo,
	notificationRepo *repository.NotificationRepo,
	userRepo *repository.UserRepo,
	wsManager *websocket.Manager,
	notificationService *NotificationService,
) *DrillFlowAdapter {
	return &DrillFlowAdapter{
		templateRepo:      templateRepo,
		drillRepo:         drillRepo,
		stepRepo:          stepRepo,
		notificationRepo:  notificationRepo,
		userRepo:          userRepo,
		wsManager:         wsManager,
		notificationService: notificationService,
		contexts:          make(map[int64]drillContext),
	}
}

func (a *DrillFlowAdapter) RegisterDrillContext(flowInstID int64, ctx drillContext) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.contexts[flowInstID] = ctx
}

func (a *DrillFlowAdapter) GetDrillContext(flowInstID int64) *drillContext {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if ctx, ok := a.contexts[flowInstID]; ok {
		return &ctx
	}
	return nil
}

func (a *DrillFlowAdapter) GetStepDef(flowDefID, stepDefID int64) (*flowengine.StepDef, error) {
	a.mu.RLock()
	ctx, ok := a.contexts[flowDefID]
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("drill context not found: flowDefID=%d", flowDefID)
	}

	template, err := a.templateRepo.FindByID(ctx.TemplateID)
	if err != nil || template == nil {
		return nil, fmt.Errorf("template not found: %d", ctx.TemplateID)
	}
	for _, st := range template.Steps {
		if int64(st.ID) == stepDefID {
			var preStepIDs []int64
			if st.PreStepIDs != "" && st.PreStepIDs != "null" {
				var uintIDs []uint64
				if err := json.Unmarshal([]byte(st.PreStepIDs), &uintIDs); err == nil {
					for _, uid := range uintIDs {
						preStepIDs = append(preStepIDs, int64(uid))
					}
				}
			}
			parentStepID := int64(0)
			if st.ParentStepID != nil {
				parentStepID = int64(*st.ParentStepID)
			}
			return &flowengine.StepDef{
				ID:                  int64(st.ID),
				Name:                st.Name,
				Seq:                 st.Seq,
				StepType:            flowengine.StepType(st.StepType),
				TimeoutMinutes:      st.TimeoutMinutes,
				PreStepIDs:          preStepIDs,
				GuideContent:        st.GuideContent,
				IsBlocking:          st.IsBlocking == 1,
				DefaultAssigneeRole: st.DefaultAssigneeRole,
				ParentStepID:        parentStepID,
			}, nil
		}
	}
	return nil, fmt.Errorf("step not found: flowDefID=%d, stepDefID=%d", flowDefID, stepDefID)
}

func (a *DrillFlowAdapter) GetAllStepDefs(flowDefID int64) ([]*flowengine.StepDef, error) {
	a.mu.RLock()
	ctx, ok := a.contexts[flowDefID]
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("drill context not found: flowDefID=%d", flowDefID)
	}

	template, err := a.templateRepo.FindByID(ctx.TemplateID)
	if err != nil || template == nil {
		return nil, fmt.Errorf("template not found: %d", ctx.TemplateID)
	}

	var defs []*flowengine.StepDef
	for _, st := range template.Steps {
		var preStepIDs []int64
		if st.PreStepIDs != "" && st.PreStepIDs != "null" {
			var uintIDs []uint64
			if err := json.Unmarshal([]byte(st.PreStepIDs), &uintIDs); err == nil {
				for _, uid := range uintIDs {
					preStepIDs = append(preStepIDs, int64(uid))
				}
			}
		}
		parentStepID := int64(0)
		if st.ParentStepID != nil {
			parentStepID = int64(*st.ParentStepID)
		}
		defs = append(defs, &flowengine.StepDef{
			ID:                  int64(st.ID),
			Name:                st.Name,
			Seq:                 st.Seq,
			StepType:            flowengine.StepType(st.StepType),
			TimeoutMinutes:      st.TimeoutMinutes,
			PreStepIDs:          preStepIDs,
			GuideContent:        st.GuideContent,
			IsBlocking:          st.IsBlocking == 1,
			DefaultAssigneeRole: st.DefaultAssigneeRole,
			ParentStepID:        parentStepID,
		})
	}
	return defs, nil
}

func (a *DrillFlowAdapter) GetCurrentStepStatus(flowInstID, stepDefID int64) (*flowengine.StepInst, error) {
	var stepInst entity.StepInstance
	err := repository.DB.
		Where("drill_instance_id = ? AND step_template_id = ?", uint64(flowInstID), uint64(stepDefID)).
		First(&stepInst).Error
	if err != nil {
		return nil, err
	}

	var assigneeIDs []int64
	if stepInst.AssigneeIDs != "" && stepInst.AssigneeIDs != "[]" {
		var uintIDs []uint64
		if err := json.Unmarshal([]byte(stepInst.AssigneeIDs), &uintIDs); err == nil {
			for _, uid := range uintIDs {
				assigneeIDs = append(assigneeIDs, int64(uid))
			}
		}
	}

	return &flowengine.StepInst{
		StepDefID:   int64(stepInst.StepTemplateID),
		Name:        stepInst.Name,
		Seq:         stepInst.Seq,
		Status:      flowengine.StepStatus(stepInst.Status),
		AssigneeIDs: assigneeIDs,
		StartTime:   stepInst.StartTime,
		EndTime:     stepInst.EndTime,
		TimeoutAt:   stepInst.TimeoutAt,
	}, nil
}

func (a *DrillFlowAdapter) OnFlowStatusChanged(flowInstID int64, oldStatus, newStatus flowengine.FlowStatus) {
	drillID := uint64(flowInstID)

	updates := map[string]interface{}{"status": string(newStatus)}
	now := time.Now()

	switch newStatus {
	case flowengine.FlowStatusRunning:
		updates["start_time"] = now
	case flowengine.FlowStatusCompleted:
		updates["end_time"] = now
		updates["progress_pct"] = 100
	case flowengine.FlowStatusTerminated, flowengine.FlowStatusIssue:
		updates["end_time"] = now
	}

	repository.DB.Model(&entity.DrillInstance{}).Where("id = ?", drillID).Updates(updates)

	if newStatus == flowengine.FlowStatusCompleted {
		repository.DB.Create(&entity.DrillInstanceLog{
			DrillInstanceID: drillID,
			Action:          "complete",
			Content:         "演练已完成",
		})
	}
	if newStatus == flowengine.FlowStatusTerminated {
		repository.DB.Create(&entity.DrillInstanceLog{
			DrillInstanceID: drillID,
			Action:          "terminate",
			Content:         "演练已终止",
		})
	}

	ctx := a.GetDrillContext(flowInstID)
	drillName := ""
	if ctx != nil {
		drillName = ctx.Name
		a.contexts[flowInstID] = drillContext{ID: ctx.ID, Name: ctx.Name, Status: string(newStatus), TemplateID: ctx.TemplateID}
	}

	if a.wsManager != nil {
		a.wsManager.SendDrillStatus(uint(drillID), websocket.DrillStatusPayload{
			DrillID:        uint(drillID),
			DrillName:      drillName,
			PreviousStatus: string(oldStatus),
			NewStatus:      string(newStatus),
		})
	}

	if a.notificationService != nil {
		var title, content string
		switch newStatus {
		case flowengine.FlowStatusCompleted:
			title, content = "演练已完成", drillName+" 已完成所有步骤"
		case flowengine.FlowStatusTerminated:
			title, content = "演练已终止", drillName+" 已被终止"
		case flowengine.FlowStatusIssue:
			title, content = "演练异常", drillName+" 出现异常步骤"
		}
		if title != "" {
			a.notificationService.CreateNotification(&entity.Notification{
				UserID:   drillID,
				Type:     entity.NotificationType(title),
				Title:    title,
				Content:  content,
				DrillID:  &drillID,
				DrillName: &drillName,
				IsRead:   false,
			})
		}
	}
}

func (a *DrillFlowAdapter) OnStepStatusChanged(stepInstID int64, oldStatus, newStatus flowengine.StepStatus) {
}

func (a *DrillFlowAdapter) SyncStepInstanceIDs(flowInstID int64) {
	var steps []entity.StepInstance
	repository.DB.Where("drill_instance_id = ?", flowInstID).Find(&steps)

	inst, ok := a.engine.GetInstance(flowInstID)
	if !ok {
		return
	}

	for _, s := range steps {
		if si, exists := inst.Steps[int64(s.StepTemplateID)]; exists {
			originalID := si.ID
			si.ID = int64(s.ID)
			if originalID == 0 {
				continue
			}
		}
	}
}

func (a *DrillFlowAdapter) OnStepStarted(stepInstID int64, timeoutAt time.Time) {
	updates := map[string]interface{}{
		"status":     string(flowengine.StepStatusRunning),
		"start_time": time.Now(),
	}
	if !timeoutAt.IsZero() {
		updates["timeout_at"] = &timeoutAt
	}
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepInstID).Updates(updates)
}

func (a *DrillFlowAdapter) OnStepCompleted(stepInstID int64, operatorID int64, remark string) {
	var stepInst entity.StepInstance
	err := repository.DB.Where("id = ?", stepInstID).First(&stepInst).Error
	if err != nil {
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":          string(flowengine.StepStatusCompleted),
		"actual_operator": operatorID,
		"end_time":        &now,
		"remark":          remark,
	}
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepInstID).Updates(updates)
}

func (a *DrillFlowAdapter) OnStepIssue(stepInstID int64, operatorID int64, issueDesc string) {
}

func (a *DrillFlowAdapter) LogAction(stepInstID int64, action string, operatorID int64, content string) {
	var stepInst entity.StepInstance
	drillID := uint64(0)

	if stepInstID > 0 {
		if err := repository.DB.Where("id = ?", stepInstID).First(&stepInst).Error; err == nil {
			drillID = stepInst.DrillInstanceID
		}
	}

	if drillID == 0 && a.engine != nil {
		for flowInstID := range a.contexts {
			if inst, ok := a.engine.GetInstance(flowInstID); ok {
				for _, si := range inst.Steps {
					if si.ID == stepInstID {
						drillID = uint64(flowInstID)
						break
					}
				}
			}
			if drillID > 0 {
				break
			}
		}
	}

	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: drillID,
		Action:          action,
		OperatorID:      uint64(operatorID),
		Content:         content,
	})
}

func (a *DrillFlowAdapter) SetupEventSubscriptions(engine *flowengine.Engine) {
	a.engine = engine
	stepStartCh := make(chan flowengine.Event, 100)
	engine.GetEventBus().Subscribe(flowengine.EventStepStart, stepStartCh)
	go func() {
		for evt := range stepStartCh {
			a.handleStepStart(evt)
		}
	}()

	stepCompleteCh := make(chan flowengine.Event, 100)
	engine.GetEventBus().Subscribe(flowengine.EventStepComplete, stepCompleteCh)
	go func() {
		for evt := range stepCompleteCh {
			a.handleStepComplete(evt)
		}
	}()

	stepTimeoutCh := make(chan flowengine.Event, 100)
	engine.GetEventBus().Subscribe(flowengine.EventStepTimeout, stepTimeoutCh)
	go func() {
		for evt := range stepTimeoutCh {
			a.handleStepTimeout(evt)
		}
	}()

	stepIssueCh := make(chan flowengine.Event, 100)
	engine.GetEventBus().Subscribe(flowengine.EventStepIssue, stepIssueCh)
	go func() {
		for evt := range stepIssueCh {
			a.handleStepIssue(evt)
		}
	}()

	flowCompleteCh := make(chan flowengine.Event, 100)
	engine.GetEventBus().Subscribe(flowengine.EventFlowComplete, flowCompleteCh)
	go func() {
		for evt := range flowCompleteCh {
			a.handleFlowComplete(evt)
		}
	}()
}

func (a *DrillFlowAdapter) findStepInstance(drillID uint64, stepDefID int64) (*entity.StepInstance, error) {
	var stepInst entity.StepInstance
	err := repository.DB.
		Where("drill_instance_id = ? AND step_template_id = ?", drillID, uint64(stepDefID)).
		First(&stepInst).Error
	if err != nil {
		return nil, fmt.Errorf("step instance not found: drillID=%d, stepDefID=%d", drillID, stepDefID)
	}
	return &stepInst, nil
}

func (a *DrillFlowAdapter) handleStepStart(evt flowengine.Event) {
	drillID := uint64(evt.FlowInstID)
	stepDefID := evt.StepDefID
	payload := evt.Payload.(map[string]interface{})
	var timeoutAt *time.Time
	if ta, ok := payload["timeout_at"]; ok {
		switch v := ta.(type) {
		case time.Time:
			if !v.IsZero() {
				timeoutAt = &v
			}
		case *time.Time:
			if v != nil && !v.IsZero() {
				timeoutAt = v
			}
		}
	}

	stepInst, err := a.findStepInstance(drillID, stepDefID)
	if err != nil {
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":     string(flowengine.StepStatusRunning),
		"start_time": now,
	}
	if timeoutAt != nil && !timeoutAt.IsZero() {
		updates["timeout_at"] = timeoutAt
	}
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepInst.ID).Updates(updates)

	ctx := a.GetDrillContext(evt.FlowInstID)
	drillName := ""
	if ctx != nil {
		drillName = ctx.Name
	}

	_ = gin.H{}

	if a.wsManager != nil {
		a.wsManager.SendStepChange(uint(drillID), websocket.StepChangePayload{
			DrillID:       uint(drillID),
			StepID:        uint(stepInst.ID),
			StepName:      stepInst.Name,
			PreviousStatus: string(flowengine.StepStatusPending),
			NewStatus:     string(flowengine.StepStatusRunning),
		})
	}

	var assigneeIDs []uint64
	if stepInst.AssigneeIDs != "" && stepInst.AssigneeIDs != "[]" && stepInst.AssigneeIDs != "null" {
		_ = json.Unmarshal([]byte(stepInst.AssigneeIDs), &assigneeIDs)
	}

	for _, userID := range assigneeIDs {
		var deadline int64
		if timeoutAt != nil && !timeoutAt.IsZero() {
			deadline = timeoutAt.Unix()
		}
		if a.wsManager != nil {
			a.wsManager.SendTaskUpdate(uint(userID), websocket.TaskAssignPayload{
				UserID:    uint(userID),
				DrillID:   uint(drillID),
				StepID:    uint(stepInst.ID),
				StepName:  stepInst.Name,
				DrillName: drillName,
				Deadline:  deadline,
				Action:    "assigned",
			})
		}
		stepIDUint := uint64(stepInst.ID)
		stepNamePtr := &stepInst.Name
		drillNamePtr := &drillName
		if a.notificationService != nil {
			a.notificationService.CreateNotification(&entity.Notification{
				UserID:    userID,
				Type:      entity.NotificationTypeTaskAssigned,
				Title:     "任务已分配",
				Content:   fmt.Sprintf("[%s] 需要你执行：%s", drillName, stepInst.Name),
				DrillID:   &stepIDUint,
				DrillName: drillNamePtr,
				StepID:    &stepIDUint,
				StepName:  stepNamePtr,
				IsRead:    false,
			})
		}
	}
}

func (a *DrillFlowAdapter) handleStepComplete(evt flowengine.Event) {
	drillID := uint64(evt.FlowInstID)
	stepDefID := evt.StepDefID
	payload := evt.Payload.(map[string]interface{})
	operatorID := payload["operator_id"].(int64)

	stepInst, err := a.findStepInstance(drillID, stepDefID)
	if err != nil {
		return
	}

	now := time.Now()
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepInst.ID).Updates(map[string]interface{}{
		"status":          string(flowengine.StepStatusCompleted),
		"actual_operator": operatorID,
		"end_time":        &now,
	})

	drill := a.GetDrillContext(evt.FlowInstID)
	drillName := ""
	if drill != nil {
		drillName = drill.Name
	}

	if a.wsManager != nil {
		a.wsManager.SendStepChange(uint(drillID), websocket.StepChangePayload{
			DrillID:       uint(drillID),
			StepID:        uint(stepInst.ID),
			StepName:      stepInst.Name,
			PreviousStatus: string(flowengine.StepStatusRunning),
			NewStatus:     string(flowengine.StepStatusCompleted),
		})
	}

	logDrillID := drillID
	logStepID := stepInst.ID
	logContent := ""
	if remark, ok := payload["remark"].(string); ok {
		logContent = remark
	}
	var operatorName string
	if a.userRepo != nil {
		if u, err := a.userRepo.FindByID(uint64(operatorID)); err == nil && u != nil {
			operatorName = u.RealName
		}
	}
	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: logDrillID,
		StepInstanceID:  &logStepID,
		Action:          "complete",
		OperatorID:      uint64(operatorID),
		OperatorName:    operatorName,
		Content:         logContent,
	})

	stepIDUint := stepInst.ID
	stepNamePtr := &stepInst.Name
	if a.notificationService != nil {
		a.notificationService.CreateNotification(&entity.Notification{
			UserID:    uint64(operatorID),
			Type:      entity.NotificationTypeStepComplete,
			Title:     "步骤已完成",
			Content:   fmt.Sprintf("[%s] 的步骤 [%s] 已完成", drillName, stepInst.Name),
			DrillID:   &logDrillID,
			DrillName: &drillName,
			StepID:    &stepIDUint,
			StepName:  stepNamePtr,
			IsRead:    false,
		}, uint64(operatorID))
	}

	a.handleSubtaskCompletion(evt.FlowInstID, stepDefID)
}

func (a *DrillFlowAdapter) handleSubtaskCompletion(flowInstID, stepDefID int64) {
	stepDef, err := a.GetStepDef(flowInstID, stepDefID)
	if err != nil || stepDef.ParentStepID == 0 {
		return
	}

	parentStepDefID := stepDef.ParentStepID
	siblings, err := a.GetAllStepDefs(flowInstID)
	if err != nil {
		return
	}

	var childIDs []int64
	for _, sd := range siblings {
		if sd.ParentStepID == parentStepDefID {
			childIDs = append(childIDs, sd.ID)
		}
	}
	if len(childIDs) == 0 {
		return
	}

	inst, ok := a.engine.GetInstanceForMutate(flowInstID)
	if !ok {
		return
	}

	for _, childID := range childIDs {
		si, exists := inst.Steps[childID]
		if !exists {
			return
		}
		switch si.Status {
		case flowengine.StepStatusCompleted:
		case flowengine.StepStatusSkipped:
		case flowengine.StepStatusTimeout:
		case flowengine.StepStatusIssue:
		default:
			return
		}
	}

	a.autoCompleteParentStep(flowInstID, parentStepDefID)
}

func (a *DrillFlowAdapter) autoCompleteParentStep(flowInstID int64, parentStepDefID int64) {
	drillID := uint64(flowInstID)

	inst, ok := a.engine.GetInstanceForMutate(flowInstID)
	if !ok {
		return
	}

	parentSI, exists := inst.Steps[parentStepDefID]
	if !exists {
		return
	}

	switch parentSI.Status {
	case flowengine.StepStatusCompleted:
	case flowengine.StepStatusSkipped:
	case flowengine.StepStatusTimeout:
	case flowengine.StepStatusIssue:
		return
	}

	parentInst, err := a.findStepInstance(drillID, parentStepDefID)
	if err != nil {
		return
	}

	now := time.Now()
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", parentInst.ID).Updates(map[string]interface{}{
		"status":   string(flowengine.StepStatusCompleted),
		"end_time": &now,
	})

	parentSI.Status = flowengine.StepStatusCompleted
	parentSI.EndTime = &now
	inst.CurrentStepIDs = removeFromParentCurrent(inst.CurrentStepIDs, parentStepDefID)

	if a.wsManager != nil {
		a.wsManager.SendStepChange(uint(drillID), websocket.StepChangePayload{
			DrillID:        uint(drillID),
			StepID:         uint(parentInst.ID),
			StepName:       parentInst.Name,
			PreviousStatus: string(flowengine.StepStatusRunning),
			NewStatus:      string(flowengine.StepStatusCompleted),
		})
	}

	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: drillID,
		StepInstanceID:  &parentInst.ID,
		Action:          "auto_complete",
		Content:         fmt.Sprintf("[%s] 子任务全部完成，父步骤自动完成", parentInst.Name),
		CreatedAt:       now,
	})

	a.engine.CompleteStep(flowInstID, parentStepDefID, 0, "子任务全部完成")
	a.handleSubtaskCompletion(flowInstID, parentStepDefID)
}

func (a *DrillFlowAdapter) propagateStatusToParent(flowInstID int64, stepDefID int64, status flowengine.StepStatus, at time.Time) {
	stepDef, err := a.GetStepDef(flowInstID, stepDefID)
	if err != nil || stepDef.ParentStepID == 0 {
		return
	}

	parentStepDefID := stepDef.ParentStepID

	drillID := uint64(flowInstID)
	inst, ok := a.engine.GetInstanceForMutate(flowInstID)
	if !ok {
		return
	}

	parentSI, exists := inst.Steps[parentStepDefID]
	if !exists {
		return
	}

	switch parentSI.Status {
	case flowengine.StepStatusCompleted:
	case flowengine.StepStatusSkipped:
	case flowengine.StepStatusTimeout:
	case flowengine.StepStatusIssue:
		return
	}

	parentInst, err := a.findStepInstance(drillID, parentStepDefID)
	if err != nil {
		return
	}

	parentSI.Status = status
	parentSI.EndTime = &at

	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", parentInst.ID).Updates(map[string]interface{}{
		"status":   string(status),
		"end_time": &at,
	})

	if a.wsManager != nil {
		a.wsManager.SendStepChange(uint(drillID), websocket.StepChangePayload{
			DrillID:        uint(drillID),
			StepID:         uint(parentInst.ID),
			StepName:       parentInst.Name,
			PreviousStatus: string(flowengine.StepStatusRunning),
			NewStatus:      string(status),
		})
	}

	label := "超时"
	if status == flowengine.StepStatusIssue {
		label = "异常"
	}
	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: drillID,
		StepInstanceID:  &parentInst.ID,
		Action:          "parent_" + string(status),
		Content:         fmt.Sprintf("[%s] 子步骤%s，父步骤同步%s", parentInst.Name, label, label),
		CreatedAt:       at,
	})
}

func (a *DrillFlowAdapter) handleStepTimeout(evt flowengine.Event) {
	drillID := uint64(evt.FlowInstID)
	stepDefID := evt.StepDefID

	stepInst, err := a.findStepInstance(drillID, stepDefID)
	if err != nil {
		return
	}

	now := time.Now()
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepInst.ID).Updates(map[string]interface{}{
		"status":   string(flowengine.StepStatusTimeout),
		"end_time": &now,
	})

	// 写操作日志
	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: drillID,
		StepInstanceID:  &stepInst.ID,
		Action:          "timeout",
		Content:         fmt.Sprintf("[%s] 步骤已超时", stepInst.Name),
		CreatedAt:       now,
	})

	ctx := a.GetDrillContext(evt.FlowInstID)
	drillName := ""
	if ctx != nil {
		drillName = ctx.Name
	}

	if a.wsManager != nil {
		a.wsManager.SendStepChange(uint(drillID), websocket.StepChangePayload{
			DrillID:       uint(drillID),
			StepID:        uint(stepInst.ID),
			StepName:      stepInst.Name,
			PreviousStatus: string(flowengine.StepStatusRunning),
			NewStatus:     string(flowengine.StepStatusTimeout),
		})
	}

	if a.engine != nil {
		inst, _ := a.engine.GetInstanceForMutate(evt.FlowInstID)
		if inst != nil {
			si, ok := inst.Steps[evt.StepDefID]
			if ok {
				si.Status = flowengine.StepStatusTimeout
				si.EndTime = &now
			}
		}
	}

	var assigneeIDs []uint64
	if stepInst.AssigneeIDs != "" && stepInst.AssigneeIDs != "[]" {
		_ = json.Unmarshal([]byte(stepInst.AssigneeIDs), &assigneeIDs)
	}

	for _, userID := range assigneeIDs {
		drillIDForNotif := drillID
		stepIDForNotif := stepInst.ID
		if a.notificationService != nil {
			a.notificationService.CreateNotification(&entity.Notification{
				UserID:   userID,
				Type:     entity.NotificationTypeStepTimeout,
				Title:    "步骤超时",
				Content:  fmt.Sprintf("[%s] 的步骤 [%s] 已超时", drillName, stepInst.Name),
				DrillID:  &drillIDForNotif,
				StepID:   &stepIDForNotif,
				IsRead:   false,
			})
		}
	}

	a.propagateStatusToParent(evt.FlowInstID, stepDefID, flowengine.StepStatusTimeout, now)
}

func (a *DrillFlowAdapter) handleStepIssue(evt flowengine.Event) {
	drillID := uint64(evt.FlowInstID)
	stepDefID := evt.StepDefID
	payload := evt.Payload.(map[string]interface{})
	operatorID := payload["operator_id"].(int64)

	stepInst, err := a.findStepInstance(drillID, stepDefID)
	if err != nil {
		return
	}

	issueDesc := ""
	if desc, ok := payload["issue_desc"].(string); ok {
		issueDesc = desc
	}

	ctx := a.GetDrillContext(evt.FlowInstID)
	drillName := ""
	if ctx != nil {
		drillName = ctx.Name
	}

	if a.wsManager != nil {
		a.wsManager.SendStepChange(uint(drillID), websocket.StepChangePayload{
			DrillID:       uint(drillID),
			StepID:        uint(stepInst.ID),
			StepName:      stepInst.Name,
			PreviousStatus: string(flowengine.StepStatusRunning),
			NewStatus:     string(flowengine.StepStatusIssue),
		})
	}

	nw := time.Now()
	repository.DB.Model(&entity.StepInstance{}).Where("id = ?", stepInst.ID).Updates(map[string]interface{}{
		"status":     string(flowengine.StepStatusIssue),
		"issue_desc": issueDesc,
		"end_time":   &nw,
	})

	if a.engine != nil {
		inst, _ := a.engine.GetInstanceForMutate(evt.FlowInstID)
		if inst != nil {
			si, ok := inst.Steps[evt.StepDefID]
			if ok {
				si.Status = flowengine.StepStatusIssue
				si.IssueDesc = issueDesc
				si.EndTime = &nw
			}
		}
	}

	drillIDForLog := drillID
	stepIDForLog := stepInst.ID
	var operatorName string
	if a.userRepo != nil {
		if u, err := a.userRepo.FindByID(uint64(operatorID)); err == nil && u != nil {
			operatorName = u.RealName
		}
	}
	repository.DB.Create(&entity.DrillInstanceLog{
		DrillInstanceID: drillIDForLog,
		StepInstanceID:  &stepIDForLog,
		Action:          "issue",
		OperatorID:      uint64(operatorID),
		OperatorName:    operatorName,
		Content:         issueDesc,
	})

	if a.notificationService != nil {
		a.notificationService.CreateNotification(&entity.Notification{
			UserID:   uint64(operatorID),
			Type:     entity.NotificationTypeStepTimeout,
			Title:    "步骤异常上报",
			Content:  fmt.Sprintf("[%s] 的步骤 [%s] 已上报异常", drillName, stepInst.Name),
			DrillID:  &drillIDForLog,
			IsRead:   false,
		}, uint64(operatorID))
	}

	a.propagateStatusToParent(evt.FlowInstID, stepDefID, flowengine.StepStatusIssue, nw)
}

func (a *DrillFlowAdapter) handleFlowComplete(evt flowengine.Event) {
	drillID := uint64(evt.FlowInstID)
	ctx := a.GetDrillContext(evt.FlowInstID)
	drillName := ""
	if ctx != nil {
		drillName = ctx.Name
	}

	if a.notificationService != nil {
		a.notificationService.CreateNotification(&entity.Notification{
			UserID:   drillID,
			Type:     entity.NotificationTypeDrillCompleted,
			Title:    "演练已全部完成",
			Content:  fmt.Sprintf("[%s] 所有步骤已完成", drillName),
			DrillID:  &drillID,
			IsRead:   false,
		})
	}
}

func parseAssigneeIDsFromTemplate(assigneeIDsJSON string) []int64 {
	if assigneeIDsJSON == "" || assigneeIDsJSON == "[]" || assigneeIDsJSON == "null" {
		return nil
	}
	var uintIDs []uint64
	if err := json.Unmarshal([]byte(assigneeIDsJSON), &uintIDs); err != nil {
		return nil
	}
	ids := make([]int64, len(uintIDs))
	for i, uid := range uintIDs {
		ids[i] = int64(uid)
	}
	return ids
}

func parseAssigneeIDsFromInstance(assigneeIDsJSON string) []int64 {
	if assigneeIDsJSON == "" || assigneeIDsJSON == "[]" || assigneeIDsJSON == "null" {
		return nil
	}
	var uintIDs []uint64
	if err := json.Unmarshal([]byte(assigneeIDsJSON), &uintIDs); err != nil {
		return nil
	}
	ids := make([]int64, len(uintIDs))
	for i, uid := range uintIDs {
		ids[i] = int64(uid)
	}
	return ids
}

func (a *DrillFlowAdapter) BuildFlowDef(template *entity.DrillTemplate) *flowengine.FlowDef {
	flowDef := &flowengine.FlowDef{
		ID:   int64(template.ID),
		Name: template.Name,
	}
	for _, st := range template.Steps {
		var preStepIDs []int64
		if st.PreStepIDs != "" && st.PreStepIDs != "null" {
			var uintIDs []uint64
			if err := json.Unmarshal([]byte(st.PreStepIDs), &uintIDs); err == nil {
				for _, uid := range uintIDs {
					preStepIDs = append(preStepIDs, int64(uid))
				}
			}
		}
		parentStepID := int64(0)
		if st.ParentStepID != nil {
			parentStepID = int64(*st.ParentStepID)
		}
		flowDef.Steps = append(flowDef.Steps, &flowengine.StepDef{
			ID:                  int64(st.ID),
			Name:                st.Name,
			Seq:                 st.Seq,
			StepType:            flowengine.StepType(st.StepType),
			TimeoutMinutes:      st.TimeoutMinutes,
			PreStepIDs:          preStepIDs,
			GuideContent:        st.GuideContent,
			IsBlocking:          st.IsBlocking == 1,
			DefaultAssigneeRole: st.DefaultAssigneeRole,
			ParentStepID:        parentStepID,
		})
	}
	return flowDef
}

func (a *DrillFlowAdapter) BuildAssignees(drillID uint64) map[int64][]int64 {
	var steps []entity.StepInstance
	repository.DB.Where("drill_instance_id = ?", drillID).Find(&steps)

	assignees := make(map[int64][]int64)
	for _, step := range steps {
		stepDefID := int64(step.StepTemplateID)
		ids := parseAssigneeIDsFromInstance(step.AssigneeIDs)
		if len(ids) > 0 {
			assignees[stepDefID] = ids
		}
	}
	return assignees
}

func removeFromParentCurrent(currentIDs []int64, removeID int64) []int64 {
	result := make([]int64, 0, len(currentIDs))
	for _, id := range currentIDs {
		if id != removeID {
			result = append(result, id)
		}
	}
	return result
}
