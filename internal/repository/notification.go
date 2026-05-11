package repository

import (
	"drill-platform/internal/domain/entity"
)

type NotificationRepo struct{}

func NewNotificationRepo() *NotificationRepo {
	return &NotificationRepo{}
}

func (r *NotificationRepo) List(userID uint64, page, pageSize int, unreadOnly bool) ([]entity.Notification, int64, error) {
	var notifications []entity.Notification
	var total int64

	query := DB.Model(&entity.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		query = query.Where("is_read = ?", 0)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&notifications).Error
	return notifications, total, err
}

func (r *NotificationRepo) FindByID(id uint64) (*entity.Notification, error) {
	var notification entity.Notification
	err := DB.First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepo) FindByUserIDAndID(userID, id uint64) (*entity.Notification, error) {
	var notification entity.Notification
	err := DB.Where("id = ? AND user_id = ?", id, userID).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepo) MarkAsRead(id uint64) error {
	return DB.Model(&entity.Notification{}).Where("id = ?", id).Update("is_read", 1).Error
}

func (r *NotificationRepo) MarkAllAsRead(userID uint64) error {
	return DB.Model(&entity.Notification{}).Where("user_id = ? AND is_read = ?", userID, 0).Update("is_read", 1).Error
}

func (r *NotificationRepo) Delete(id uint64) error {
	return DB.Delete(&entity.Notification{}, id).Error
}

func (r *NotificationRepo) Create(notification *entity.Notification) error {
	return DB.Create(notification).Error
}
