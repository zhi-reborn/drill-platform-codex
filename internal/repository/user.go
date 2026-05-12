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
