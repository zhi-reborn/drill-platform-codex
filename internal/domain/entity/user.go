package entity

import "time"

// User 系统用户
type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username     string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_username;column:username" json:"username"`
	RealName     string    `gorm:"type:varchar(64);not null;column:real_name" json:"real_name"`
	PasswordHash string    `gorm:"type:varchar(256);not null;column:password_hash" json:"-"`
	Email        string    `gorm:"type:varchar(128);column:email" json:"email"`
	Role         string    `gorm:"type:varchar(32);not null;column:role" json:"role"`
	Department   string    `gorm:"type:varchar(64);column:department" json:"department"`
	Phone        string    `gorm:"type:varchar(20);column:phone" json:"phone"`
	Status       int8      `gorm:"type:tinyint;not null;default:1;column:status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "user"
}

// DrillAssignee 演练人员分配
type DrillAssignee struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	DrillInstanceID uint64 `gorm:"type:bigint unsigned;not null;column:drill_instance_id;uniqueIndex:uk_drill_step_user" json:"drill_instance_id"`
	StepInstanceID  uint64 `gorm:"type:bigint unsigned;not null;column:step_instance_id;uniqueIndex:uk_drill_step_user" json:"step_instance_id"`
	UserID          uint64 `gorm:"type:bigint unsigned;not null;column:user_id;uniqueIndex:uk_drill_step_user" json:"user_id"`
	NotifySent      int8   `gorm:"type:tinyint;not null;default:0;column:notify_sent" json:"notify_sent"`

	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (DrillAssignee) TableName() string {
	return "drill_assignee"
}
