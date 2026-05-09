package model

import "time"

// URL описывает сущность короткой ссылки
type URL struct {
	ID          string    `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// Click описывает один факт перехода по короткой ссылке (для аналитики)
type Click struct {
	ID        string    `json:"id"`
	URLID     string    `json:"url_id"`
	UserAgent string    `json:"user_agent"`
	ClickedAt time.Time `json:"clicked_at"`
}
