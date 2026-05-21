package flowengine

import (
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
			StepDefID:    stepDef.ID,
			Name:         stepDef.Name,
			Seq:          stepDef.Seq,
			Status:       StepStatusPending,
			ParentStepID: stepDef.ParentStepID,
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
	var firstSteps []*StepInst
	for _, si := range inst.Steps {
		if e.hasNoPredecessors(inst, si.StepDefID) {
			firstSteps = append(firstSteps, si)
		}
	}

	for _, si := range firstSteps {
		e.activateStep(inst, si)
	}
}

func (e *Engine) hasNoPredecessors(inst *FlowInst, stepDefID int64) bool {
	loader := e.getStepLoader()
	if loader == nil {
		return false
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, stepDefID)
	if err != nil {
		return false
	}

	return len(stepDef.PreStepIDs) == 0
}

func (e *Engine) activateStep(inst *FlowInst, si *StepInst) {
	now := time.Now()
	oldStatus := si.Status
	si.Status = StepStatusRunning
	si.StartTime = &now

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

	e.timeoutScheduler.Register(inst.ID, si.StepDefID, si.ID, timeoutAt)

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
}

func (e *Engine) handleStepCompletion(inst *FlowInst, completedStepDefID int64) {
	loader := e.getStepLoader()
	if loader == nil {
		return
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, completedStepDefID)
	if err != nil {
		return
	}

	switch stepDef.StepType {
	case StepTypeSerial, "":
		e.advanceSerialSteps(inst, completedStepDefID)
	case StepTypeParallel:
		e.advanceParallelSteps(inst, completedStepDefID)
	case StepTypeAnyOf:
		e.advanceAnyOfSteps(inst, completedStepDefID)
	case StepTypeCondition:
		result := ConditionPass
		if si, ok := inst.Steps[completedStepDefID]; ok && si.ConditionResult != "" {
			result = si.ConditionResult
		}
		e.advanceConditionSteps(inst, completedStepDefID, result)
	}
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
