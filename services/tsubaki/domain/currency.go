package domain

import (
	"sync/atomic"
	"time"
)

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

type CurrencyStore struct {
	ptr atomic.Pointer[Currency]
}

func (c *CurrencyStore) Store(curr *Currency) {
	c.ptr.Store(curr)
}

func (c *CurrencyStore) Load() *Currency {
	return c.ptr.Load()
}

func CurrencyNew(c *Currency) *CurrencyStore {
	s := &CurrencyStore{}
	if c != nil {
		s.ptr.Store(c)
	}
	return s
}
