package entity

import "time"

// WorkerEpoch is the singleton row that tracks the currently elected worker
// and a monotonically increasing epoch counter. Each leadership transition
// bumps the epoch and transfers worker_id, invalidating any stale command
// ownership stamped with the previous epoch.
type WorkerEpoch struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement:false;column:id" json:"id"`
	WorkerID   string     `gorm:"type:varchar(128);not null;column:worker_id" json:"worker_id"`
	Epoch      uint64     `gorm:"not null;default:0;column:epoch" json:"epoch"`
	LeaseUntil *time.Time `gorm:"column:lease_until" json:"lease_until,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (WorkerEpoch) TableName() string {
	return "drill_worker_epoch"
}
