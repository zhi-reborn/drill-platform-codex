package flowengine

func (e *Engine) advanceSerialSteps(inst *FlowInst, completedStepDefID int64) {
	nextSteps := e.findNextSteps(inst, completedStepDefID)
	for _, nextStepDefID := range nextSteps {
		if !e.allPredecessorsDone(inst, nextStepDefID) {
			continue
		}

		si, exists := inst.Steps[nextStepDefID]
		if !exists || si.Status != StepStatusPending {
			continue
		}

		e.activateStep(inst, si)
	}

	e.checkFlowCompletion(inst)
}

func (e *Engine) findNextSteps(inst *FlowInst, completedStepDefID int64) []int64 {
	var nextSteps []int64
	for stepDefID, si := range inst.Steps {
		for _, preID := range si.PreStepIDs {
			if preID == completedStepDefID {
				nextSteps = append(nextSteps, stepDefID)
				break
			}
		}
	}
	return nextSteps
}

func (e *Engine) allPredecessorsDone(inst *FlowInst, stepDefID int64) bool {
	si, exists := inst.Steps[stepDefID]
	if !exists {
		return false
	}

	for _, preID := range si.PreStepIDs {
		preSI, ok := inst.Steps[preID]
		if !ok {
			return false
		}
		if preSI.Status != StepStatusCompleted && preSI.Status != StepStatusSkipped {
			return false
		}
	}

	return true
}