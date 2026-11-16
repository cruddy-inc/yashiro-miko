package normalize

import (
	"fmt"
	"strings"
)

func NormalizeIBAN(text string) (string, error) {
	var c strings.Builder
	c.Grow(len(text))

	for _, s := range text {
		switch {
		case s >= 'A' && s <= 'Z':
			c.WriteRune(s)
		case s >= 'a' && s <= 'z':
			c.WriteRune(s - 32)
		case s >= '0' && s <= '9':
			c.WriteRune(s)
		case s == ' ' || s == '-':
		default:
			return "", fmt.Errorf("Invalid character passed: %s", string(s))
		}
	}

	if len(c.String()) < 14 || len(c.String()) > 34 {
		return "", fmt.Errorf("Invalid IBAN length: %d", len(c.String()))
	}

	return c.String(), nil
}
