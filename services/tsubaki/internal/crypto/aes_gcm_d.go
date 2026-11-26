package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func Decrypt(secretkey string, ciphertext []byte) (string, error) {
	block, err := aes.NewCipher([]byte(secretkey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("Given ciphertext is too short: %d", len(ciphertext))
	}

	// Since the nonce is the first part - we simply can strip it to gain it back / same for the 'sealed'
	nonce, sealed := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}
