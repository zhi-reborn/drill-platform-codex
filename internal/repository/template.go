package repository

import (
	"drill-platform/internal/domain/entity"

	"gorm.io/gorm"
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

	query := DB.Model(&entity.DrillTemplate{})
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

func (r *TemplateRepo) UpdateSteps(templateID uint64, steps []entity.StepTemplate) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("drill_template_id = ?", templateID).Delete(&entity.StepTemplate{}).Error; err != nil {
			return err
		}
		for i := range steps {
			steps[i].DrillTemplateID = templateID
			if steps[i].PreStepIDs == "" {
				steps[i].PreStepIDs = "[]"
			}
		}
		return tx.Create(&steps).Error
	})
}

func (r *TemplateRepo) Delete(id uint64) error {
	return DB.Delete(&entity.DrillTemplate{}, id).Error
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

		for _, step := range template.Steps {
			s := step
			s.ID = 0
			s.DrillTemplateID = clone.ID
			if err := tx.Create(&s).Error; err != nil {
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
		
		// 更新或创建传入的分类
		for i := range categories {
			var existing entity.TemplateCategory
			result := tx.Where("value = ?", categories[i].Value).First(&existing)
			
			if result.Error == nil {
				// 已存在，更新
				existing.Label = categories[i].Label
				existing.TagType = categories[i].TagType
				existing.SortOrder = categories[i].SortOrder
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			} else {
				// 不存在，创建
				if err := tx.Create(&categories[i]).Error; err != nil {
					return err
				}
			}
		}
		
		return nil
	})
}
