package flowengine

import (
	"time"
)

func (e *Engine) CompleteStep(instanceID, stepDefID int64, operatorID int64, remark string) error {
	e.mu.RLock()
	inst, ok := e.instances[instanceID]
	e.mu.RUnlock()
	if !ok {
		return ErrInstanceNotFound
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.Status != FlowStatusRunning {
		return ErrInstanceNotRunning
	}

	si, exists := inst.Steps[stepDefID]
	if !exists {
		return ErrStepNotFound
	}

	if si.Status != StepStatusRunning {
		return ErrStepNotActive
	}

	now := time.Now()
	oldStatus := si.Status
	si.Status = StepStatusCompleted
	si.ActualOperator = &operatorID
	si.EndTime = &now
	si.Remark = remark

	inst.CurrentStepIDs = e.removeFromCurrentSteps(inst.CurrentStepIDs, stepDefID)

	e.timeoutScheduler.Unregister(inst.ID, stepDefID)

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnStepStatusChanged(si.ID, oldStatus, StepStatusCompleted)
		cbs.OnStepCompleted(si.ID, operatorID, remark)
		cbs.LogAction(si.ID, "complete", operatorID, remark)
	}

	e.eventBus.emit(EventStepComplete, inst.ID, si.ID, si.StepDefID, map[string]interface{}{
		"step_name":     si.Name,
		"operator_id":   operatorID,
		"remark":        remark,
		"completed_at":  now,
	})

	e.handleStepCompletion(inst, stepDefID)

	e.updateProgress(inst)

	return nil
}

func (e *Engine) ReportIssue(instanceID, stepDefID int64, operatorID int64, issue string) error {
	e.mu.RLock()
	inst, ok := e.instances[instanceID]
	e.mu.RUnlock()
	if !ok {
		return ErrInstanceNotFound
	}

	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.Status != FlowStatusRunning {
		return ErrInstanceNotRunning
	}

	si, exists := inst.Steps[stepDefID]
	if !exists {
		return ErrStepNotFound
	}

	if si.Status != StepStatusRunning {
		return ErrStepNotActive
	}

	oldStatus := si.Status
	si.Status = StepStatusIssue
	si.IssueDesc = issue

	inst.Status = FlowStatusIssue

	e.timeoutScheduler.Unregister(inst.ID, stepDefID)

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnStepStatusChanged(si.ID, oldStatus, StepStatusIssue)
		cbs.OnStepIssue(si.ID, operatorID, issue)
		cbs.OnFlowStatusChanged(inst.ID, FlowStatusRunning, FlowStatusIssue)
		cbs.LogAction(si.ID, "issue", operatorID, issue)
	}

	e.eventBus.emit(EventStepIssue, inst.ID, si.ID, si.StepDefID, map[string]interface{}{
		"step_name":   si.Name,
		"operator_id": operatorID,
		"issue_desc":  issue,
	})

	return nil
}

func (e *Engine) updateProgress(inst *FlowInst) {
	total, completed := e.countLeafProgress(inst)
	if total == 0 {
		return
	}

	inst.ProgressPct = (completed * 100) / total
}

func (e *Engine) countLeafProgress(inst *FlowInst) (total, completed int) {
	parentIDs := make(map[int64]bool)
	for _, si := range inst.Steps {
		if si.ParentStepID != 0 {
			parentIDs[si.ParentStepID] = true
		}
	}

	for _, si := range inst.Steps {
		if !parentIDs[si.StepDefID] {
			total++
			if si.Status == StepStatusCompleted || si.Status == StepStatusSkipped {
				completed++
			}
		}
	}
	return
}

func (e *Engine) checkFlowCompletion(inst *FlowInst) {
	parentIDs := make(map[int64]bool)
	for _, si := range inst.Steps {
		if si.ParentStepID != 0 {
			parentIDs[si.ParentStepID] = true
		}
	}

	for _, si := range inst.Steps {
		if !parentIDs[si.StepDefID] && si.Status != StepStatusCompleted && si.Status != StepStatusSkipped {
			return
		}
	}

	now := time.Now()
	oldStatus := inst.Status
	inst.Status = FlowStatusCompleted
	inst.EndTime = &now

	if cbs := e.getCallbacks(); cbs != nil {
		cbs.OnFlowStatusChanged(inst.ID, oldStatus, FlowStatusCompleted)
	}

	e.eventBus.emit(EventFlowComplete, inst.ID, 0, 0, map[string]interface{}{
		"progress_pct": inst.ProgressPct,
		"end_time":     now,
	})
}
