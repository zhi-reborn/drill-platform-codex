package flowengine

import (
	"sync"
	"time"
)

type FlowStatus string

const (
	FlowStatusPending    FlowStatus = "pending"
	FlowStatusRunning    FlowStatus = "running"
	FlowStatusPaused     FlowStatus = "paused"
	FlowStatusCompleted  FlowStatus = "completed"
	FlowStatusTerminated FlowStatus = "terminated"
	FlowStatusIssue      FlowStatus = "issue"
)

type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusTimeout   StepStatus = "timeout"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusIssue     StepStatus = "issue"
)

type StepType string

const (
	StepTypeSerial    StepType = "serial"
	StepTypeParallel  StepType = "parallel"
	StepTypeAnyOf     StepType = "any_of"
	StepTypeCondition StepType = "condition"
)

type InterveneAction string

const (
	ActionPause        InterveneAction = "pause"
	ActionResume       InterveneAction = "resume"
	ActionTerminate    InterveneAction = "terminate"
	ActionSkip         InterveneAction = "skip"
	ActionForceComplete InterveneAction = "force_complete"
	ActionResumeTask   InterveneAction = "resume_task"
)

type EventType string

const (
	EventFlowStart       EventType = "flow_start"
	EventFlowComplete    EventType = "flow_complete"
	EventFlowPause       EventType = "flow_pause"
	EventFlowResume      EventType = "flow_resume"
	EventFlowTerminate   EventType = "flow_terminate"
	EventStepStart       EventType = "step_start"
	EventStepComplete    EventType = "step_complete"
	EventStepTimeout     EventType = "step_timeout"
	EventStepIssue       EventType = "step_issue"
	EventStepSkip        EventType = "step_skip"
	EventStepForceComplete EventType = "step_force_complete"
)

type ConditionResult string

const (
	ConditionPass ConditionResult = "pass"
	ConditionFail ConditionResult = "fail"
)

type FlowDef struct {
	ID    int64      `json:"id"`
	Name  string     `json:"name"`
	Steps []*StepDef `json:"steps"`
}

type StepDef struct {
	ID                       int64         `json:"id"`
	Name                     string        `json:"name"`
	Seq                      int           `json:"seq"`
	StepType                 StepType      `json:"step_type"`
	TimeoutMinutes           int           `json:"timeout_minutes"`
	PreStepIDs               []int64       `json:"pre_step_ids"`
	GuideContent             string        `json:"guide_content"`
	IsBlocking               bool          `json:"is_blocking"`
	DefaultAssigneeRole      string        `json:"default_assignee_role"`
	Condition                *ConditionDef `json:"condition,omitempty"`
	ParentStepID             int64         `json:"parent_step_id"`
	Phase                    string        `json:"phase"`
	PhaseStep                string        `json:"phase_step"`
	ExecutionMode            string        `json:"execution_mode"`
	EstimatedDurationMinutes *int           `json:"estimated_duration_minutes,omitempty"`
	EstimatedStartOffset     *int           `json:"estimated_start_offset,omitempty"`
	JSONAttributes           string         `json:"attributes"`
}

type ConditionDef struct {
	Expression   string  `json:"expression"`
	TrueStepIDs  []int64 `json:"true_step_ids"`
	FalseStepIDs []int64 `json:"false_step_ids"`
}

type FlowInst struct {
	mu             sync.RWMutex
	ID             int64               `json:"id"`
	FlowDefID      int64               `json:"flow_def_id"`
	Name           string              `json:"name"`
	Status         FlowStatus          `json:"status"`
	StartTime      *time.Time          `json:"start_time,omitempty"`
	EndTime        *time.Time          `json:"end_time,omitempty"`
	CurrentStepIDs []int64             `json:"current_step_ids"`
	ProgressPct    int                 `json:"progress_pct"`
	CreatedBy      int64               `json:"created_by"`
	CreatedAt      time.Time           `json:"created_at"`
	Steps          map[int64]*StepInst `json:"steps"`
	Assignees      map[int64][]int64   `json:"-"`
}

type StepInst struct {
	mu                       sync.RWMutex
	ID                       int64           `json:"id"`
	StepDefID                int64           `json:"step_def_id"`
	Name                     string          `json:"name"`
	Seq                      int             `json:"seq"`
	Status                   StepStatus      `json:"status"`
	AssigneeIDs              []int64         `json:"assignee_ids"`
	ActualOperator           *int64          `json:"actual_operator,omitempty"`
	StartTime                *time.Time      `json:"start_time,omitempty"`
	EndTime                  *time.Time      `json:"end_time,omitempty"`
	TimeoutAt                *time.Time      `json:"timeout_at,omitempty"`
	Remark                   string          `json:"remark,omitempty"`
	IssueDesc                string          `json:"issue_desc,omitempty"`
	ConditionResult          ConditionResult `json:"condition_result,omitempty"`
	ParentStepID             int64           `json:"parent_step_id"`
	Phase                    string          `json:"phase"`
	PhaseStep                string          `json:"phase_step"`
	ExecutionMode            string          `json:"execution_mode"`
	EstimatedDurationMinutes *int            `json:"estimated_duration_minutes,omitempty"`
	EstimatedStartOffset     *int            `json:"estimated_start_offset,omitempty"`
	JSONAttributes           string          `json:"attributes"`
}

type Event struct {
	Type       EventType   `json:"event_type"`
	FlowInstID int64       `json:"flow_inst_id"`
	StepInstID int64       `json:"step_inst_id,omitempty"`
	StepDefID  int64       `json:"step_def_id,omitempty"`
	Payload    interface{} `json:"payload,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

type PersistenceCallbacks interface {
	OnFlowStatusChanged(flowInstID int64, oldStatus, newStatus FlowStatus)
	OnStepStatusChanged(stepInstID int64, oldStatus, newStatus StepStatus)
	OnStepStarted(stepInstID int64, timeoutAt time.Time)
	OnStepCompleted(stepInstID int64, operatorID int64, remark string)
	OnStepIssue(stepInstID int64, operatorID int64, issueDesc string)
	LogAction(stepInstID int64, action string, operatorID int64, content string)
}

type StepLoader interface {
	GetStepDef(flowDefID, stepDefID int64) (*StepDef, error)
	GetAllStepDefs(flowDefID int64) ([]*StepDef, error)
	GetCurrentStepStatus(flowInstID int64, stepDefID int64) (*StepInst, error)
}
