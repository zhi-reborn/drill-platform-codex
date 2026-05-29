package repository

import (
	"drill-platform/internal/domain/entity"
	"sort"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TemplateRepo struct{}

func NewTemplateRepo() *TemplateRepo {
	return &TemplateRepo{}
}

func (r *TemplateRepo) FindByID(id uint64) (*entity.DrillTemplate, error) {
	var template entity.DrillTemplate
	err := DB.Preload("Steps").First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepo) List(page, pageSize int, category string) ([]entity.DrillTemplate, int64, error) {
	var templates []entity.DrillTemplate
	var total int64

	query := DB.Model(&entity.DrillTemplate{}).Preload("Steps")
	if category != "" {
		query = query.Where("category = ?", category)
	}

	query.Count(&total)
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error
	return templates, total, err
}

func (r *TemplateRepo) Create(template *entity.DrillTemplate) error {
	return DB.Create(template).Error
}

func (r *TemplateRepo) Update(template *entity.DrillTemplate) error {
	return DB.Save(template).Error
}

func (r *TemplateRepo) UpdateStep(step *entity.StepTemplate) error {
		return DB.Model(&entity.StepTemplate{}).Where("id = ? AND drill_template_id = ?", step.ID, step.DrillTemplateID).Updates(map[string]interface{}{
		"name":                        step.Name,
		"seq":                         step.Seq,
		"step_type":                   step.StepType,
		"timeout_minutes":            step.TimeoutMinutes,
		"guide_content":              step.GuideContent,
		"default_assignee_role":      step.DefaultAssigneeRole,
		"executor_team":              step.ExecutorTeam,
		"parent_step_id":             step.ParentStepID,
		"phase":                      step.Phase,
		"phase_step":                 step.PhaseStep,
		"estimated_duration_minutes": step.EstimatedDurationMinutes,
		"estimated_start_offset":     step.EstimatedStartOffset,
		"attributes":                 step.JSONAttributes,
	}).Error
}

func (r *TemplateRepo) UpdateSteps(templateID uint64, steps []entity.StepTemplate) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		sort.Slice(steps, func(i, j int) bool { return steps[i].Seq < steps[j].Seq })

		if err := tx.Where("drill_template_id = ?", templateID).Delete(&entity.StepTemplate{}).Error; err != nil {
			return err
		}
		if len(steps) == 0 {
			return nil
		}

		// idxToNewID: 1-based array index → new DB auto-increment ID
		// Frontend sends parent_step_id as 1-based position in the sorted array.
		idxToNewID := make(map[int]uint64)

		for i := range steps {
			steps[i].DrillTemplateID = templateID
			if steps[i].PreStepIDs == "" {
				steps[i].PreStepIDs = "[]"
			}

			if steps[i].ParentStepID != nil {
				parentPos := int(*steps[i].ParentStepID)
				if parentPos == 0 {
					steps[i].ParentStepID = nil
				} else if newParentID, ok := idxToNewID[parentPos]; ok {
					steps[i].ParentStepID = &newParentID
				} else {
					steps[i].ParentStepID = nil
				}
			}

			steps[i].ID = 0
			if err := tx.Create(&steps[i]).Error; err != nil {
				return err
			}
			idxToNewID[i+1] = steps[i].ID
		}
		return nil
	})
}

func (r *TemplateRepo) Delete(id uint64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("drill_template_id = ?", id).Delete(&entity.StepTemplate{}).Error; err != nil {
			return err
		}
		return tx.Delete(&entity.DrillTemplate{}, id).Error
	})
}

func (r *TemplateRepo) Clone(id uint64) (*entity.DrillTemplate, error) {
	template, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}

	var newTemplate *entity.DrillTemplate
	err = DB.Transaction(func(tx *gorm.DB) error {
		clone := *template
		clone.ID = 0
		clone.Name = template.Name + " (副本)"
		clone.Steps = nil
		if err := tx.Create(&clone).Error; err != nil {
			return err
		}

		steps := make([]entity.StepTemplate, len(template.Steps))
		for i, step := range template.Steps {
			s := step
			s.ID = 0
			s.DrillTemplateID = clone.ID
			steps[i] = s
		}
		if len(steps) > 0 {
			if err := tx.Create(&steps).Error; err != nil {
				return err
			}
		}

		loaded, err2 := r.FindByID(clone.ID)
		if err2 != nil {
			return err2
		}
		newTemplate = loaded
		return nil
	})
	return newTemplate, err
}

func (r *TemplateRepo) GetCategories() ([]entity.TemplateCategory, error) {
	var categories []entity.TemplateCategory
	err := DB.Order("sort_order ASC").Find(&categories).Error
	return categories, err
}

func (r *TemplateRepo) SaveCategories(categories []entity.TemplateCategory) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 收集传入分类的 value 列表
		valueMap := make(map[string]bool)
		for i := range categories {
			categories[i].SortOrder = i + 1
			valueMap[categories[i].Value] = true
		}
		
		// 删除不在传入列表中的分类
		var values []string
		for v := range valueMap {
			values = append(values, v)
		}
		if err := tx.Where("value NOT IN ?", values).Delete(&entity.TemplateCategory{}).Error; err != nil {
			return err
		}
		
		// 批量 upsert：INSERT ... ON DUPLICATE KEY UPDATE
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "value"}},
			DoUpdates: clause.AssignmentColumns([]string{"label", "tag_type", "sort_order"}),
		}).Create(&categories).Error; err != nil {
			return err
		}
		
		return nil
	})
}
