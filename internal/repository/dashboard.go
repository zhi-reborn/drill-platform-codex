package repository

import (
	"context"
	"database/sql"
	"math"

	"drill-platform/internal/domain/entity"
)

type DashboardRepo struct{}

func NewDashboardRepo() *DashboardRepo {
	return &DashboardRepo{}
}

func (r *DashboardRepo) CountTeam(ctx context.Context, onlineIDs []uint64) (total, online int64, err error) {
	var counts struct {
		Total  int64
		Online int64
	}
	err = DB.WithContext(ctx).Model(&entity.User{}).
		Select("COUNT(*) AS total, COUNT(CASE WHEN id IN ? THEN 1 END) AS online", onlineIDs).
		Where("status = ?", 1).Scan(&counts).Error
	return counts.Total, counts.Online, err
}

func (r *DashboardRepo) CountMyTemplates(ctx context.Context, userID uint64) (int64, error) {
	var total int64
	err := DB.WithContext(ctx).Model(&entity.DrillTemplate{}).
		Where("created_by = ? AND deleted_at IS NULL", userID).Count(&total).Error
	return total, err
}

func (r *DashboardRepo) AverageStepDuration(ctx context.Context) (*int64, error) {
	duration := "TIMESTAMPDIFF(MICROSECOND, s.start_time, s.end_time) / 1000000.0"
	if DB.Dialector.Name() == "sqlite" {
		// SQLite date functions resolve milliseconds; remove Julian-day floating-point noise.
		duration = "ROUND((julianday(s.end_time) - julianday(s.start_time)) * 86400.0, 3)"
	}
	var result struct {
		Seconds sql.NullFloat64
	}
	err := DB.WithContext(ctx).Table("drill_instance_step AS s").
		Select("AVG("+duration+") AS seconds").
		Joins("JOIN drill_instance AS d ON d.id = s.drill_instance_id").
		Where("s.status = ?", "completed").
		Where("s.start_time IS NOT NULL AND s.end_time IS NOT NULL AND s.end_time >= s.start_time").
		Where("COALESCE(s.phase, '') = '' OR COALESCE(s.phase_step, '') = '' OR s.phase <> s.phase_step").
		Where(`NOT EXISTS (
			SELECT 1 FROM drill_instance_step AS child
			WHERE child.drill_instance_id = s.drill_instance_id AND child.parent_step_id = s.id
		)`).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	if !result.Seconds.Valid {
		return nil, nil
	}
	seconds := int64(math.Round(result.Seconds.Float64))
	return &seconds, nil
}
