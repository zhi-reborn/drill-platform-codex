package repository

import (
	"drill-platform/internal/domain/entity"
	"strings"

	"gorm.io/gorm"
)

type DrillRepo struct{}

func NewDrillRepo() *DrillRepo {
	return &DrillRepo{}
}

func (r *DrillRepo) FindByID(id uint64) (*entity.DrillInstance, error) {
	var drill entity.DrillInstance
	err := DB.Where("id = ?", id).Preload("Template").First(&drill).Error
	if err != nil {
		return nil, err
	}
	return &drill, nil
}

func (r *DrillRepo) FindByIDWithSteps(id uint64) (*entity.DrillInstance, error) {
	var drill entity.DrillInstance
	err := DB.Where("id = ?", id).Preload("Template").Preload("Steps").First(&drill).Error
	if err != nil {
		return nil, err
	}
	return &drill, nil
}

func (r *DrillRepo) List(page, pageSize int, status string, keyword string) ([]entity.DrillInstance, int64, error) {
	var drills []entity.DrillInstance
	var total int64

	query := DB.Model(&entity.DrillInstance{}).Preload("Template")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)
	err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&drills).Error
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
	if err := DB.Select("current_task_id").First(&drill, id).Error; err != nil {
		return nil, err
	}
	return drill.CurrentStepID, nil
}

func (r *DrillRepo) CreateLog(log *entity.DrillInstanceLog) error {
	return DB.Create(log).Error
}

func (r *DrillRepo) GetLogs(drillID uint64, limit int) ([]entity.DrillInstanceLog, error) {
	if limit <= 0 {
		limit = 200
	}
	var logs []entity.DrillInstanceLog
	err := DB.Where("drill_instance_id = ?", drillID).Order("created_at DESC, id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *DrillRepo) Delete(id uint64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 获取所有步骤实例 ID
		var stepIDs []uint64
		if err := tx.Where("drill_instance_id = ?", id).
			Model(&entity.StepInstance{}).
			Pluck("id", &stepIDs).Error; err != nil {
			return err
		}

		// 删除步骤日志（task_instance_id 不为空的记录）
		if len(stepIDs) > 0 {
			if err := tx.Where("drill_instance_id = ? AND task_instance_id IN ?", id, stepIDs).
				Delete(&entity.DrillInstanceLog{}).Error; err != nil {
				return err
			}
		}

		// 删除演练日志
		if err := tx.Where("drill_instance_id = ?", id).
			Delete(&entity.DrillInstanceLog{}).Error; err != nil {
			return err
		}

		// 删除人员分配
		if err := tx.Where("drill_instance_id = ?", id).
			Delete(&entity.DrillAssignee{}).Error; err != nil {
			return err
		}

		// 删除步骤实例
		if err := tx.Where("drill_instance_id = ?", id).
			Delete(&entity.StepInstance{}).Error; err != nil {
			return err
		}

		// 删除演练实例
		return tx.Delete(&entity.DrillInstance{}, id).Error
	})
}
