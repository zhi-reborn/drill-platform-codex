package flowengine

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var _ PersistenceCallbacks = (*testCallbacks)(nil)

type testCallbacks struct {
	mu       sync.Mutex
	flowLog  []string
	stepLog  []string
	actionLog []string
}

func (c *testCallbacks) OnFlowStatusChanged(flowInstID int64, oldStatus, newStatus FlowStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.flowLog = append(c.flowLog, fmt.Sprintf("flow %d: %s -> %s", flowInstID, oldStatus, newStatus))
}

func (c *testCallbacks) OnStepStatusChanged(stepInstID int64, oldStatus, newStatus StepStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stepLog = append(c.stepLog, fmt.Sprintf("step %d: %s -> %s", stepInstID, oldStatus, newStatus))
}

func (c *testCallbacks) OnStepStarted(stepInstID int64, timeoutAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stepLog = append(c.stepLog, fmt.Sprintf("step %d: started (timeout: %v)", stepInstID, timeoutAt))
}

func (c *testCallbacks) OnStepCompleted(stepInstID int64, operatorID int64, remark string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.actionLog = append(c.actionLog, fmt.Sprintf("step %d: completed by %d", stepInstID, operatorID))
}

func (c *testCallbacks) OnStepIssue(stepInstID int64, operatorID int64, issueDesc string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.actionLog = append(c.actionLog, fmt.Sprintf("step %d: issue by %d: %s", stepInstID, operatorID, issueDesc))
}

func (c *testCallbacks) LogAction(stepInstID int64, action string, operatorID int64, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.actionLog = append(c.actionLog, fmt.Sprintf("step %d: %s by %d: %s", stepInstID, action, operatorID, content))
}

type testStepLoader struct {
	steps map[int64]*StepDef
}

func (l *testStepLoader) GetStepDef(flowDefID, stepDefID int64) (*StepDef, error) {
	if sd, ok := l.steps[stepDefID]; ok {
		return sd, nil
	}
	return nil, fmt.Errorf("step %d not found", stepDefID)
}

func (l *testStepLoader) GetAllStepDefs(flowDefID int64) ([]*StepDef, error) {
	var result []*StepDef
	for _, sd := range l.steps {
		result = append(result, sd)
	}
	return result, nil
}

func (l *testStepLoader) GetCurrentStepStatus(flowInstID int64, stepDefID int64) (*StepInst, error) {
	return nil, nil
}

func newTestEngine() (*Engine, *testCallbacks) {
	cb := &testCallbacks{}
	e := NewEngine()
	e.SetCallbacks(cb)
	return e, cb
}

