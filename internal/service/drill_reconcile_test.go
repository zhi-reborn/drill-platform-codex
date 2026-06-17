package service

import (
	"testing"

	"drill-platform/internal/domain/entity"
	"drill-platform/internal/pkg/flowengine"
)

func TestAdvanceRunningDrillFromTerminalStepsStartsSamePhasePendingTask(t *testing.T) {
	engine := flowengine.NewEngine()
	flowDef := &flowengine.FlowDef{
		ID:   49,
		Name: "same-phase-reconcile-flow",
		Steps: []*flowengine.StepDef{
			{ID: 100, Name: "phase", Seq: 1, StepType: flowengine.StepTypeSerial},
			{ID: 110, Name: "section", Seq: 2, StepType: flowengine.StepTypeSerial, ParentStepID: 100},
			{ID: 120, Name: "completed task", Seq: 3, StepType: flowengine.StepTypeSerial, ParentStepID: 110},
			{ID: 121, Name: "completed child", Seq: 4, StepType: flowengine.StepTypeSerial, ParentStepID: 120},
			{ID: 130, Name: "next task", Seq: 5, StepType: flowengine.StepTypeSerial, ParentStepID: 110, PreStepIDs: []int64{120}},
			{ID: 131, Name: "next child", Seq: 6, StepType: flowengine.StepTypeSerial, ParentStepID: 130, PreStepIDs: []int64{120}},
			{ID: 200, Name: "next phase", Seq: 7, StepType: flowengine.StepTypeSerial, PreStepIDs: []int64{100}},
			{ID: 210, Name: "next phase child", Seq: 8, StepType: flowengine.StepTypeSerial, ParentStepID: 200, PreStepIDs: []int64{100}},
		},
	}
	inst, err := engine.CreateInstance(flowDef, nil, 1)
	if err != nil {
		t.Fatalf("CreateInstance error: %v", err)
	}
	inst.Status = flowengine.FlowStatusRunning
	inst.Steps[100].Status = flowengine.StepStatusRunning
	inst.Steps[110].Status = flowengine.StepStatusRunning
	inst.Steps[120].Status = flowengine.StepStatusCompleted
	inst.Steps[121].Status = flowengine.StepStatusCompleted

	svc := &DrillService{engine: engine}
	steps := []entity.StepInstance{
		{ID: 1, DrillInstanceID: 49, StepTemplateID: 121, Status: "completed"},
		{ID: 2, DrillInstanceID: 49, StepTemplateID: 120, Status: "completed"},
		{ID: 3, DrillInstanceID: 49, StepTemplateID: 100, Status: "completed"},
	}

	advanced := svc.advanceRunningDrillFromTerminalSteps(49, steps)

	if !advanced {
		t.Fatalf("expected same-phase pending task to be advanced")
	}
	if inst.Steps[130].Status != flowengine.StepStatusRunning {
		t.Fatalf("expected same-phase next task running, got %s", inst.Steps[130].Status)
	}
	if inst.Steps[131].Status != flowengine.StepStatusRunning {
		t.Fatalf("expected same-phase next child running, got %s", inst.Steps[131].Status)
	}
	if inst.Steps[200].Status != flowengine.StepStatusPending {
		t.Fatalf("expected next phase to remain pending, got %s", inst.Steps[200].Status)
	}
	if inst.Steps[210].Status != flowengine.StepStatusPending {
		t.Fatalf("expected next phase child to remain pending, got %s", inst.Steps[210].Status)
	}
}
