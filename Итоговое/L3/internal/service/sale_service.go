package service

import (
	"context"
	"errors"
	"time"

	"notification-service/internal/model"
)

// Интерфейс для хранилища продаж/транзакций
type SaleStorage interface {
	CreateSale(ctx context.Context, sale *model.Sale) error
	GetSales(ctx context.Context) ([]*model.Sale, error)
	UpdateSale(ctx context.Context, id string, sale *model.Sale) error
	DeleteSale(ctx context.Context, id string) error
	GetAnalytics(ctx context.Context, startDate, endDate time.Time) (*model.AnalyticsResult, error)
}

type SaleService struct {
	storage SaleStorage
}

func NewSaleService(s SaleStorage) *SaleService {
	return &SaleService{storage: s}
}

// для валидации
func (s *SaleService) validate(sale *model.Sale) error {
	if sale.Amount < 0 {
		return errors.New("сумма транзакции не может быть отрицательной")
	}
	if sale.Type != "income" && sale.Type != "expense" {
		return errors.New("тип должен быть 'income' (доход) или 'expense' (расход)")
	}
	if sale.Category == "" {
		return errors.New("категория не может быть пустой")
	}
	return nil
}

func (s *SaleService) CreateSale(ctx context.Context, sale *model.Sale) error {
	if err := s.validate(sale); err != nil {
		return err
	}
	// Если дату не передали, ставим текущую
	if sale.SaleDate.IsZero() {
		sale.SaleDate = time.Now()
	}
	return s.storage.CreateSale(ctx, sale)
}

func (s *SaleService) GetSales(ctx context.Context) ([]*model.Sale, error) {
	return s.storage.GetSales(ctx)
}

func (s *SaleService) UpdateSale(ctx context.Context, id string, sale *model.Sale) error {
	if err := s.validate(sale); err != nil {
		return err
	}
	return s.storage.UpdateSale(ctx, id, sale)
}

func (s *SaleService) DeleteSale(ctx context.Context, id string) error {
	return s.storage.DeleteSale(ctx, id)
}

func (s *SaleService) GetAnalytics(ctx context.Context, startDate, endDate time.Time) (*model.AnalyticsResult, error) {
	if endDate.Before(startDate) {
		return nil, errors.New("конечная дата не может быть раньше начальной")
	}
	return s.storage.GetAnalytics(ctx, startDate, endDate)
}
