package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"drill-platform/internal/domain/entity"

	"gorm.io/gorm"
)

type FlowCommandRepo struct{ db *gorm.DB }

func NewFlowCommandRepo(db ...*gorm.DB) *FlowCommandRepo {
	if len(db) > 0 && db[0] != nil {
		return &FlowCommandRepo{db: db[0]}
	}
	return &FlowCommandRepo{db: DB}
}

func (r *FlowCommandRepo) CreateOrGet(cmd *entity.FlowCommand) (*entity.FlowCommand, bool, error) {
	if err := r.db.Create(cmd).Error; err != nil {
		if isDuplicateIdempotencyKeyError(err) {
			var existing entity.FlowCommand
			if findErr := r.db.Where("idempotency_key = ?", cmd.IdempotencyKey).First(&existing).Error; findErr != nil {
				return nil, false, findErr
			}
			return &existing, false, nil
		}
		return nil, false, err
	}
	return cmd, true, nil
}

func (r *FlowCommandRepo) FindByID(id uint64) (*entity.FlowCommand, error) {
	var cmd entity.FlowCommand
	if err := r.db.First(&cmd, id).Error; err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (r *FlowCommandRepo) FindByIDForOperator(id, operatorID uint64) (*entity.FlowCommand, error) {
	var cmd entity.FlowCommand
	if err := r.db.Where("id = ? AND operator_id = ?", id, operatorID).First(&cmd).Error; err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (r *FlowCommandRepo) MarkSucceeded(id uint64, result any) error {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	now := time.Now()
	return r.db.Model(&entity.FlowCommand{}).Where("id = ?", id).Updates(map[string]any{
		"status":        entity.FlowCommandSucceeded,
		"result":        string(resultBytes),
		"finished_at":   now,
		"error_code":    nil,
		"error_message": nil,
	}).Error
}

func (r *FlowCommandRepo) MarkFailed(id uint64, code, message string) error {
	now := time.Now()
	return r.db.Model(&entity.FlowCommand{}).Where("id = ?", id).Updates(map[string]any{
		"status":        entity.FlowCommandFailed,
		"error_code":    code,
		"error_message": message,
		"finished_at":   now,
	}).Error
}

func (r *FlowCommandRepo) RequeueExpired(now time.Time) (int64, error) {
	res := r.db.Model(&entity.FlowCommand{}).
		Where("status = ? AND lease_until IS NOT NULL AND lease_until <= ?", entity.FlowCommandProcessing, now).
		Updates(map[string]any{
			"status":      entity.FlowCommandPending,
			"worker_id":   nil,
			"lease_until": nil,
		})
	return res.RowsAffected, res.Error
}

func (r *FlowCommandRepo) ClaimNext(ctx context.Context, workerID string, lease time.Duration) (*entity.FlowCommand, error) {
	var claimed *entity.FlowCommand
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var id uint64
		row := tx.Raw("SELECT id FROM drill_flow_command WHERE status = ? ORDER BY created_at, id LIMIT 1 FOR UPDATE SKIP LOCKED", entity.FlowCommandPending).Row()
		if err := row.Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}

		now := time.Now()
		leaseUntil := now.Add(lease)
		if err := tx.Model(&entity.FlowCommand{}).Where("id = ?", id).Updates(map[string]any{
			"status":      entity.FlowCommandProcessing,
			"worker_id":   workerID,
			"attempts":    gorm.Expr("attempts + ?", 1),
			"lease_until": leaseUntil,
			"started_at":  now,
		}).Error; err != nil {
			return err
		}

		var cmd entity.FlowCommand
		if err := tx.First(&cmd, id).Error; err != nil {
			return err
		}
		claimed = &cmd
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

func (r *FlowCommandRepo) ExtendLease(ctx context.Context, id uint64, workerID string, until time.Time) (bool, error) {
	res := r.db.WithContext(ctx).Model(&entity.FlowCommand{}).
		Where("id = ? AND worker_id = ? AND status = ?", id, workerID, entity.FlowCommandProcessing).
		Update("lease_until", until)
	return res.RowsAffected == 1, res.Error
}

func isDuplicateIdempotencyKeyError(err error) bool {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "idempotency_key") || strings.Contains(msg, "uk_flow_command_idempotency") {
		return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique constraint") || strings.Contains(msg, "duplicated key")
	}
	return false
}
