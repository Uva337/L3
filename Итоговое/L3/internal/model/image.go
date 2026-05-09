package model

import "time"

type Image struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	Status    string    `json:"status"` // pending, processing, ready, failed
	Format    string    `json:"format"` // jpg, png, gif
	CreatedAt time.Time `json:"created_at"`
}
