package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"notification-service/internal/model"

	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
)

//структура, которая отвечает за работу с базой данных
type PostgresStorage struct {
	db *pgxdriver.Postgres
}

type RowsScanner interface {
	Next() bool
	Scan(dest ...any) error
}

//конструктор (создает новый объект хранилища)
func NewPostgresStorage(db *pgxdriver.Postgres) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// метод для сохранения нового уведомления в базу
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

// метод для поиска уведомления в базе по его ID
func (s *PostgresStorage) GetByID(ctx context.Context, id string) (*model.Notification, error) {
	query, args, err := s.db.Select("id", "message", "recipient", "channel", "status", "scheduled_at", "created_at").
		From("notifications").
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("ошибка генерации SQL: %w", err)
	}

	var n model.Notification

	err = s.db.QueryRow(ctx, query, args...).Scan(
		&n.ID, &n.Message, &n.Recipient, &n.Channel, &n.Status, &n.ScheduledAt, &n.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("уведомление не найдено: %w", err)
	}

	return &n, nil
}

// метод для отмены уведомления (меняем статус)
func (s *PostgresStorage) Cancel(ctx context.Context, id string) error {
	query, args, err := s.db.Update("notifications").
		Set("status", "canceled").
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

// меняет статус уведомления
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

// находит все уведомления, время которых пришло
func (s *PostgresStorage) GetDueNotifications(ctx context.Context) ([]model.Notification, error) {
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
			continue 
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

//сдвигает время отправки при ошибке (экспоненциально)
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

	queryFailed := `UPDATE notifications SET status = 'failed' WHERE id = $1 AND retry_count >= 5`
	s.db.Exec(ctx, queryFailed, id)

	return nil
}

// сохраняет новую короткую ссылку в БД
func (s *PostgresStorage) SaveURL(ctx context.Context, u *model.URL) error {
	query := `INSERT INTO urls (id, short_code, original_url, created_at) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(ctx, query, u.ID, u.ShortCode, u.OriginalURL, u.CreatedAt)
	return err
}

// ищет оригинальную ссылку по короткому коду
func (s *PostgresStorage) GetURLByCode(ctx context.Context, code string) (*model.URL, error) {
	var u model.URL
	query := `SELECT id, short_code, original_url, created_at FROM urls WHERE short_code = $1`
	err := s.db.QueryRow(ctx, query, code).Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

//записывает факт перехода для аналитики
func (s *PostgresStorage) SaveClick(ctx context.Context, click *model.Click) error {
	query := `INSERT INTO clicks (id, url_id, user_agent, clicked_at) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(ctx, query, click.ID, click.URLID, click.UserAgent, click.ClickedAt)
	return err
}

//считает общее количество переходов по ID ссылки
func (s *PostgresStorage) GetClicksCount(ctx context.Context, urlID string) (int, error) {
	var count int
	query := `SELECT count(*) FROM clicks WHERE url_id = $1`
	err := s.db.QueryRow(ctx, query, urlID).Scan(&count)
	return count, err
}

// группирует клики по User-Agent
func (s *PostgresStorage) GetClickStats(ctx context.Context, urlID string) (map[string]int, error) {
	stats := make(map[string]int)

	query := `SELECT COALESCE(NULLIF(user_agent, ''), 'Unknown'), count(*) FROM clicks WHERE url_id = $1 GROUP BY user_agent`

	rows, err := s.db.Query(ctx, query, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ua string
		var count int
		if err := rows.Scan(&ua, &count); err == nil {
			stats[ua] = count
		}
	}
	return stats, nil
}

// сохраняет новый комментарий
func (s *PostgresStorage) CreateComment(ctx context.Context, c *model.Comment) error {
	query := `INSERT INTO comments (id, parent_id, author, text, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(ctx, query, c.ID, c.ParentID, c.Author, c.Text, c.CreatedAt)
	return err
}

// удаляет комментарий.
func (s *PostgresStorage) DeleteComment(ctx context.Context, id string) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := s.db.Exec(ctx, query, id)
	return err
}

// достает корневой комментарий и ВСЕ его вложенные ответы любой глубины
func (s *PostgresStorage) GetCommentTree(ctx context.Context, rootID string) ([]*model.Comment, error) {
	query := `
	WITH RECURSIVE tree AS (
		SELECT id, parent_id, author, text, created_at FROM comments WHERE id = $1
		UNION ALL
		SELECT c.id, c.parent_id, c.author, c.text, c.created_at
		FROM comments c
		INNER JOIN tree t ON c.parent_id = t.id
	)
	SELECT id, parent_id, author, text, created_at FROM tree ORDER BY created_at ASC;
	`
	rows, err := s.db.Query(ctx, query, rootID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanComments(rows)
}

// достает корневые комментарии (без родителя) с пагинацией и сортировкой
func (s *PostgresStorage) GetRootComments(ctx context.Context, limit, offset int, sortDesc bool) ([]*model.Comment, error) {
	order := "DESC"
	if !sortDesc {
		order = "ASC"
	}

	query := `SELECT id, parent_id, author, text, created_at FROM comments WHERE parent_id IS NULL ORDER BY created_at ` + order + ` LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanComments(rows)
}

//выполняет полнотекстовый поиск по словам
func (s *PostgresStorage) SearchComments(ctx context.Context, keyword string) ([]*model.Comment, error) {
	query := `SELECT id, parent_id, author, text, created_at FROM comments 
	          WHERE to_tsvector('russian', text) @@ plainto_tsquery('russian', $1) 
	          ORDER BY created_at DESC LIMIT 50`

	rows, err := s.db.Query(ctx, query, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanComments(rows)
}

func scanComments(rows RowsScanner) ([]*model.Comment, error) {
	var comments []*model.Comment
	for rows.Next() {
		var c model.Comment
		var parentID sql.NullString

		if err := rows.Scan(&c.ID, &parentID, &c.Author, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}

		if parentID.Valid {
			c.ParentID = &parentID.String
		}
		comments = append(comments, &c)
	}
	return comments, nil
}

// сохраняет информацию о новой картинке (со статусом pending)
func (s *PostgresStorage) SaveImage(ctx context.Context, img *model.Image) error {
	query := `INSERT INTO images (id, filename, status, format, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(ctx, query, img.ID, img.Filename, img.Status, img.Format, img.CreatedAt)
	return err
}

// получает информацию о картинке по её ID
func (s *PostgresStorage) GetImage(ctx context.Context, id string) (*model.Image, error) {
	var img model.Image
	query := `SELECT id, filename, status, format, created_at FROM images WHERE id = $1`
	err := s.db.QueryRow(ctx, query, id).Scan(&img.ID, &img.Filename, &img.Status, &img.Format, &img.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

// меняет статус картинки (например, pending -> processing -> ready)
func (s *PostgresStorage) UpdateImageStatus(ctx context.Context, id, status string) error {
	query := `UPDATE images SET status = $1 WHERE id = $2`
	_, err := s.db.Exec(ctx, query, status, id)
	return err
}

//  удаляет запись о картинке из БД
func (s *PostgresStorage) DeleteImage(ctx context.Context, id string) error {
	query := `DELETE FROM images WHERE id = $1`
	_, err := s.db.Exec(ctx, query, id)
	return err
}

// достает все загруженные картинки
func (s *PostgresStorage) GetAllImages(ctx context.Context) ([]*model.Image, error) {
	query := `SELECT id, filename, status, format, created_at FROM images ORDER BY created_at DESC`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*model.Image
	for rows.Next() {
		var img model.Image
		if err := rows.Scan(&img.ID, &img.Filename, &img.Status, &img.Format, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, &img)
	}
	return images, nil
}
