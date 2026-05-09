package model

import "time"

// Sale описывает одну запись дохода или расхода
type Sale struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // income или expense
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	SaleDate  time.Time `json:"sale_date"`
	CreatedAt time.Time `json:"created_at"`
}

// Pезультат аналитического запроса
type AnalyticsResult struct {
	Count        int     `json:"count"`
	TotalSum     float64 `json:"total_sum"`
	Average      float64 `json:"average"`
	Median       float64 `json:"median"`
	Percentile90 float64 `json:"percentile_90"`
}
