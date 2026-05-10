package storage

import (
	"context"
	"time"

	"notification-service/internal/model"
)

// создает новую запись
func (s *PostgresStorage) CreateSale(ctx context.Context, sale *model.Sale) error {
	query := `
		INSERT INTO sales (type, amount, category, sale_date) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at`

	return s.db.QueryRow(ctx, query, sale.Type, sale.Amount, sale.Category, sale.SaleDate).Scan(&sale.ID, &sale.CreatedAt)
}

// возвращает все записи с базовой сортировкой
func (s *PostgresStorage) GetSales(ctx context.Context) ([]*model.Sale, error) {
	query := `SELECT id, type, amount, category, sale_date, created_at FROM sales ORDER BY sale_date DESC`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []*model.Sale
	for rows.Next() {
		var item model.Sale
		if err := rows.Scan(&item.ID, &item.Type, &item.Amount, &item.Category, &item.SaleDate, &item.CreatedAt); err != nil {
			return nil, err
		}
		sales = append(sales, &item)
	}
	return sales, nil
}

// обновляет запись по ID
func (s *PostgresStorage) UpdateSale(ctx context.Context, id string, sale *model.Sale) error {
	query := `
		UPDATE sales 
		SET type = $1, amount = $2, category = $3, sale_date = $4 
		WHERE id = $5`

	_, err := s.db.Exec(ctx, query, sale.Type, sale.Amount, sale.Category, sale.SaleDate, id)
	return err
}

// удаляет запись
func (s *PostgresStorage) DeleteSale(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM sales WHERE id = $1`, id)
	return err
}

func (s *PostgresStorage) GetAnalytics(ctx context.Context, startDate, endDate time.Time) (*model.AnalyticsResult, error) {
	query := `
		SELECT 
			COUNT(*),
			COALESCE(SUM(amount), 0),
			COALESCE(AVG(amount), 0),
			COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY amount), 0) AS median,
			COALESCE(percentile_cont(0.9) WITHIN GROUP (ORDER BY amount), 0) AS percentile_90
		FROM sales
		WHERE sale_date >= $1 AND sale_date <= $2
	`

	var res model.AnalyticsResult
	err := s.db.QueryRow(ctx, query, startDate, endDate).Scan(
		&res.Count,
		&res.TotalSum,
		&res.Average,
		&res.Median,
		&res.Percentile90,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}
