package normalize

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

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

func AmountToEurCent(ctx context.Context, redisC *redis.Client, amountCent string, normalizedCurrency string) (int64, error) {
	cents, err := strconv.ParseInt(amountCent, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount cents %q: %w", amountCent, err)
	}

	if normalizedCurrency == "EUR" {
		return cents, nil
	}

	currToEurStr, err := redisC.HGet(ctx, "yaotose:currency", normalizedCurrency).Result()
	if err != nil {
		return 0, fmt.Errorf("get rate for %s: %w", normalizedCurrency, err)
	}

	rate, err := strconv.ParseFloat(currToEurStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rate %q for %s: %w", currToEurStr, normalizedCurrency, err)
	}

	return int64(math.Round(float64(cents) * rate)), nil
}
