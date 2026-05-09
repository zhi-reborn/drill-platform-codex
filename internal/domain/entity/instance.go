package entity

import "time"

// DrillInstance 演练实例
type DrillInstance struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TemplateID    uint64     `gorm:"type:bigint unsigned;not null;column:template_id;index:idx_template_id" json:"template_id"`
	Name          string     `gorm:"type:varchar(128);not null;column:name" json:"name"`
	Status        string     `gorm:"type:varchar(32);not null;default:pending;column:status;index:idx_status" json:"status"`
	StartTime     *time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime       *time.Time `gorm:"column:end_time" json:"end_time"`
	PlannedStart  *time.Time `gorm:"column:planned_start" json:"planned_start"`
	CurrentStepID *uint64    `gorm:"type:bigint unsigned;column:current_step_id" json:"current_step_id"`
	ProgressPct   int        `gorm:"type:int;not null;default:0;column:progress_pct" json:"progress_pct"`
	CreatedBy     uint64     `gorm:"type:bigint unsigned;not null;column:created_by" json:"created_by"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime;index:idx_created_at" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Template DrillTemplate `gorm:"foreignKey:TemplateID;references:ID" json:"template,omitempty"`
	Steps    []StepInstance `gorm:"foreignKey:DrillInstanceID" json:"steps,omitempty"`
}

func (DrillInstance) TableName() string {
	return "drill_instance"
}

// StepInstance 步骤实例
type StepInstance struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	DrillInstanceID uint64     `gorm:"type:bigint unsigned;not null;column:drill_instance_id;index:idx_drill_step" json:"drill_instance_id"`
	StepTemplateID  uint64     `gorm:"type:bigint unsigned;not null;column:step_template_id" json:"step_template_id"`
	Name            string     `gorm:"type:varchar(128);not null;column:name" json:"name"`
	Seq             int        `gorm:"type:int;not null;column:seq" json:"seq"`
	Status          string     `gorm:"type:varchar(32);not null;default:pending;column:status;index:idx_drill_step" json:"status"`
	AssigneeIDs     string     `gorm:"type:json;not null;column:assignee_ids" json:"assignee_ids"`
	ActualOperator  *uint64    `gorm:"type:bigint unsigned;column:actual_operator" json:"actual_operator"`
	StartTime       *time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime         *time.Time `gorm:"column:end_time" json:"end_time"`
	TimeoutAt       *time.Time `gorm:"column:timeout_at" json:"timeout_at"`
	Remark          string     `gorm:"type:text;column:remark" json:"remark"`
	IssueDesc       string     `gorm:"type:text;column:issue_desc" json:"issue_desc"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	DrillInstance DrillInstance `gorm:"foreignKey:DrillInstanceID;references:ID" json:"drill_instance,omitempty"`
	Logs          []StepInstanceLog `gorm:"foreignKey:StepInstanceID;constraint:OnDelete:CASCADE" json:"logs,omitempty"`
}

func (StepInstance) TableName() string {
	return "step_instance"
}
