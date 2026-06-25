package entity

import "time"

// DrillInstanceLog 演练操作日志（统一使用 drill_instance_step_log 表）
// drill_instance_id 有值、step_instance_id 为空 = 演练级别日志
// drill_instance_id 有值、step_instance_id 有值 = 步骤级别日志
type DrillInstanceLog struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	DrillInstanceID uint64    `gorm:"type:bigint unsigned;not null;column:drill_instance_id;index:idx_drill_instance" json:"drill_instance_id"`
	StepInstanceID  *uint64   `gorm:"type:bigint unsigned;column:task_instance_id;index:idx_step_instance;uniqueIndex:uk_log_command_action_step,priority:3" json:"step_instance_id"`
	CommandID       *uint64   `gorm:"type:bigint unsigned;column:command_id;uniqueIndex:uk_log_command_action_step,priority:1" json:"command_id,omitempty"`
	Action          string    `gorm:"type:varchar(32);not null;column:action;uniqueIndex:uk_log_command_action_step,priority:2" json:"action"`
	OperatorID      uint64    `gorm:"type:bigint unsigned;not null;column:operator_id" json:"operator_id"`
	OperatorName    string    `gorm:"type:varchar(64);not null;column:operator_name" json:"operator_name"`
	Content         string    `gorm:"type:text;column:content" json:"content"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	DrillInstance DrillInstance `gorm:"foreignKey:DrillInstanceID;references:ID" json:"drill_instance,omitempty"`
	StepInstance  StepInstance  `gorm:"foreignKey:StepInstanceID;references:ID" json:"step_instance,omitempty"`
}

func (DrillInstanceLog) TableName() string {
	return "drill_instance_step_log"
}
