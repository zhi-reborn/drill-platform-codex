package entity

import "time"

// DrillTemplate 演练模板
type DrillTemplate struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string    `gorm:"type:varchar(128);not null;column:name" json:"name"`
	Category    string    `gorm:"type:varchar(64);not null;column:category;index:idx_category" json:"category"`
	Description string    `gorm:"type:text;column:description" json:"description"`
	Status      int8      `gorm:"type:tinyint;not null;default:1;column:status" json:"status"`
	CreatedBy   uint64    `gorm:"type:bigint unsigned;not null;column:created_by" json:"created_by"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`

	Steps []StepTemplate `gorm:"foreignKey:DrillTemplateID;constraint:OnDelete:CASCADE" json:"steps,omitempty"`
}

func (DrillTemplate) TableName() string {
	return "drill_template"
}

// StepTemplate 步骤模板
type StepTemplate struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	DrillTemplateID    uint64    `gorm:"type:bigint unsigned;not null;column:drill_template_id;index:idx_drill_template_id" json:"drill_template_id"`
	Name               string    `gorm:"type:varchar(128);not null;column:name" json:"name"`
	Seq                int       `gorm:"type:int;not null;column:seq" json:"seq"`
	StepType           string    `gorm:"type:varchar(32);not null;column:step_type" json:"step_type"`
	TimeoutMinutes     int       `gorm:"type:int;not null;default:5;column:timeout_minutes" json:"timeout_minutes"`
	PreStepIDs         string    `gorm:"type:json;column:pre_step_ids" json:"pre_step_ids"`
	GuideContent       string    `gorm:"type:text;column:guide_content" json:"guide_content"`
	IsBlocking         int8      `gorm:"type:tinyint;not null;default:1;column:is_blocking" json:"is_blocking"`
	DefaultAssigneeRole string   `gorm:"type:varchar(64);column:default_assignee_role" json:"default_assignee_role"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	DrillTemplate DrillTemplate `gorm:"foreignKey:DrillTemplateID;references:ID" json:"drill_template,omitempty"`
}

func (StepTemplate) TableName() string {
	return "step_template"
}
