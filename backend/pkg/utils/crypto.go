package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var encryptionKey []byte

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Warn().Msg("Error loading .env file in crypto utils")
	}

	key := os.Getenv("AI_CREDENTIALS_ENC_KEY")
	if key == "" {
		key = "MyDefaultEncryptionKey32BytesLong"
	}

	encryptionKey = []byte(key)
	if len(encryptionKey) != 32 {
		if len(encryptionKey) < 32 {
			// Pad to 32 bytes
			padding := make([]byte, 32-len(encryptionKey))
			encryptionKey = append(encryptionKey, padding...)
		} else {
			encryptionKey = encryptionKey[:32]
		}
	}
}

// Encrypt encrypts plaintext using AES-GCM
func Encrypt(plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-GCM
func Decrypt(ciphertext []byte) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptString encrypts a string and returns a base64-encoded string
func EncryptString(plaintext string) (string, error) {
	ciphertext, err := Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a base64-encoded string
func DecryptString(encodedCiphertext string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}
	return Decrypt(ciphertext)
}

// MaskAPIKey returns a masked version of the API key for display
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}

	// Show first 4 and last 4 characters
	prefix := apiKey[:4]
	suffix := apiKey[len(apiKey)-4:]
	return fmt.Sprintf("%s****%s", prefix, suffix)
}
