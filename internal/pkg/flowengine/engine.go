package flowengine

import (
	"log"
	"sync"
	"time"
)

type Engine struct {
	mu               sync.RWMutex
	instances        map[int64]*FlowInst
	callbacks        PersistenceCallbacks
	callbacksMu      sync.RWMutex
	stepLoader       StepLoader
	stepLoaderMu     sync.RWMutex
	eventBus         *EventBus
	timeoutScheduler *TimeoutScheduler
}

func NewEngine() *Engine {
	eventBus := NewEventBus()
	timeoutScheduler := NewTimeoutScheduler(eventBus)
	return &Engine{
		instances:        make(map[int64]*FlowInst),
		eventBus:         eventBus,
		timeoutScheduler: timeoutScheduler,
	}
}

func (e *Engine) SetCallbacks(callbacks PersistenceCallbacks) {
	e.callbacksMu.Lock()
	defer e.callbacksMu.Unlock()
	e.callbacks = callbacks
}

func (e *Engine) SetStepLoader(loader StepLoader) {
	e.stepLoaderMu.Lock()
	defer e.stepLoaderMu.Unlock()
	e.stepLoader = loader
}

func (e *Engine) GetEventBus() *EventBus {
	return e.eventBus
}

func (e *Engine) TimeoutScheduler() *TimeoutScheduler {
	return e.timeoutScheduler
}

func (e *Engine) getCallbacks() PersistenceCallbacks {
	e.callbacksMu.RLock()
	defer e.callbacksMu.RUnlock()
	return e.callbacks
}

func (e *Engine) getStepLoader() StepLoader {
	e.stepLoaderMu.RLock()
	defer e.stepLoaderMu.RUnlock()
	return e.stepLoader
}

func (e *Engine) CreateInstance(flowDef *FlowDef, assignees map[int64][]int64, createdBy int64) (*FlowInst, error) {
	if len(flowDef.Steps) == 0 {
		return nil, ErrInvalidFlowDef
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	inst := &FlowInst{
		ID:        flowDef.ID,
		FlowDefID: flowDef.ID,
		Name:      flowDef.Name,
		Status:    FlowStatusPending,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		Steps:     make(map[int64]*StepInst),
		Assignees: assignees,
	}

	for _, stepDef := range flowDef.Steps {
		si := &StepInst{
			StepDefID:                stepDef.ID,
			Name:                     stepDef.Name,
			Seq:                      stepDef.Seq,
			Status:                   StepStatusPending,
			ParentStepID:             stepDef.ParentStepID,
			PreStepIDs:               stepDef.PreStepIDs,
			StepType:                 stepDef.StepType,
			Phase:                    stepDef.Phase,
			PhaseStep:                stepDef.PhaseStep,
			ExecutionMode:            stepDef.ExecutionMode,
			EstimatedDurationMinutes: stepDef.EstimatedDurationMinutes,
			EstimatedStartOffset:     stepDef.EstimatedStartOffset,
			JSONAttributes:           stepDef.JSONAttributes,
		}
		if users, ok := assignees[stepDef.ID]; ok {
			si.AssigneeIDs = users
		}
		inst.Steps[stepDef.ID] = si
	}

	e.instances[inst.ID] = inst

	e.eventBus.emit(EventFlowStart, inst.ID, 0, 0, nil)

	return inst, nil
}

func (e *Engine) Start(instanceID int64) error {
	e.mu.RLock()
	inst, ok := e.instances[instanceID]
	e.mu.RUnlock()
	if !ok {
		return ErrInstanceNotFound
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.Status != FlowStatusPending {
		return ErrInvalidStatus
	}

	oldStatus := inst.Status
	inst.Status = FlowStatusRunning
	now := time.Now()
	inst.StartTime = &now

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnFlowStatusChanged(instanceID, oldStatus, inst.Status)
	}

	e.activateInitialSteps(inst)

	e.eventBus.emit(EventFlowStart, instanceID, 0, 0, map[string]interface{}{
		"current_steps": inst.CurrentStepIDs,
	})

	return nil
}

func (e *Engine) GetInstance(instanceID int64) (*FlowInst, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	inst, ok := e.instances[instanceID]
	if !ok {
		return nil, false
	}
	return inst, true
}

func (e *Engine) GetInstanceForMutate(instanceID int64) (*FlowInst, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	inst, ok := e.instances[instanceID]
	return inst, ok
}

func (e *Engine) activateInitialSteps(inst *FlowInst) {
	// 只激活根级步骤（无父步骤）且无前序依赖的步骤
	var firstSteps []*StepInst
	for _, si := range inst.Steps {
		if len(si.PreStepIDs) == 0 && si.ParentStepID == 0 {
			firstSteps = append(firstSteps, si)
		}
	}

	for _, si := range firstSteps {
		e.activateStep(inst, si)
	}
}

func (e *Engine) activateStep(inst *FlowInst, si *StepInst) {
	// 已激活的步骤不重复激活
	if si.Status != StepStatusPending {
		return
	}

	// 自动启动 pending 状态的父步骤
	if si.ParentStepID != 0 {
		parentSI, parentExists := inst.Steps[si.ParentStepID]
		if parentExists && parentSI.Status == StepStatusPending {
			e.activateStep(inst, parentSI)
		}
	}

	// 父步骤的 activateChildSteps 可能已经激活了本步骤，再次检查
	if si.Status != StepStatusPending {
		return
	}

	now := time.Now()
	oldStatus := si.Status
	si.Status = StepStatusRunning
	si.StartTime = &now

	log.Printf("[FLOW] activateStep: step=%d name=%s old=%s new=%s instID=%d",
		si.StepDefID, si.Name, oldStatus, si.Status, si.ID)

	var timeoutAt time.Time
	loader := e.getStepLoader()
	if loader != nil {
		stepDef, err := loader.GetStepDef(inst.FlowDefID, si.StepDefID)
		if err == nil && stepDef.TimeoutMinutes > 0 {
			timeoutAt = now.Add(time.Duration(stepDef.TimeoutMinutes) * time.Minute)
			si.TimeoutAt = &timeoutAt
		}
	}

	inst.CurrentStepIDs = append(inst.CurrentStepIDs, si.StepDefID)

	if !timeoutAt.IsZero() {
		e.timeoutScheduler.Register(inst.ID, si.StepDefID, si.ID, timeoutAt)
	}

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnStepStatusChanged(si.ID, oldStatus, si.Status)
		if !timeoutAt.IsZero() {
			cbs.OnStepStarted(si.ID, timeoutAt)
		}
	}

	e.eventBus.emit(EventStepStart, inst.ID, si.ID, si.StepDefID, map[string]interface{}{
		"step_name":  si.Name,
		"timeout_at": timeoutAt,
	})

	// 激活子步骤：父步骤开始后，自动激活其首批子步骤（无前序依赖的子步骤）
	e.activateChildSteps(inst, si.StepDefID)
}

