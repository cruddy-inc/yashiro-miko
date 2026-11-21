package normalize

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"tsubaki/domain"
)

type Amount struct {
	Currency *domain.Currency
}

func AmountNew(currency *domain.Currency) *Amount {
	return &Amount{
		Currency: currency,
	}
}

// Note on development: I love this strategy so much
func AmountToCent(amount string) string {
	var c strings.Builder
	// We append 2 additional allocation spaces (no reallocation at runtime)
	c.Grow(len(amount) + 2)

	var aftDec int
	var dot bool
	for i := 0; i < len(amount); i++ {
		b := amount[i]
		if b >= '0' && b <= '9' {
			if dot {
				if aftDec == 2 {
					break
				}
				aftDec++
			}
			c.WriteByte(b)
		} else if b == '.' {
			dot = true
		}
	}

	switch aftDec {
	case 0:
		c.WriteString("00")
	case 1:
		c.WriteByte('0')
	}

	return c.String()
}

func (a *Amount) AmountToEurCent(ctx context.Context, amountCent string, normalizedCurrency string) (int64, error) {
	cents, err := strconv.ParseInt(amountCent, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount cents %q: %w", amountCent, err)
	}

	if normalizedCurrency == "EUR" {
		return cents, nil
	}

	rate := a.Currency.Rate[normalizedCurrency]
	if rate == 0.0 {
		return 0.0, fmt.Errorf("unknown currency: %s", normalizedCurrency)
	}

	return int64(math.Round(rate) * float64(cents)), nil
}
