package storage

import (
	"context"

	"notification-service/internal/model"
)

// добавляем товар
func (s *PostgresStorage) CreateItem(ctx context.Context, item *model.Item, username string) error {
	query := `
		WITH setup AS (
			SELECT set_config('myapp.current_user', $1, true) AS cfg
		)
		INSERT INTO items (name, quantity, price) 
		SELECT $2, $3, $4 FROM setup
		RETURNING id, created_at`

	return s.db.QueryRow(ctx, query, username, item.Name, item.Quantity, item.Price).Scan(&item.ID, &item.CreatedAt)
}

// получаем все товары
func (s *PostgresStorage) GetItems(ctx context.Context) ([]*model.Item, error) {
	query := `SELECT id, name, quantity, price, created_at FROM items ORDER BY created_at DESC`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.Item
	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &item.Price, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func (s *PostgresStorage) UpdateItem(ctx context.Context, id string, item *model.Item, username string) error {
	query := `
		WITH setup AS (
			SELECT set_config('myapp.current_user', $1, true) AS cfg
		)
		UPDATE items 
		SET name = $2, quantity = $3, price = $4 
		FROM setup
		WHERE id = $5`

	_, err := s.db.Exec(ctx, query, username, item.Name, item.Quantity, item.Price, id)
	return err
}

// удаляем
func (s *PostgresStorage) DeleteItem(ctx context.Context, id string, username string) error {
	query := `
		WITH setup AS (
			SELECT set_config('myapp.current_user', $1, true) AS cfg
		)
		DELETE FROM items 
		USING setup
		WHERE id = $2`

	_, err := s.db.Exec(ctx, query, username, id)
	return err
}

// достаем логи из той самой таблицы, которую заполняют триггеры
func (s *PostgresStorage) GetItemHistory(ctx context.Context, itemID string) ([]*model.ItemHistory, error) {
	query := `
		SELECT id, item_id, action, 
		       COALESCE(old_data::text, '{}'), 
			   COALESCE(new_data::text, '{}'), 
			   changed_by, changed_at 
		FROM items_history 
		WHERE item_id = $1 
		ORDER BY changed_at DESC`

	rows, err := s.db.Query(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*model.ItemHistory
	for rows.Next() {
		var h model.ItemHistory
		if err := rows.Scan(&h.ID, &h.ItemID, &h.Action, &h.OldData, &h.NewData, &h.ChangedBy, &h.ChangedAt); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}
