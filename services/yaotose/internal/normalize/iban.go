package normalize

import (
	"fmt"
	"strings"
)

func NormalizeIBAN(text string) (string, error) {
	var c strings.Builder
	c.Grow(len(text))

	for i := 0; i < len(text); i++ {
		b := text[i]
		switch {
		case b >= 'A' && b <= 'Z':
			c.WriteByte(b)
		case b >= 'a' && b <= 'z':
			c.WriteByte(b - 32)
		case b >= '0' && b <= '9':
			c.WriteByte(b)
		case b == ' ' || b == '-':
		default:
			return "", fmt.Errorf("Invalid character passed: %b", string(b))
		}
	}

	if c.Len() < 14 || c.Len() > 34 {
		return "", fmt.Errorf("Invalid IBAN length: %d", len(c.String()))
	}

	return c.String(), nil
}
