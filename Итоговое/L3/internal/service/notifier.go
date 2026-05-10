package service

import (
	"context"
	"time"

	"notification-service/internal/model"

	"github.com/google/uuid"
)


type Storage interface {
	Create(ctx context.Context, n *model.Notification) error
	GetByID(ctx context.Context, id string) (*model.Notification, error)
	Cancel(ctx context.Context, id string) error
}

type Broker interface {
	PublishID(ctx context.Context, id string) error
}

type Cache interface {
	SetNotification(ctx context.Context, notif *model.Notification) error
	GetNotification(ctx context.Context, id string) (*model.Notification, error)
}


type NotifierService struct {
	storage Storage
	broker  Broker
	cache   Cache
}

func NewNotifierService(s Storage, b Broker, c Cache) *NotifierService {
	return &NotifierService{storage: s, broker: b, cache: c}
}

func (s *NotifierService) CreateNotification(ctx context.Context, message, recipient, channel string, scheduledAt time.Time) (*model.Notification, error) {

	notification := &model.Notification{
		ID:          uuid.New().String(),
		Message:     message,
		Recipient:   recipient,
		Channel:     channel,
		Status:      model.StatusPending,
		ScheduledAt: scheduledAt,
		CreatedAt:   time.Now(),
	}

	err := s.storage.Create(ctx, notification)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

// логика получения уведомления
func (s *NotifierService) GetStatus(ctx context.Context, id string) (*model.Notification, error) {

	notif, err := s.cache.GetNotification(ctx, id)
	if err == nil && notif != nil {
		return notif, nil
	}

	notif, err = s.storage.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if notif.Status == "sent" || notif.Status == "failed" || notif.Status == "canceled" {
		_ = s.cache.SetNotification(ctx, notif)
	}

	return notif, nil
}

// логика отмены
func (s *NotifierService) CancelNotification(ctx context.Context, id string) error {
	return s.storage.Cancel(ctx, id)
}
