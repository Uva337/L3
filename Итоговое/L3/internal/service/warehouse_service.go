package service

import (
	"context"

	"notification-service/internal/model"
)

// Интерфейс для хранилища склада
type WarehouseStorage interface {
	CreateItem(ctx context.Context, item *model.Item, username string) error
	GetItems(ctx context.Context) ([]*model.Item, error)
	UpdateItem(ctx context.Context, id string, item *model.Item, username string) error
	DeleteItem(ctx context.Context, id string, username string) error
	GetItemHistory(ctx context.Context, itemID string) ([]*model.ItemHistory, error)
}

type WarehouseService struct {
	storage WarehouseStorage
}

func NewWarehouseService(s WarehouseStorage) *WarehouseService {
	return &WarehouseService{storage: s}
}

func (s *WarehouseService) CreateItem(ctx context.Context, item *model.Item, username string) error {
	return s.storage.CreateItem(ctx, item, username)
}

func (s *WarehouseService) GetItems(ctx context.Context) ([]*model.Item, error) {
	return s.storage.GetItems(ctx)
}

func (s *WarehouseService) UpdateItem(ctx context.Context, id string, item *model.Item, username string) error {
	return s.storage.UpdateItem(ctx, id, item, username)
}

func (s *WarehouseService) DeleteItem(ctx context.Context, id string, username string) error {
	return s.storage.DeleteItem(ctx, id, username)
}

func (s *WarehouseService) GetItemHistory(ctx context.Context, itemID string) ([]*model.ItemHistory, error) {
	return s.storage.GetItemHistory(ctx, itemID)
}
