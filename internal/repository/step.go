package repository

import (
	"encoding/json"

	"drill-platform/internal/domain/entity"

	"gorm.io/gorm"
)

type StepRepo struct{}

func NewStepRepo() *StepRepo {
	return &StepRepo{}
}

// TransitionStatus performs a conditional status update: the row is updated
// only if its current status is in from. Returns true when the row was
// updated, false when the status did not match (idempotent or invalid).
func (r *StepRepo) TransitionStatus(tx *gorm.DB, id uint64, from []string, to string, updates map[string]any) (bool, error) {
	if len(from) == 0 {
		return false, nil
	}
	if updates == nil {
		updates = map[string]any{}
	}
	updates["status"] = to
	res := tx.Model(&entity.StepInstance{}).
		Where("id = ? AND status IN ?", id, from).
		Updates(updates)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *StepRepo) FindByID(id uint64) (*entity.StepInstance, error) {
	var step entity.StepInstance
	err := DB.Preload("DrillInstance").Preload("Logs").First(&step, id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

// FindStepsByDrillIDs 批量获取多个演练的步骤（轻量查询，仅用于进度计算等聚合场景）。
// 返回 drill_instance_id -> steps 的映射。
func (r *StepRepo) FindStepsByDrillIDs(drillIDs []uint64) (map[uint64][]entity.StepInstance, error) {
	result := make(map[uint64][]entity.StepInstance)
	if len(drillIDs) == 0 {
		return result, nil
	}
	var steps []entity.StepInstance
	err := DB.Where("drill_instance_id IN ?", drillIDs).Order("drill_instance_id, seq ASC").Find(&steps).Error
	if err != nil {
		return nil, err
	}
	for i := range steps {
		result[steps[i].DrillInstanceID] = append(result[steps[i].DrillInstanceID], steps[i])
	}
	return result, nil
}

func (r *StepRepo) FindStepsByDrillID(drillID uint64) ([]entity.StepInstance, error) {
	var steps []entity.StepInstance
	err := DB.Where("drill_instance_id = ?", drillID).Order("seq ASC").Find(&steps).Error
	if err != nil {
		return nil, err
	}
	// 兼容旧实体:实体 StepInstance.JSONAttributes 的列标签写的是 column:attributes,
	// 但实际数据库列名是 action_params。GORM 不会自动把 action_params 读到 JSONAttributes,
	// 这里手动把 action_params 解析为 JSON 对象并填入 Attributes,保证上层 API 看到的内容最新。
	if len(steps) > 0 {
		ids := make([]uint64, 0, len(steps))
		for _, s := range steps {
			ids = append(ids, s.ID)
		}
		type row struct {
			ID          uint64
			ActionParam string
		}
		var rows []row
		DB.Table("drill_instance_step").Select("id, action_params").Where("id IN ?", ids).Scan(&rows)
		m := make(map[uint64]string, len(rows))
		for _, r := range rows {
			m[r.ID] = r.ActionParam
		}
		for i := range steps {
			if v, ok := m[steps[i].ID]; ok && v != "" && v != "null" {
				steps[i].JSONAttributes = v
			}
		}
	}
	return steps, nil
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
	allIDs := make(map[uint64]bool)
	for i := range steps {
		if steps[i].AssigneeIDs == "" || steps[i].AssigneeIDs == "[]" {
			continue
		}
		var ids []uint64
		if json.Unmarshal([]byte(steps[i].AssigneeIDs), &ids) == nil {
			for _, id := range ids {
				allIDs[id] = true
			}
		}
	}

	if len(allIDs) == 0 {
		return steps
	}

	ids := make([]uint64, 0, len(allIDs))
	for id := range allIDs {
		ids = append(ids, id)
	}

	var users []entity.User
	DB.Where("id IN ?", ids).Find(&users)
	nameMap := make(map[uint64]string, len(users))
	for _, u := range users {
		nameMap[u.ID] = u.RealName
	}

	for i := range steps {
		if steps[i].AssigneeIDs == "" || steps[i].AssigneeIDs == "[]" {
			continue
		}
		var ids []uint64
		if json.Unmarshal([]byte(steps[i].AssigneeIDs), &ids) == nil && len(ids) > 0 {
			var names []string
			for _, id := range ids {
				if n, ok := nameMap[id]; ok {
					names = append(names, n)
				}
			}
			steps[i].AssigneeNames = namesStr(names)
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