func (e *Engine) handleStepCompletion(inst *FlowInst, completedStepDefID int64) {
	si, exists := inst.Steps[completedStepDefID]
	if !exists {
		return
	}

	log.Printf("[FLOW] handleStepCompletion: step=%d name=%s type=%s", completedStepDefID, si.Name, si.StepType)

	switch si.StepType {
	case StepTypeSerial, "":
		e.advanceSerialSteps(inst, completedStepDefID)
	case StepTypeParallel:
		e.advanceParallelSteps(inst, completedStepDefID)
	case StepTypeAnyOf:
		e.advanceAnyOfSteps(inst, completedStepDefID)
	case StepTypeCondition:
		result := ConditionPass
		if si.ConditionResult != "" {
			result = si.ConditionResult
		}
		e.advanceConditionSteps(inst, completedStepDefID, result)
	}
}

// ManualStartStep 手动启动步骤（外部 API 调用入口）
// 校验前序步骤完成、父步骤已开始、步骤状态为 pending，然后通过 activateStep 激活
func (e *Engine) ManualStartStep(instanceID int64, stepDefID int64) error {
	e.mu.RLock()
	inst, ok := e.instances[instanceID]
	e.mu.RUnlock()
	if !ok {
		return ErrInstanceNotFound
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.Status != FlowStatusRunning {
		log.Printf("[FLOW] ManualStartStep rejected: instance=%d flowStatus=%s", instanceID, inst.Status)
		return ErrInstanceNotRunning
	}

	si, exists := inst.Steps[stepDefID]
	if !exists {
		return ErrStepNotFound
	}

	if si.Status != StepStatusPending {
		log.Printf("[FLOW] ManualStartStep rejected: step=%d name=%s status=%s expected=pending instID=%d",
			stepDefID, si.Name, si.Status, instanceID)
		return ErrInvalidStatus
	}

	if !e.allPredecessorsDone(inst, stepDefID) {
		return ErrPreStepsNotDone
	}

	// 父步骤如果还是 pending，activateStep 会自动启动它
	if si.ParentStepID != 0 {
		_, parentExists := inst.Steps[si.ParentStepID]
		if !parentExists {
			return ErrStepNotFound
		}
	}

	e.activateStep(inst, si)
	e.updateProgress(inst)

	return nil
}

func (e *Engine) removeFromCurrentSteps(currentIDs []int64, removeID int64) []int64 {
	result := make([]int64, 0, len(currentIDs))
	for _, id := range currentIDs {
		if id != removeID {
			result = append(result, id)
		}
	}
	return result
}

// activateChildSteps 激活父步骤下的首批子步骤（无前序依赖的子步骤）
func (e *Engine) activateChildSteps(inst *FlowInst, parentStepDefID int64) {
	for _, childSI := range inst.Steps {
		if childSI.ParentStepID == parentStepDefID && len(childSI.PreStepIDs) == 0 && childSI.Status == StepStatusPending {
			e.activateStep(inst, childSI)
		}
	}
}

// EnsureChildrenActivated 确保运行中/已完成的父步骤的 pending 子步骤被激活
// 用于 Recover 场景：状态从数据库同步后，running 步骤的子步骤可能未被激活
func (e *Engine) EnsureChildrenActivated(instanceID int64, parentStepDefID int64) {
	e.mu.RLock()
	inst, ok := e.instances[instanceID]
	e.mu.RUnlock()
	if !ok {
		return
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	parentSI, ok := inst.Steps[parentStepDefID]
	if !ok {
		return
	}
	if parentSI.Status != StepStatusRunning && parentSI.Status != StepStatusCompleted {
		return
	}
	e.activateChildSteps(inst, parentStepDefID)
}

// AdvanceFlow 在步骤达到终态（超时/异常等）后推进流程
func (e *Engine) AdvanceFlow(instanceID int64, stepDefID int64) error {
	e.mu.RLock()
	inst, ok := e.instances[instanceID]
	e.mu.RUnlock()
	if !ok {
		return ErrInstanceNotFound
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	inst.CurrentStepIDs = e.removeFromCurrentSteps(inst.CurrentStepIDs, stepDefID)
	e.timeoutScheduler.Unregister(inst.ID, stepDefID)

	e.handleStepCompletion(inst, stepDefID)
	e.updateProgress(inst)

	return nil
}
