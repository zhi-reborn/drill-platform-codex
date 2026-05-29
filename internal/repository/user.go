package repository

import (
	"drill-platform/internal/domain/entity"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) FindByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(id uint64) (*entity.User, error) {
	var user entity.User
	err := DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) List(page, pageSize int, role string) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := DB.Model(&entity.User{})
	if role != "" {
		query = query.Where("role = ?", role)
	}

	query.Count(&total)
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *UserRepo) Create(user *entity.User) error {
	return DB.Create(user).Error
}

func (r *UserRepo) Update(user *entity.User) error {
	return DB.Save(user).Error
}

func (r *UserRepo) Delete(id uint64) error {
	return DB.Delete(&entity.User{}, id).Error
}

func (r *UserRepo) ListAll() ([]entity.User, error) {
	var users []entity.User
	err := DB.Where("status = 1").Order("role, id").Find(&users).Error
	return users, err
}

// FindByIDs 批量查询用户，返回 ID→用户 映射
func (r *UserRepo) FindByIDs(ids []uint64) (map[uint64]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []entity.User
	if err := DB.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	result := make(map[uint64]*entity.User, len(users))
	for i := range users {
		result[users[i].ID] = &users[i]
	}
	return result, nil
}

// FindByDepartments 批量按部门查询用户，返回 department→用户列表 映射
func (r *UserRepo) FindByDepartments(departments []string) (map[string][]entity.User, error) {
	if len(departments) == 0 {
		return nil, nil
	}
	var users []entity.User
	if err := DB.Where("department IN ? AND status = 1", departments).Find(&users).Error; err != nil {
		return nil, err
	}
	result := make(map[string][]entity.User)
	for _, u := range users {
		result[u.Department] = append(result[u.Department], u)
	}
	return result, nil
}
func (r *UserRepo) GetDistinctDepartments() ([]string, error) {
	var departments []string
	err := DB.Model(&entity.User{}).
		Where("department IS NOT NULL AND department != ''").
		Distinct("department").
		Order("department").
		Pluck("department", &departments).Error
	return departments, err
}