func newSerialFlowDef() *FlowDef {
	return &FlowDef{
		ID:   1,
		Name: "serial-flow",
		Steps: []*StepDef{
			{ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			{ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
			{ID: 103, Name: "step3", Seq: 3, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{102}},
		},
	}
}

func TestCreateInstance(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	assignees := map[int64][]int64{101: {1, 2}, 102: {3}}

	inst, err := e.CreateInstance(flowDef, assignees, 1)
	if err != nil {
		t.Fatalf("CreateInstance error: %v", err)
	}

	if inst.Status != FlowStatusPending {
		t.Errorf("expected status pending, got %s", inst.Status)
	}

	if len(inst.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(inst.Steps))
	}

	if len(inst.Steps[101].AssigneeIDs) != 2 {
		t.Errorf("expected 2 assignees for step 101, got %d", len(inst.Steps[101].AssigneeIDs))
	}
}

func TestStartFlow(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			102: {ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
			103: {ID: 103, Name: "step3", Seq: 3, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{102}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	err := e.Start(inst.ID)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	if inst.Status != FlowStatusRunning {
		t.Errorf("expected status running, got %s", inst.Status)
	}

	if len(inst.CurrentStepIDs) != 1 || inst.CurrentStepIDs[0] != 101 {
		t.Errorf("expected current step 101, got %v", inst.CurrentStepIDs)
	}

	if inst.Steps[101].Status != StepStatusRunning {
		t.Errorf("expected step1 status running, got %s", inst.Steps[101].Status)
	}
}

func TestCompleteStep_Serial(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			102: {ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
			103: {ID: 103, Name: "step3", Seq: 3, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{102}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	err := e.CompleteStep(inst.ID, 101, 1, "done")
	if err != nil {
		t.Fatalf("CompleteStep error: %v", err)
	}

	if inst.Steps[101].Status != StepStatusCompleted {
		t.Errorf("expected step1 completed, got %s", inst.Steps[101].Status)
	}

	if inst.Steps[102].Status != StepStatusRunning {
		t.Errorf("expected step2 running, got %s", inst.Steps[102].Status)
	}

	if inst.ProgressPct != 33 {
		t.Errorf("expected progress 33, got %d", inst.ProgressPct)
	}
}

func TestCompleteStep_AllSteps(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			102: {ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
			103: {ID: 103, Name: "step3", Seq: 3, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{102}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	e.CompleteStep(inst.ID, 101, 1, "")
	e.CompleteStep(inst.ID, 102, 2, "")
	e.CompleteStep(inst.ID, 103, 3, "")

	if inst.Status != FlowStatusCompleted {
		t.Errorf("expected flow completed, got %s", inst.Status)
	}

	if inst.ProgressPct != 100 {
		t.Errorf("expected progress 100, got %d", inst.ProgressPct)
	}
}

func TestReportIssue(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			102: {ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
			103: {ID: 103, Name: "step3", Seq: 3, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{102}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	err := e.ReportIssue(inst.ID, 101, 1, "cannot connect")
	if err != nil {
		t.Fatalf("ReportIssue error: %v", err)
	}

	if inst.Steps[101].Status != StepStatusIssue {
		t.Errorf("expected step1 issue, got %s", inst.Steps[101].Status)
	}

	if inst.Status != FlowStatusIssue {
		t.Errorf("expected flow issue, got %s", inst.Status)
	}
}

func TestIntervene_PauseResume(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	err := e.Intervene(inst.ID, ActionPause, nil, 1)
	if err != nil {
		t.Fatalf("Intervene pause error: %v", err)
	}

	if inst.Status != FlowStatusPaused {
		t.Errorf("expected paused, got %s", inst.Status)
	}

	err = e.Intervene(inst.ID, ActionResume, nil, 1)
	if err != nil {
		t.Fatalf("Intervene resume error: %v", err)
	}

	if inst.Status != FlowStatusRunning {
		t.Errorf("expected running after resume, got %s", inst.Status)
	}
}

func TestIntervene_Terminate(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	err := e.Intervene(inst.ID, ActionTerminate, nil, 1)
	if err != nil {
		t.Fatalf("Intervene terminate error: %v", err)
	}

	if inst.Status != FlowStatusTerminated {
		t.Errorf("expected terminated, got %s", inst.Status)
	}

	if inst.EndTime == nil {
		t.Error("expected end time to be set")
	}
}

func TestIntervene_SkipStep(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			102: {ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	target := int64(102)
	err := e.Intervene(inst.ID, ActionSkip, &target, 1)
	if err != nil {
		t.Fatalf("Intervene skip error: %v", err)
	}

	if inst.Steps[102].Status != StepStatusSkipped {
		t.Errorf("expected step2 skipped, got %s", inst.Steps[102].Status)
	}
}

func TestIntervene_ForceComplete(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			102: {ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
			103: {ID: 103, Name: "step3", Seq: 3, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{102}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	target := int64(102)
	err := e.Intervene(inst.ID, ActionForceComplete, &target, 1)
	if err != nil {
		t.Fatalf("Intervene force complete error: %v", err)
	}

	if inst.Steps[102].Status != StepStatusCompleted {
		t.Errorf("expected step2 completed, got %s", inst.Steps[102].Status)
	}

	if inst.Steps[103].Status != StepStatusRunning {
		t.Errorf("expected step3 running, got %s", inst.Steps[103].Status)
	}
}

func TestGetInstance_NotFound(t *testing.T) {
	e, _ := newTestEngine()
	_, ok := e.GetInstance(999)
	if !ok {
		t.Log("Instance 999 not found as expected")
	}
}

func TestStart_NonPending(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := newSerialFlowDef()
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	err := e.Start(inst.ID)
	if err == nil {
		t.Error("expected error when starting already running instance")
	}
}

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	eb := NewEventBus()
	ch := make(chan Event, 10)

	eb.Subscribe(EventStepComplete, ch)

	eb.emit(EventStepComplete, 1, 101, 101, "test")

	select {
	case evt := <-ch:
		if evt.Type != EventStepComplete {
			t.Errorf("expected event type %s, got %s", EventStepComplete, evt.Type)
		}
	default:
		t.Error("expected event in channel")
	}
}

func TestTimeoutScheduler(t *testing.T) {
	eb := NewEventBus()
	ts := NewTimeoutScheduler(eb)
	ts.Start()
	defer ts.Stop()

	eventCh := make(chan Event, 10)
	eb.Subscribe(EventStepTimeout, eventCh)

	ts.Register(1, 101, 101, time.Now().Add(50*time.Millisecond))

	select {
	case evt := <-eventCh:
		if evt.Type != EventStepTimeout {
			t.Errorf("expected timeout event, got %s", evt.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout event not received within 2s")
	}
}

func TestErrorCases(t *testing.T) {
	e, _ := newTestEngine()

	err := e.Start(999)
	if err != ErrInstanceNotFound {
		t.Errorf("expected ErrInstanceNotFound, got %v", err)
	}

	err = e.CompleteStep(999, 1, 1, "")
	if err != ErrInstanceNotFound {
		t.Errorf("expected ErrInstanceNotFound, got %v", err)
	}

	err = e.ReportIssue(999, 1, 1, "")
	if err != ErrInstanceNotFound {
		t.Errorf("expected ErrInstanceNotFound, got %v", err)
	}

	emptyDef := &FlowDef{ID: 2, Name: "empty"}
	_, err = e.CreateInstance(emptyDef, nil, 1)
	if err != ErrInvalidFlowDef {
		t.Errorf("expected ErrInvalidFlowDef, got %v", err)
	}
}

func TestSkipAndComplete_ShouldReach100Progress(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := &FlowDef{
		ID:   1,
		Name: "two-step-flow",
		Steps: []*StepDef{
			{ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			{ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
		},
	}
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			102: {ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	target := int64(101)
	if err := e.Intervene(inst.ID, ActionSkip, &target, 1); err != nil {
		t.Fatalf("Intervene skip error: %v", err)
	}

	if inst.Steps[101].Status != StepStatusSkipped {
		t.Errorf("expected step1 skipped, got %s", inst.Steps[101].Status)
	}

	if inst.ProgressPct != 50 {
		t.Errorf("expected progress 50 after skip, got %d", inst.ProgressPct)
	}

	if err := e.CompleteStep(inst.ID, 102, 2, ""); err != nil {
		t.Fatalf("CompleteStep error: %v", err)
	}

	if inst.Status != FlowStatusCompleted {
		t.Errorf("expected flow completed, got %s", inst.Status)
	}

	if inst.ProgressPct != 100 {
		t.Errorf("expected progress 100, got %d", inst.ProgressPct)
	}
}

func TestForceComplete_ShouldAdvanceToNextStep(t *testing.T) {
	e, _ := newTestEngine()
	flowDef := &FlowDef{
		ID:   1,
		Name: "two-step-flow",
		Steps: []*StepDef{
			{ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			{ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
		},
	}
	loader := &testStepLoader{
		steps: map[int64]*StepDef{
			101: {ID: 101, Name: "step1", Seq: 1, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{}},
			102: {ID: 102, Name: "step2", Seq: 2, StepType: StepTypeSerial, TimeoutMinutes: 5, PreStepIDs: []int64{101}},
		},
	}
	e.SetStepLoader(loader)

	inst, _ := e.CreateInstance(flowDef, nil, 1)
	e.Start(inst.ID)

	if inst.Steps[101].Status != StepStatusRunning {
		t.Fatalf("expected step1 running, got %s", inst.Steps[101].Status)
	}

	target := int64(101)
	err := e.Intervene(inst.ID, ActionForceComplete, &target, 1)
	if err != nil {
		t.Fatalf("Intervene force complete error: %v", err)
	}

	if inst.Steps[101].Status != StepStatusCompleted {
		t.Errorf("expected step1 completed, got %s", inst.Steps[101].Status)
	}

	if inst.Steps[102].Status != StepStatusRunning {
		t.Errorf("expected step2 running after force complete, got %s", inst.Steps[102].Status)
	}

	if inst.ProgressPct != 50 {
		t.Errorf("expected progress 50, got %d", inst.ProgressPct)
	}
}
