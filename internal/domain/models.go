package domain

import "time"

type PriceUpdate struct {
	Symbol    string    `json:"pair_name"`
	Exchange  string    `json:"exchange"`
	Timestamp time.Time `json:"timestamp"`
	Price     float64   `json:"price"`
}

type PriceStats struct {
	Exchange     string    `json:"exchange"`
	Symbol       string    `json:"pair_name"`
	Timestamp    time.Time `json:"timestamp"`
	AveragePrice float64   `json:"average_price"`
	MinPrice     float64   `json:"min_price"`
	MaxPrice     float64   `json:"max_price"`
}
