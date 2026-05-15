package repository

import (
	"drill-platform/internal/domain/entity"
)

type DrillRepo struct{}

func NewDrillRepo() *DrillRepo {
	return &DrillRepo{}
}

func (r *DrillRepo) FindByID(id uint64) (*entity.DrillInstance, error) {
	var drill entity.DrillInstance
	err := DB.Preload("Template").Preload("Steps").First(&drill, id).Error
	if err != nil {
		return nil, err
	}
	return &drill, nil
}

func (r *DrillRepo) List(page, pageSize int, status string) ([]entity.DrillInstance, int64, error) {
	var drills []entity.DrillInstance
	var total int64

	query := DB.Model(&entity.DrillInstance{}).Preload("Template")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&drills).Error
	return drills, total, err
}

func (r *DrillRepo) Create(drill *entity.DrillInstance) error {
	return DB.Create(drill).Error
}

func (r *DrillRepo) Update(drill *entity.DrillInstance) error {
	return DB.Save(drill).Error
}

func (r *DrillRepo) UpdateStatus(id uint64, status string) error {
	return DB.Model(&entity.DrillInstance{}).Where("id = ?", id).Update("status", status).Error
}

func (r *DrillRepo) GetCurrentStepID(id uint64) (*uint64, error) {
	var drill entity.DrillInstance
	if err := DB.Select("current_step_id").First(&drill, id).Error; err != nil {
		return nil, err
	}
	return drill.CurrentStepID, nil
}

func (r *DrillRepo) CreateLog(log *entity.DrillInstanceLog) error {
	return DB.Create(log).Error
}

func (r *DrillRepo) GetLogs(drillID uint64) ([]entity.DrillInstanceLog, error) {
	var logs []entity.DrillInstanceLog
	err := DB.Where("drill_instance_id = ?", drillID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}
