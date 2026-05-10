package storage

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/model"
)

// CreateEvent создает новое мероприятие
func (s *PostgresStorage) CreateEvent(ctx context.Context, e *model.Event) error {
	query := `INSERT INTO events (title, date, total_spots, available_spots, booking_ttl_minutes) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return s.db.QueryRow(ctx, query, e.Title, e.Date, e.TotalSpots, e.TotalSpots, e.BookingTTLMinutes).Scan(&e.ID)
}

// GetAllEvents возвращает список мероприятий
func (s *PostgresStorage) GetAllEvents(ctx context.Context) ([]*model.Event, error) {
	query := `SELECT id, title, date, total_spots, available_spots, booking_ttl_minutes FROM events ORDER BY created_at DESC`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Date, &e.TotalSpots, &e.AvailableSpots, &e.BookingTTLMinutes); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

func (s *PostgresStorage) BookSpot(ctx context.Context, eventID, userID string) (*model.Booking, error) {
	var ttl int

	updateQuery := `
		UPDATE events 
		SET available_spots = available_spots - 1 
		WHERE id = $1 AND available_spots > 0 
		RETURNING booking_ttl_minutes`

	err := s.db.QueryRow(ctx, updateQuery, eventID).Scan(&ttl)
	if err != nil {
		return nil, fmt.Errorf("нет свободных мест или мероприятие не найдено")
	}

	expiresAt := time.Now().Add(time.Duration(ttl) * time.Minute)
	b := &model.Booking{
		EventID:   eventID,
		UserID:    userID,
		Status:    "pending",
		ExpiresAt: expiresAt,
	}

	insertQuery := `INSERT INTO bookings (event_id, user_id, status, expires_at) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err = s.db.QueryRow(ctx, insertQuery, b.EventID, b.UserID, b.Status, b.ExpiresAt).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *PostgresStorage) ConfirmBooking(ctx context.Context, bookingID string) error {
	query := `UPDATE bookings SET status = 'confirmed' WHERE id = $1 AND status = 'pending'`
	tag, err := s.db.Exec(ctx, query, bookingID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("бронь не найдена или уже не в статусе pending")
	}
	return nil
}


func (s *PostgresStorage) CancelExpiredBookings(ctx context.Context) (int, error) {
	now := time.Now()

	query := `
		UPDATE bookings 
		SET status = 'cancelled' 
		WHERE status = 'pending' AND expires_at <= $1
		RETURNING event_id
	`
	rows, err := s.db.Query(ctx, query, now)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	eventCounts := make(map[string]int)
	cancelledCount := 0

	for rows.Next() {
		var eventID string
		if err := rows.Scan(&eventID); err != nil {
			return 0, err
		}
		eventCounts[eventID]++
		cancelledCount++
	}

	for evID, count := range eventCounts {
		_, _ = s.db.Exec(ctx, `UPDATE events SET available_spots = available_spots + $1 WHERE id = $2`, count, evID)
	}

	return cancelledCount, nil
}

func (s *PostgresStorage) GetEventBookings(ctx context.Context, eventID string) ([]*model.Booking, error) {
	query := `SELECT id, event_id, user_id, status, expires_at, created_at FROM bookings WHERE event_id = $1 AND status != 'cancelled' ORDER BY created_at`
	rows, err := s.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.ExpiresAt, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, &b)
	}
	return bookings, nil
}

func (s *PostgresStorage) GetEventByID(ctx context.Context, id string) (*model.Event, error) {
	var e model.Event
	query := `SELECT id, title, date, total_spots, available_spots, booking_ttl_minutes FROM events WHERE id = $1`
	err := s.db.QueryRow(ctx, query, id).Scan(&e.ID, &e.Title, &e.Date, &e.TotalSpots, &e.AvailableSpots, &e.BookingTTLMinutes)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
