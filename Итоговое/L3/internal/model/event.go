package model

import "time"

// описывает мероприятие
type Event struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Date              time.Time `json:"date"`
	TotalSpots        int       `json:"total_spots"`
	AvailableSpots    int       `json:"available_spots"`
	BookingTTLMinutes int       `json:"booking_ttl_minutes"`
	CreatedAt         time.Time `json:"created_at"`
}

// описывает бронь места пользователем
type Booking struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"` // pending, confirmed, cancelled
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
