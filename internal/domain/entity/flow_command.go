package entity

import "time"

type FlowCommandStatus string

const (
	FlowCommandPending    FlowCommandStatus = "pending"
	FlowCommandProcessing FlowCommandStatus = "processing"
	FlowCommandSucceeded  FlowCommandStatus = "succeeded"
	FlowCommandFailed     FlowCommandStatus = "failed"
)

type FlowCommand struct {
	ID              uint64            `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CommandType     string            `gorm:"type:varchar(64);not null;column:command_type" json:"command_type"`
	DrillInstanceID uint64            `gorm:"type:bigint unsigned;not null;column:drill_instance_id" json:"drill_instance_id"`
	StepInstanceID  *uint64           `gorm:"type:bigint unsigned;column:step_instance_id" json:"step_instance_id,omitempty"`
	OperatorID      uint64            `gorm:"type:bigint unsigned;not null;column:operator_id" json:"operator_id"`
	IdempotencyKey  string            `gorm:"type:varchar(128);not null;uniqueIndex:uk_flow_command_idempotency;column:idempotency_key" json:"idempotency_key"`
	Payload         string            `gorm:"type:json;not null;column:payload" json:"payload"`
	Status          FlowCommandStatus `gorm:"type:varchar(20);not null;index:idx_flow_command_pending,priority:1;column:status" json:"status"`
	WorkerID        *string           `gorm:"type:varchar(128);column:worker_id" json:"worker_id,omitempty"`
	WorkerEpoch     uint64            `gorm:"not null;default:0;column:worker_epoch" json:"worker_epoch"`
	LeaseToken      string            `gorm:"type:varchar(128);not null;default:'';column:lease_token" json:"lease_token"`
	LeaseUntil      *time.Time        `gorm:"column:lease_until;index:idx_flow_command_lease,priority:2" json:"lease_until,omitempty"`
	Attempts        int               `gorm:"not null;default:0;column:attempts" json:"attempts"`
	AttemptCount    int               `gorm:"not null;default:0;column:attempt_count" json:"attempt_count"`
	Result          *string           `gorm:"type:json;column:result" json:"result,omitempty"`
	ErrorCode       *string           `gorm:"type:varchar(64);column:error_code" json:"error_code,omitempty"`
	ErrorMessage    *string           `gorm:"type:varchar(500);column:error_message" json:"error_message,omitempty"`
	CreatedAt       time.Time         `gorm:"column:created_at;autoCreateTime;index:idx_flow_command_pending,priority:2" json:"created_at"`
	StartedAt       *time.Time        `gorm:"column:started_at" json:"started_at,omitempty"`
	FinishedAt      *time.Time        `gorm:"column:finished_at" json:"finished_at,omitempty"`
	UpdatedAt       time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (FlowCommand) TableName() string {
	return "drill_flow_command"
}

func (c FlowCommand) IsTerminal() bool {
	return c.Status == FlowCommandSucceeded || c.Status == FlowCommandFailed
}
