package flowengine

func (e *Engine) advanceSerialSteps(inst *FlowInst, completedStepDefID int64) {
	loader := e.getStepLoader()
	if loader == nil {
		return
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, completedStepDefID)
	if err != nil {
		return
	}

	if stepDef.StepType != StepTypeSerial {
		return
	}

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
	loader := e.getStepLoader()
	if loader == nil {
		return nil
	}

	allSteps, err := loader.GetAllStepDefs(inst.FlowDefID)
	if err != nil {
		return nil
	}

	var nextSteps []int64
	for _, sd := range allSteps {
		for _, preID := range sd.PreStepIDs {
			if preID == completedStepDefID {
				nextSteps = append(nextSteps, sd.ID)
				break
			}
		}
	}

	return nextSteps
}

func (e *Engine) allPredecessorsDone(inst *FlowInst, stepDefID int64) bool {
	loader := e.getStepLoader()
	if loader == nil {
		return false
	}

	stepDef, err := loader.GetStepDef(inst.FlowDefID, stepDefID)
	if err != nil {
		return false
	}

	for _, preID := range stepDef.PreStepIDs {
		si, exists := inst.Steps[preID]
		if !exists {
			return false
		}
		if si.Status != StepStatusCompleted && si.Status != StepStatusSkipped {
			return false
		}
	}

	return true
}
