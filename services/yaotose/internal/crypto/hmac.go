package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
)

func HMAC(secretKey []byte, normalizedIBAN string) []byte {
	match := hmac.New(sha256.New, secretKey)
	match.Write([]byte(normalizedIBAN))
	return match.Sum(nil)
}
