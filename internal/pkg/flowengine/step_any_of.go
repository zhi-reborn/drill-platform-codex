package flowengine

import "time"

func (e *Engine) advanceAnyOfSteps(inst *FlowInst, completedStepDefID int64) {
	completedStep := e.getStepDefByType(inst, completedStepDefID, StepTypeAnyOf)
	if completedStep != nil {
		e.handleAnyOfGroupComplete(inst, completedStepDefID)
		return
	}

	siblingIDs := e.getAnyOfSiblings(inst, completedStepDefID)
	if len(siblingIDs) == 0 {
		return
	}

	e.skipPendingAnyOfSiblings(inst, completedStepDefID, siblingIDs)

	e.activateDependentSteps(inst, append(siblingIDs, completedStepDefID))
	e.checkFlowCompletion(inst)
}

func (e *Engine) handleAnyOfGroupComplete(inst *FlowInst, completedStepDefID int64) {
	completedSet := make(map[int64]bool)
	completedSet[completedStepDefID] = true

	e.activateDependentSteps(inst, []int64{completedStepDefID})
}

func (e *Engine) getAnyOfSiblings(inst *FlowInst, stepDefID int64) []int64 {
	loader := e.getStepLoader()
	if loader == nil {
		return nil
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, stepDefID)
	if err != nil || len(stepDef.PreStepIDs) == 0 {
		return nil
	}

	parentStepID := stepDef.PreStepIDs[0]
	parentDef, err := loader.GetStepDef(inst.FlowDefID, parentStepID)
	if err != nil || parentDef.StepType != StepTypeAnyOf {
		return nil
	}

	var siblings []int64
	for _, childID := range parentDef.Condition.TrueStepIDs {
		if childID != stepDefID {
			siblings = append(siblings, childID)
		}
	}

	return siblings
}

func (e *Engine) skipPendingAnyOfSiblings(inst *FlowInst, completedStepDefID int64, siblingIDs []int64) {
	now := time.Now()
	for _, siblingID := range siblingIDs {
		si, exists := inst.Steps[siblingID]
		if !exists {
			continue
		}

		if si.Status == StepStatusPending || si.Status == StepStatusRunning {
			oldStatus := si.Status
			si.Status = StepStatusSkipped
			si.EndTime = &now

			inst.CurrentStepIDs = e.removeFromCurrentSteps(inst.CurrentStepIDs, si.StepDefID)

			e.timeoutScheduler.Unregister(inst.ID, si.StepDefID)

			if cbs := e.getCallbacks(); cbs != nil {
				cbs.OnStepStatusChanged(si.ID, oldStatus, StepStatusSkipped)
				cbs.LogAction(si.ID, "skip_auto", 0, "auto-skipped due to any_of sibling completion")
			}

			e.eventBus.emit(EventStepSkip, inst.ID, si.ID, si.StepDefID, map[string]interface{}{
				"reason": "any_of_sibling_completed",
			})
		}
	}
}

func (e *Engine) activateAnyOfGroup(inst *FlowInst, groupStepDefIDs []int64) {
	for _, stepDefID := range groupStepDefIDs {
		si, exists := inst.Steps[stepDefID]
		if !exists || si.Status != StepStatusPending {
			continue
		}
		if !e.allPredecessorsDone(inst, stepDefID) {
			continue
		}
		e.activateStep(inst, si)
	}
}
