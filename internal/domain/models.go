package domain

type PriceUpdate struct {
	Symbol       string  `json:"pair_name"`
	Exchange     string  `json:"exchange"`
	Timestamp    int64   `json:"timestamp"`
	Price        float64 `json:"price"`
	AveragePrice float64 `json:"average_price"`
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
}
