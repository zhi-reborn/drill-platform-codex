package entity

import "time"

type NotificationType string

const (
	NotificationTypeTaskAssigned    NotificationType = "task_assigned"
	NotificationTypeStepComplete    NotificationType = "step_complete"
	NotificationTypeStepTimeout     NotificationType = "step_timeout"
	NotificationTypeDrillStarted    NotificationType = "drill_started"
	NotificationTypeDrillPaused     NotificationType = "drill_paused"
	NotificationTypeDrillResumed    NotificationType = "drill_resumed"
	NotificationTypeDrillCompleted  NotificationType = "drill_completed"
	NotificationTypeDrillTerminated NotificationType = "drill_terminated"
	NotificationTypeSystemAlert     NotificationType = "system_alert"
)

type Notification struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    uint64         `gorm:"type:bigint unsigned;not null;column:user_id;index:idx_user_created" json:"user_id"`
	Type      NotificationType `gorm:"type:varchar(50);not null;column:type" json:"type"`
	Title     string         `gorm:"type:varchar(200);not null;column:title" json:"title"`
	Content   string         `gorm:"type:text;column:content" json:"content"`
	DrillID   *uint64        `gorm:"type:bigint unsigned;column:drill_id" json:"drill_id,omitempty"`
	DrillName *string        `gorm:"type:varchar(200);column:drill_name" json:"drill_name,omitempty"`
	StepID    *uint64        `gorm:"type:bigint unsigned;column:step_id" json:"step_id,omitempty"`
	StepName  *string        `gorm:"type:varchar(200);column:step_name" json:"step_name,omitempty"`
	IsRead    bool           `gorm:"type:tinyint;not null;default:0;column:is_read;index:idx_user_unread" json:"is_read"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;index:idx_user_created" json:"created_at"`
}

func (Notification) TableName() string {
	return "notification"
}
