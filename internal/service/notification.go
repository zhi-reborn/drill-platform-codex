package service

import (
	"drill-platform/internal/domain/dto"
	"drill-platform/internal/domain/entity"
	"drill-platform/internal/repository"
	"errors"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepo
}

func NewNotificationService(notificationRepo *repository.NotificationRepo) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

func (s *NotificationService) GetList(userID uint64, query *dto.NotificationQuery) (*dto.NotificationListResponse, error) {
	query.Normalize()
	
	notifications, total, err := s.notificationRepo.List(userID, query.Page, query.PageSize, query.UnreadOnly)
	if err != nil {
		return nil, err
	}

	return &dto.NotificationListResponse{
		Total: total,
		Items: notifications,
	}, nil
}

func (s *NotificationService) MarkAsRead(userID, id uint64) (*entity.Notification, error) {
	notification, err := s.notificationRepo.FindByUserIDAndID(userID, id)
	if err != nil {
		return nil, errors.New("通知不存在或无权访问")
	}

	if err := s.notificationRepo.MarkAsRead(id); err != nil {
		return nil, err
	}

	notification.IsRead = true
	return notification, nil
}

func (s *NotificationService) MarkAllAsRead(userID uint64) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

func (s *NotificationService) Delete(userID, id uint64) error {
	notification, err := s.notificationRepo.FindByUserIDAndID(userID, id)
	if err != nil {
		return errors.New("通知不存在或无权访问")
	}

	return s.notificationRepo.Delete(notification.ID)
}

func (s *NotificationService) CreateNotification(notification *entity.Notification) error {
	return s.notificationRepo.Create(notification)
}
