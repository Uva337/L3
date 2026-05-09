package model

import "time"

// Item описывает товар на складе
type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

// ItemHistory описывает запись аудита
type ItemHistory struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"item_id"`
	Action    string    `json:"action"`
	OldData   string    `json:"old_data"`
	NewData   string    `json:"new_data"`
	ChangedBy string    `json:"changed_by"`
	ChangedAt time.Time `json:"changed_at"`
}
