package domain

import "time"

type FrankfurterRate struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}

type Currency struct {
	Rate      map[string]float64
	UpdatedAt time.Time
}
