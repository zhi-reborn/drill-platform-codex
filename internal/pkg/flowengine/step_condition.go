package flowengine

import "time"

func (e *Engine) advanceConditionSteps(inst *FlowInst, stepDefID int64, result ConditionResult) {
	loader := e.getStepLoader()
	if loader == nil {
		return
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, stepDefID)
	if err != nil || stepDef.StepType != StepTypeCondition || stepDef.Condition == nil {
		return
	}

	si, exists := inst.Steps[stepDefID]
	if !exists {
		return
	}
	si.ConditionResult = result

	var nextStepIDs []int64
	if result == ConditionPass {
		nextStepIDs = stepDef.Condition.TrueStepIDs
	} else {
		nextStepIDs = stepDef.Condition.FalseStepIDs
	}

	for _, nextStepDefID := range nextStepIDs {
		nextSI, exists := inst.Steps[nextStepDefID]
		if !exists || nextSI.Status != StepStatusPending {
			continue
		}
		e.activateStep(inst, nextSI)
	}

	e.checkFlowCompletion(inst)
}

func (e *Engine) evaluateCondition(inst *FlowInst, stepDefID int64, result ConditionResult) {
	loader := e.getStepLoader()
	if loader == nil {
		return
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, stepDefID)
	if err != nil || stepDef.StepType != StepTypeCondition {
		return
	}

	if stepDef.Condition == nil {
		return
	}

	if result == ConditionPass {
		for _, trueStepID := range stepDef.Condition.TrueStepIDs {
			si, exists := inst.Steps[trueStepID]
			if exists && si.Status == StepStatusPending {
				e.activateStep(inst, si)
			}
		}
		for _, falseStepID := range stepDef.Condition.FalseStepIDs {
			e.skipStep(inst, falseStepID, "condition_false")
		}
	} else {
		for _, falseStepID := range stepDef.Condition.FalseStepIDs {
			si, exists := inst.Steps[falseStepID]
			if exists && si.Status == StepStatusPending {
				e.activateStep(inst, si)
			}
		}
		for _, trueStepID := range stepDef.Condition.TrueStepIDs {
			e.skipStep(inst, trueStepID, "condition_true")
		}
	}
}

func (e *Engine) skipStep(inst *FlowInst, stepDefID int64, reason string) {
	si, exists := inst.Steps[stepDefID]
	if !exists || si.Status == StepStatusCompleted || si.Status == StepStatusSkipped {
		return
	}

	if si.Status == StepStatusRunning {
		inst.CurrentStepIDs = e.removeFromCurrentSteps(inst.CurrentStepIDs, si.StepDefID)
		e.timeoutScheduler.Unregister(inst.ID, si.StepDefID)
	}

	now := time.Now()
	oldStatus := si.Status
	si.Status = StepStatusSkipped
	si.EndTime = &now

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnStepStatusChanged(si.ID, oldStatus, StepStatusSkipped)
		cbs.LogAction(si.ID, "skip", 0, reason)
	}

	e.eventBus.emit(EventStepSkip, inst.ID, si.ID, si.StepDefID, map[string]interface{}{
		"reason": reason,
	})
}
