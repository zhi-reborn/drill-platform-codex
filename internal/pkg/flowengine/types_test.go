package flowengine

import (
	"testing"
)

// ============ FlowStatus 测试 ============

func TestFlowStatusConstants(t *testing.T) {
	expected := map[FlowStatus]string{
		FlowStatusPending:    "pending",
		FlowStatusRunning:    "running",
		FlowStatusPaused:     "paused",
		FlowStatusCompleted:  "completed",
		FlowStatusTerminated: "terminated",
		FlowStatusIssue:      "issue",
	}

	for status, expectedStr := range expected {
		if string(status) != expectedStr {
			t.Errorf("FlowStatus %q != expected %q", status, expectedStr)
		}
	}
}

// ============ StepStatus 测试 ============

func TestStepStatusConstants(t *testing.T) {
	expected := map[StepStatus]string{
		StepStatusPending:   "pending",
		StepStatusRunning:   "running",
		StepStatusCompleted: "completed",
		StepStatusTimeout:   "timeout",
		StepStatusSkipped:   "skipped",
		StepStatusIssue:     "issue",
	}

	for status, expectedStr := range expected {
		if string(status) != expectedStr {
			t.Errorf("StepStatus %q != expected %q", status, expectedStr)
		}
	}
}

// ============ StepType 测试 ============

func TestStepTypeConstants(t *testing.T) {
	expected := map[StepType]string{
		StepTypeSerial:    "serial",
		StepTypeParallel:  "parallel",
		StepTypeAnyOf:     "any_of",
		StepTypeCondition: "condition",
	}

	for stepType, expectedStr := range expected {
		if string(stepType) != expectedStr {
			t.Errorf("StepType %q != expected %q", stepType, expectedStr)
		}
	}
}

// ============ InterveneAction 测试 ============

func TestInterveneActionConstants(t *testing.T) {
	expected := map[InterveneAction]string{
		ActionPause:         "pause",
		ActionResume:        "resume",
		ActionTerminate:     "terminate",
		ActionSkip:          "skip",
		ActionForceComplete: "force_complete",
	}

	for action, expectedStr := range expected {
		if string(action) != expectedStr {
			t.Errorf("InterveneAction %q != expected %q", action, expectedStr)
		}
	}
}

// ============ EventType 测试 ============

func TestEventTypeConstants(t *testing.T) {
	expected := map[EventType]string{
		EventFlowStart:       "flow_start",
		EventFlowComplete:    "flow_complete",
		EventFlowPause:       "flow_pause",
		EventFlowResume:      "flow_resume",
		EventFlowTerminate:   "flow_terminate",
		EventStepStart:       "step_start",
		EventStepComplete:    "step_complete",
		EventStepTimeout:     "step_timeout",
		EventStepIssue:       "step_issue",
		EventStepSkip:        "step_skip",
		EventStepForceComplete: "step_force_complete",
	}

	for eventType, expectedStr := range expected {
		if string(eventType) != expectedStr {
			t.Errorf("EventType %q != expected %q", eventType, expectedStr)
		}
	}
}

// ============ ConditionResult 测试 ============

func TestConditionResultConstants(t *testing.T) {
	if string(ConditionPass) != "pass" {
		t.Errorf("ConditionPass = %q, want %q", ConditionPass, "pass")
	}
	if string(ConditionFail) != "fail" {
		t.Errorf("ConditionFail = %q, want %q", ConditionFail, "fail")
	}
}

// ============ FlowDef 测试 ============

func TestFlowDefValidWithSteps(t *testing.T) {
	fd := &FlowDef{
		ID:   1,
		Name: "test-flow",
		Steps: []*StepDef{
			{ID: 1, Name: "step1", Seq: 1, StepType: StepTypeSerial},
		},
	}

	if fd.ID != 1 {
		t.Errorf("FlowDef.ID = %d, want 1", fd.ID)
	}
	if fd.Name != "test-flow" {
		t.Errorf("FlowDef.Name = %q, want %q", fd.Name, "test-flow")
	}
	if len(fd.Steps) != 1 {
		t.Errorf("FlowDef.Steps length = %d, want 1", len(fd.Steps))
	}
}

func TestFlowDefEmptySteps(t *testing.T) {
	fd := &FlowDef{
		ID:    1,
		Name:  "empty-flow",
		Steps: []*StepDef{},
	}

	if len(fd.Steps) != 0 {
		t.Errorf("FlowDef.Steps length = %d, want 0", len(fd.Steps))
	}
}

// ============ StepDef 测试 ============

func TestStepDefSerial(t *testing.T) {
	sd := &StepDef{
		ID:             1,
		Name:           "deploy-service",
		Seq:            1,
		StepType:       StepTypeSerial,
		TimeoutMinutes: 30,
		PreStepIDs:     []int64{},
		IsBlocking:     true,
	}

	if sd.StepType != StepTypeSerial {
		t.Errorf("StepDef.StepType = %q, want %q", sd.StepType, StepTypeSerial)
	}
	if sd.TimeoutMinutes != 30 {
		t.Errorf("StepDef.TimeoutMinutes = %d, want 30", sd.TimeoutMinutes)
	}
	if !sd.IsBlocking {
		t.Error("StepDef.IsBlocking should be true")
	}
}

func TestStepDefConditionWithCondition(t *testing.T) {
	sd := &StepDef{
		ID:       1,
		Name:     "health-check",
		StepType: StepTypeCondition,
		Condition: &ConditionDef{
			Expression:   "healthy == true",
			TrueStepIDs:  []int64{2},
			FalseStepIDs: []int64{3},
		},
	}

	if sd.Condition == nil {
		t.Fatal("Condition is nil")
	}
	if len(sd.Condition.TrueStepIDs) != 1 || sd.Condition.TrueStepIDs[0] != 2 {
		t.Errorf("TrueStepIDs = %v, want [2]", sd.Condition.TrueStepIDs)
	}
	if len(sd.Condition.FalseStepIDs) != 1 || sd.Condition.FalseStepIDs[0] != 3 {
		t.Errorf("FalseStepIDs = %v, want [3]", sd.Condition.FalseStepIDs)
	}
}

// ============ Error 测试 ============

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		err         error
		expectedMsg string
	}{
		{ErrInstanceNotFound, "flow instance not found"},
		{ErrStepNotFound, "step not found"},
		{ErrInvalidStatus, "invalid status transition"},
		{ErrInstanceNotRunning, "flow instance is not running"},
		{ErrStepNotActive, "step is not in running status"},
		{ErrPreStepsNotDone, "predecessor steps not completed"},
		{ErrInvalidFlowDef, "invalid flow definition: no steps"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.expectedMsg {
			t.Errorf("Error message = %q, want %q", tt.err.Error(), tt.expectedMsg)
		}
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	errors := []error{
		ErrInstanceNotFound,
		ErrStepNotFound,
		ErrInvalidStatus,
		ErrInstanceNotRunning,
		ErrStepNotActive,
		ErrPreStepsNotDone,
		ErrInvalidFlowDef,
	}

	for i := 0; i < len(errors); i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[i] == errors[j] {
				t.Errorf("errors[%d] == errors[%d], should be distinct", i, j)
			}
		}
	}
}
