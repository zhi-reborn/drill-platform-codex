package flowengine

func (e *Engine) advanceParallelSteps(inst *FlowInst, completedStepDefID int64) {
	si, exists := inst.Steps[completedStepDefID]
	if !exists || si.StepType != StepTypeParallel {
		return
	}

	siblingStepDefIDs := e.getParallelSiblings(inst, completedStepDefID)

	allDone := true
	for _, siblingID := range siblingStepDefIDs {
		sib, exists := inst.Steps[siblingID]
		if !exists {
			allDone = false
			continue
		}
		if sib.Status == StepStatusPending || sib.Status == StepStatusRunning {
			allDone = false
		}
	}

	if allDone {
		allGroupIDs := append([]int64{completedStepDefID}, siblingStepDefIDs...)
		e.activateDependentSteps(inst, allGroupIDs)
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

	if stepDef.StepType != StepTypeParallel || stepDef.Condition == nil {
		return nil
	}

	var siblings []int64
	for _, childID := range stepDef.Condition.TrueStepIDs {
		if childID != stepDefID {
			siblings = append(siblings, childID)
		}
	}
	return siblings
}

func (e *Engine) activateDependentSteps(inst *FlowInst, finishedGroupIDs []int64) {
	finishedSet := make(map[int64]bool)
	for _, id := range finishedGroupIDs {
		finishedSet[id] = true
	}

	for stepDefID, si := range inst.Steps {
		if e.hasAllPredecessorsInSet(si, finishedSet) {
			if si.Status == StepStatusPending {
				e.activateStep(inst, inst.Steps[stepDefID])
			}
		}
	}
}

func (e *Engine) hasAllPredecessorsInSet(si *StepInst, set map[int64]bool) bool {
	for _, preID := range si.PreStepIDs {
		if !set[preID] {
			return false
		}
	}
	return len(si.PreStepIDs) > 0
}