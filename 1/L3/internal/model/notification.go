package model

import "time"

const (
	StatusPending   = "pending"
	StatusSent      = "sent"
	StatusFailed    = "failed"
	StatusCancellde = "cancelled"
)

type Notification struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"`
	Recipient   string    `json:"recipient"`
	Channel     string    `json:"channel"`
	Status      string    `json:"status"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CreatedAt   time.Time `json:"created_at"`
}
