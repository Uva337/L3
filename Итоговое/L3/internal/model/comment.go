package model

import "time"

type Comment struct {
	ID        string    `json:"id"`
	ParentID  *string   `json:"parent_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`

	Children []*Comment `json:"children,omitempty"`
}
