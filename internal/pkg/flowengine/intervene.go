package flowengine

import (
	"time"
)

func (e *Engine) Intervene(instanceID int64, action InterveneAction, targetStepID *int64, operatorID int64) error {
	e.mu.RLock()
	inst, ok := e.instances[instanceID]
	e.mu.RUnlock()
	if !ok {
		return ErrInstanceNotFound
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	switch action {
	case ActionPause:
		return e.handlePause(inst)
	case ActionResume:
		return e.handleResume(inst)
	case ActionTerminate:
		return e.handleTerminate(inst)
	case ActionSkip:
		if targetStepID == nil {
			return ErrStepNotFound
		}
		return e.handleSkip(inst, *targetStepID, operatorID)
	case ActionForceComplete:
		if targetStepID == nil {
			return ErrStepNotFound
		}
		return e.handleForceComplete(inst, *targetStepID, operatorID)
	default:
		return ErrInvalidStatus
	}
}

func (e *Engine) handlePause(inst *FlowInst) error {
	if inst.Status != FlowStatusRunning && inst.Status != FlowStatusIssue {
		return ErrInvalidStatus
	}

	oldStatus := inst.Status
	inst.Status = FlowStatusPaused

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnFlowStatusChanged(inst.ID, oldStatus, FlowStatusPaused)
	}

	e.eventBus.emit(EventFlowPause, inst.ID, 0, 0, nil)
	return nil
}

func (e *Engine) handleResume(inst *FlowInst) error {
	if inst.Status != FlowStatusPaused {
		return ErrInvalidStatus
	}

	oldStatus := inst.Status
	inst.Status = FlowStatusRunning

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnFlowStatusChanged(inst.ID, oldStatus, FlowStatusRunning)
	}

	e.eventBus.emit(EventFlowResume, inst.ID, 0, 0, nil)
	return nil
}

func (e *Engine) handleTerminate(inst *FlowInst) error {
	if inst.Status == FlowStatusCompleted {
		return ErrInvalidStatus
	}

	oldStatus := inst.Status
	inst.Status = FlowStatusTerminated
	now := time.Now()
	inst.EndTime = &now

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnFlowStatusChanged(inst.ID, oldStatus, FlowStatusTerminated)
	}

	e.eventBus.emit(EventFlowTerminate, inst.ID, 0, 0, nil)
	return nil
}

func (e *Engine) handleSkip(inst *FlowInst, stepDefID int64, operatorID int64) error {
	if inst.Status != FlowStatusRunning && inst.Status != FlowStatusIssue && inst.Status != FlowStatusPaused {
		return ErrInstanceNotRunning
	}

	si, exists := inst.Steps[stepDefID]
	if !exists {
		return ErrStepNotFound
	}

	loader := e.getStepLoader()
	var stepDef *StepDef
	if loader != nil {
		stepDef, _ = loader.GetStepDef(inst.FlowDefID, stepDefID)
	}

	if stepDef != nil && stepDef.IsBlocking && si.Status == StepStatusPending {
		return ErrInvalidStatus
	}

	if si.Status == StepStatusRunning {
		inst.CurrentStepIDs = e.removeFromCurrentSteps(inst.CurrentStepIDs, stepDefID)
		e.timeoutScheduler.Unregister(inst.ID, stepDefID)
	}

	now := time.Now()
	oldStatus := si.Status
	si.Status = StepStatusSkipped
	si.EndTime = &now

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnStepStatusChanged(si.ID, oldStatus, StepStatusSkipped)
		cbs.LogAction(si.ID, "skip", operatorID, "director skipped")
	}

	e.eventBus.emit(EventStepSkip, inst.ID, si.ID, si.StepDefID, map[string]interface{}{
		"operator_id": operatorID,
		"reason":      "director_intervention",
	})

	if inst.Status == FlowStatusIssue {
		inst.Status = FlowStatusRunning
		if cbs := e.getCallbacks(); cbs != nil {
			cbs.OnFlowStatusChanged(inst.ID, FlowStatusIssue, FlowStatusRunning)
		}
	}

	e.advanceSerialSteps(inst, stepDefID)

	return nil
}

func (e *Engine) handleForceComplete(inst *FlowInst, stepDefID int64, operatorID int64) error {
	if inst.Status != FlowStatusRunning && inst.Status != FlowStatusIssue && inst.Status != FlowStatusPaused {
		return ErrInstanceNotRunning
	}

	si, exists := inst.Steps[stepDefID]
	if !exists {
		return ErrStepNotFound
	}

	if si.Status == StepStatusCompleted || si.Status == StepStatusSkipped {
		return ErrInvalidStatus
	}

	now := time.Now()
	oldStatus := si.Status
	si.Status = StepStatusCompleted
	si.ActualOperator = &operatorID
	si.EndTime = &now

	inst.CurrentStepIDs = e.removeFromCurrentSteps(inst.CurrentStepIDs, stepDefID)
	e.timeoutScheduler.Unregister(inst.ID, stepDefID)

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnStepStatusChanged(si.ID, oldStatus, StepStatusCompleted)
		cbs.OnStepCompleted(si.ID, operatorID, "forced by director")
		cbs.LogAction(si.ID, "force_complete", operatorID, "director forced completion")
	}

	e.eventBus.emit(EventStepForceComplete, inst.ID, si.ID, si.StepDefID, map[string]interface{}{
		"operator_id": operatorID,
	})

	if inst.Status == FlowStatusIssue {
		inst.Status = FlowStatusRunning
		if cbs := e.getCallbacks(); cbs != nil {
			cbs.OnFlowStatusChanged(inst.ID, FlowStatusIssue, FlowStatusRunning)
		}
	}

	e.handleStepCompletion(inst, stepDefID)
	e.updateProgress(inst)

	return nil
}
