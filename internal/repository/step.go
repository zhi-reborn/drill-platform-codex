package repository

import (
	"drill-platform/internal/domain/entity"
)

type StepRepo struct{}

func NewStepRepo() *StepRepo {
	return &StepRepo{}
}

func (r *StepRepo) FindByID(id uint64) (*entity.StepInstance, error) {
	var step entity.StepInstance
	err := DB.Preload("Logs").First(&step, id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *StepRepo) FindStepsByDrillID(drillID uint64) ([]entity.StepInstance, error) {
	var steps []entity.StepInstance
	err := DB.Where("drill_instance_id = ?", drillID).Order("seq ASC").Find(&steps).Error
	return steps, err
}

func (r *StepRepo) UpdateStatus(id uint64, status, remark string) error {
	updates := map[string]interface{}{"status": status}
	if remark != "" {
		updates["remark"] = remark
	}
	return DB.Model(&entity.StepInstance{}).Where("id = ?", id).Updates(updates).Error
}

func (r *StepRepo) CreateLogs(logs []entity.DrillInstanceLog) error {
	return DB.Create(&logs).Error
}
