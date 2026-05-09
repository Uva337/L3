package service

import (
	"context"

	"notification-service/internal/model"
)

// Интерфейс для хранилища мероприятий
type EventStorage interface {
	CreateEvent(ctx context.Context, e *model.Event) error
	GetAllEvents(ctx context.Context) ([]*model.Event, error)
	BookSpot(ctx context.Context, eventID, userID string) (*model.Booking, error)
	ConfirmBooking(ctx context.Context, bookingID string) error
	GetEventBookings(ctx context.Context, eventID string) ([]*model.Booking, error)
	GetEventByID(ctx context.Context, id string) (*model.Event, error)
}

type EventService struct {
	storage EventStorage
}

func NewEventService(s EventStorage) *EventService {
	return &EventService{storage: s}
}

func (s *EventService) CreateEvent(ctx context.Context, e *model.Event) error {
	// По умолчанию, если TTL не задан, ставим 15 минут
	if e.BookingTTLMinutes <= 0 {
		e.BookingTTLMinutes = 15
	}
	// При создании доступные места равны всем местам
	e.AvailableSpots = e.TotalSpots
	return s.storage.CreateEvent(ctx, e)
}

func (s *EventService) GetAllEvents(ctx context.Context) ([]*model.Event, error) {
	return s.storage.GetAllEvents(ctx)
}

func (s *EventService) BookSpot(ctx context.Context, eventID, userID string) (*model.Booking, error) {
	return s.storage.BookSpot(ctx, eventID, userID)
}

func (s *EventService) ConfirmBooking(ctx context.Context, bookingID string) error {
	return s.storage.ConfirmBooking(ctx, bookingID)
}

func (s *EventService) GetEventBookings(ctx context.Context, eventID string) ([]*model.Booking, error) {
	return s.storage.GetEventBookings(ctx, eventID)
}

func (s *EventService) GetEventByID(ctx context.Context, id string) (*model.Event, error) {
	return s.storage.GetEventByID(ctx, id)
}
