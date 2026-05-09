package storage

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/model"

	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
)

// PostgresStorage — структура, которая отвечает за работу с базой данных
type PostgresStorage struct {
	db *pgxdriver.Postgres
}

// NewPostgresStorage — конструктор (создает новый объект хранилища)
func NewPostgresStorage(db *pgxdriver.Postgres) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// Create — метод для сохранения нового уведомления в базу
func (s *PostgresStorage) Create(ctx context.Context, n *model.Notification) error {
	query, args, err := s.db.Insert("notifications").
		Columns("id", "message", "recipient", "channel", "status", "scheduled_at", "created_at").
		Values(n.ID, n.Message, n.Recipient, n.Channel, n.Status, n.ScheduledAt, n.CreatedAt).
		ToSql()

	if err != nil {
		return fmt.Errorf("ошибка генерации SQL: %w", err)
	}

	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка сохранения в БД: %w", err)
	}

	return nil

}

// GetByID — метод для поиска уведомления в базе по его ID
func (s *PostgresStorage) GetByID(ctx context.Context, id string) (*model.Notification, error) {
	// Собираем SELECT запрос
	query, args, err := s.db.Select("id", "message", "recipient", "channel", "status", "scheduled_at", "created_at").
		From("notifications").
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("ошибка генерации SQL: %w", err)
	}

	var n model.Notification

	// Выполняем запрос и записываем результат в структуру n
	err = s.db.QueryRow(ctx, query, args...).Scan(
		&n.ID, &n.Message, &n.Recipient, &n.Channel, &n.Status, &n.ScheduledAt, &n.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("уведомление не найдено: %w", err)
	}

	return &n, nil
}

// Cancel — метод для отмены уведомления (меняем статус)
func (s *PostgresStorage) Cancel(ctx context.Context, id string) error {
	query, args, err := s.db.Update("notifications").
		Set("status", "canceled"). // Ставим статус "отменено"
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return fmt.Errorf("ошибка генерации SQL: %w", err)
	}

	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка обновления БД: %w", err)
	}

	return nil
}

// UpdateStatus — меняет статус уведомления (например, на "sent")
func (s *PostgresStorage) UpdateStatus(ctx context.Context, id string, status string) error {
	query, args, err := s.db.Update("notifications").
		Set("status", status).
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return fmt.Errorf("ошибка генерации SQL: %w", err)
	}

	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка обновления БД: %w", err)
	}

	return nil
}

// GetDueNotifications — находит все уведомления, время которых пришло
func (s *PostgresStorage) GetDueNotifications(ctx context.Context) ([]model.Notification, error) {
	// Ищем записи со статусом pending, у которых scheduled_at меньше или равно текущему времени
	query, args, err := s.db.Select("id", "message", "recipient", "channel", "status", "scheduled_at", "created_at").
		From("notifications").
		Where("status = ?", "pending").
		Where("scheduled_at <= ?", time.Now()).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("ошибка SQL: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к БД: %w", err)
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		err := rows.Scan(&n.ID, &n.Message, &n.Recipient, &n.Channel, &n.Status, &n.ScheduledAt, &n.CreatedAt)
		if err != nil {
			continue // если какую-то строку не смогли прочитать, пропускаем её
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// RescheduleNotification  сдвигает время отправки при ошибке (экспоненциально)
func (s *PostgresStorage) RescheduleNotification(ctx context.Context, id string) error {
	query := `
		UPDATE notifications 
		SET status = 'pending', 
			retry_count = retry_count + 1, 
			scheduled_at = NOW() + (POWER(2, retry_count) * INTERVAL '10 seconds')
		WHERE id = $1 AND retry_count < 5` // Даем максимум 5 попыток

	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка сдвига времени: %w", err)
	}

	// Если попыток уже 5 или больше  = статус failed (ошибка)
	queryFailed := `UPDATE notifications SET status = 'failed' WHERE id = $1 AND retry_count >= 5`
	s.db.Exec(ctx, queryFailed, id)

	return nil
}
