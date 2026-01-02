package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateKey generates a random 32-byte key and returns it as a hex-encoded string
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return hex.EncodeToString(key), nil
}

// GenerateKeyBytes generates a random 32-byte key and returns it as bytes
func GenerateKeyBytes() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}

