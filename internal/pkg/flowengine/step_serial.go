package flowengine

import (
	"log"
)

func (e *Engine) advanceSerialSteps(inst *FlowInst, completedStepDefID int64) {
	nextSteps := e.findNextSteps(inst, completedStepDefID)
	log.Printf("[FLOW] advanceSerialSteps: completedStep=%d, nextSteps=%v", completedStepDefID, nextSteps)

	for _, nextStepDefID := range nextSteps {
		if !e.allPredecessorsDone(inst, nextStepDefID) {
			si, _ := inst.Steps[nextStepDefID]
			var preIDs []int64
			if si != nil {
				preIDs = si.PreStepIDs
			}
			log.Printf("[FLOW] advanceSerialSteps: step=%d preNotDone, preIDs=%v", nextStepDefID, preIDs)
			continue
		}

		si, exists := inst.Steps[nextStepDefID]
		if !exists {
			log.Printf("[FLOW] advanceSerialSteps: step=%d NOT in inst.Steps", nextStepDefID)
			continue
		}
		if si.Status != StepStatusPending {
			log.Printf("[FLOW] advanceSerialSteps: step=%d status=%s (expected pending)", nextStepDefID, si.Status)
			continue
		}

		log.Printf("[FLOW] advanceSerialSteps: ACTIVATING step=%d name=%s", nextStepDefID, si.Name)
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
		if !isDependencySatisfiedStatus(preSI.Status) {
			return false
		}
	}

	return true
}

// isDependencySatisfiedStatus 判断前序依赖是否已满足。
// 跳过是终态，但不代表流程依赖完成，不能解锁后续步骤。
func isDependencySatisfiedStatus(status StepStatus) bool {
	return status == StepStatusCompleted || status == StepStatusTimeout || status == StepStatusIssue
}

// isTerminalStatus 判断步骤是否处于终态（已完成/已跳过/已超时/异常）
func isTerminalStatus(status StepStatus) bool {
	return status == StepStatusCompleted || status == StepStatusSkipped || status == StepStatusTimeout || status == StepStatusIssue
}
