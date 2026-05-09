package entity

import "time"

// StepInstanceLog 步骤操作日志
type StepInstanceLog struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StepInstanceID  uint64    `gorm:"type:bigint unsigned;not null;column:step_instance_id;index:idx_step_instance" json:"step_instance_id"`
	Action          string    `gorm:"type:varchar(32);not null;column:action" json:"action"`
	OperatorID      uint64    `gorm:"type:bigint unsigned;not null;column:operator_id" json:"operator_id"`
	OperatorName    string    `gorm:"type:varchar(64);not null;column:operator_name" json:"operator_name"`
	Content         string    `gorm:"type:text;column:content" json:"content"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	StepInstance StepInstance `gorm:"foreignKey:StepInstanceID;references:ID" json:"step_instance,omitempty"`
}

func (StepInstanceLog) TableName() string {
	return "step_instance_log"
}
