package normalize

import "strings"

func Upper(s string) string {
	var c strings.Builder
	c.Grow(len(s))

	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 'A' && b <= 'Z' {
			c.WriteByte(b)
		}
		if b >= 'a' && b <= 'z' {
			c.WriteByte(b - 32)
		}
	}

	return c.String()
}
