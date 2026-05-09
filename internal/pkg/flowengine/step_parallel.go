package flowengine

func (e *Engine) advanceParallelSteps(inst *FlowInst, completedStepDefID int64) {
	completedStep := e.getStepDefByType(inst, completedStepDefID, StepTypeParallel)
	if completedStep == nil {
		return
	}

	siblingStepDefIDs := e.getParallelSiblings(inst, completedStepDefID)

	allDone := true
	for _, siblingID := range siblingStepDefIDs {
		si, exists := inst.Steps[siblingID]
		if !exists {
			allDone = false
			continue
		}
		if si.Status == StepStatusPending || si.Status == StepStatusRunning {
			allDone = false
		}
	}

	if allDone {
		e.activateDependentSteps(inst, siblingStepDefIDs)
		e.checkFlowCompletion(inst)
	}
}

func (e *Engine) activateParallelGroup(inst *FlowInst, groupStepDefIDs []int64) {
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

func (e *Engine) getParallelSiblings(inst *FlowInst, stepDefID int64) []int64 {
	loader := e.getStepLoader()
	if loader == nil {
		return nil
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, stepDefID)
	if err != nil {
		return nil
	}

	if len(stepDef.PreStepIDs) == 0 {
		return nil
	}

	parentStepID := stepDef.PreStepIDs[0]
	parentDef, err := loader.GetStepDef(inst.FlowDefID, parentStepID)
	if err != nil || parentDef.StepType != StepTypeParallel {
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

func (e *Engine) getStepDefByType(inst *FlowInst, stepDefID int64, stepType StepType) *StepDef {
	loader := e.getStepLoader()
	if loader == nil {
		return nil
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, stepDefID)
	if err != nil {
		return nil
	}

	if stepDef.StepType == stepType {
		return stepDef
	}

	return nil
}

func (e *Engine) activateDependentSteps(inst *FlowInst, finishedGroupIDs []int64) {
	finishedSet := make(map[int64]bool)
	for _, id := range finishedGroupIDs {
		finishedSet[id] = true
	}

	loader := e.getStepLoader()
	if loader == nil {
		return
	}

	allSteps, err := loader.GetAllStepDefs(inst.FlowDefID)
	if err != nil {
		return
	}

	for _, sd := range allSteps {
		if e.hasAllPredecessorsInSet(sd, finishedSet) {
			si, exists := inst.Steps[sd.ID]
			if exists && si.Status == StepStatusPending {
				e.activateStep(inst, si)
			}
		}
	}
}

func (e *Engine) hasAllPredecessorsInSet(stepDef *StepDef, set map[int64]bool) bool {
	for _, preID := range stepDef.PreStepIDs {
		if !set[preID] {
			return false
		}
	}
	return len(stepDef.PreStepIDs) > 0
}
