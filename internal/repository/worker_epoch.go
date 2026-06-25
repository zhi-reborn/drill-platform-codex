package repository

import (
	"context"
	"errors"
	"time"

	"drill-platform/internal/domain/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrOwnershipLost is returned by fenced repository methods when the WHERE
// clause matches zero rows, meaning the caller no longer owns the command
// (stale epoch, mismatched token, expired lease, or command re-claimed by
// another worker).
var ErrOwnershipLost = errors.New("command ownership lost")

// CommandOwnership is the value object carried by fenced command mutations.
// It must match the worker_id, worker_epoch, and lease_token stamped on the
// FlowCommand at claim time; otherwise the fenced update affects zero rows.
type CommandOwnership struct {
	WorkerID string
	Epoch    uint64
	Token    string
}

// WorkerEpochRepo persists the singleton worker epoch row used to fence
// command ownership across leadership transitions.
type WorkerEpochRepo struct{ db *gorm.DB }

func NewWorkerEpochRepo(db ...*gorm.DB) *WorkerEpochRepo {
	if len(db) > 0 && db[0] != nil {
		return &WorkerEpochRepo{db: db[0]}
	}
	return &WorkerEpochRepo{db: DB}
}

// AdvanceEpoch atomically increments the singleton epoch, transfers
// worker_id to the calling worker, and sets lease_until = now + leaseTTL.
// The singleton row is created on first call with epoch = 1.
func (r *WorkerEpochRepo) AdvanceEpoch(ctx context.Context, workerID string, leaseTTL time.Duration) (*entity.WorkerEpoch, error) {
	var result *entity.WorkerEpoch
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.WorkerEpoch
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", 1).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			leaseUntil := time.Now().Add(leaseTTL)
			row := &entity.WorkerEpoch{
				ID:         1,
				WorkerID:   workerID,
				Epoch:      1,
				LeaseUntil: &leaseUntil,
			}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
			result = row
			return nil
		}
		if err != nil {
			return err
		}

		newEpoch := existing.Epoch + 1
		leaseUntil := time.Now().Add(leaseTTL)
		if err := tx.Model(&entity.WorkerEpoch{}).Where("id = ?", 1).Updates(map[string]any{
			"epoch":       newEpoch,
			"worker_id":   workerID,
			"lease_until": leaseUntil,
		}).Error; err != nil {
			return err
		}
		existing.Epoch = newEpoch
		existing.WorkerID = workerID
		existing.LeaseUntil = &leaseUntil
		result = &existing
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CurrentEpoch reads the singleton row. It returns the zero value when the
// row does not exist yet, which callers interpret as "no worker has
// registered an epoch".
func (r *WorkerEpochRepo) CurrentEpoch(ctx context.Context) (*entity.WorkerEpoch, error) {
	var row entity.WorkerEpoch
	err := r.db.WithContext(ctx).Where("id = ?", 1).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
