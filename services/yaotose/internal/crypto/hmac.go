package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
)

func HMAC(secretKey []byte, normalizedIBAN string) string {
	match := hmac.New(sha256.New, secretKey)
	match.Write([]byte(normalizedIBAN))
	return string(match.Sum(nil))
}
