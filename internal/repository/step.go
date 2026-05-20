package repository

import (
	"encoding/json"
	"drill-platform/internal/domain/entity"
)

type StepRepo struct{}

func NewStepRepo() *StepRepo {
	return &StepRepo{}
}

func (r *StepRepo) FindByID(id uint64) (*entity.StepInstance, error) {
	var step entity.StepInstance
	err := DB.Preload("DrillInstance").Preload("Logs").First(&step, id).Error
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

func EnrichAssigneeNames(steps []entity.StepInstance) []entity.StepInstance {
	for i := range steps {
		if steps[i].AssigneeIDs != "" && steps[i].AssigneeIDs != "[]" {
			var ids []uint64
			if json.Unmarshal([]byte(steps[i].AssigneeIDs), &ids) == nil && len(ids) > 0 {
				var names []string
				DB.Model(&entity.User{}).Where("id IN ?", ids).Pluck("real_name", &names)
				steps[i].AssigneeNames = namesStr(names)
			}
		}
	}
	return steps
}

func namesStr(names []string) string {
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}
