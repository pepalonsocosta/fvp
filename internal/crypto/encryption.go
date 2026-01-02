package crypto

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

var (
	ErrInvalidKeyLength = errors.New("key must be exactly 32 bytes")
	ErrDecryptionFailed = errors.New("decryption failed - invalid data or key")
	ErrKeyNotFound      = errors.New("client key not found")
)

type CryptoError struct {
	Operation string
	Err       error
}

func (e *CryptoError) Error() string {
	return fmt.Sprintf("crypto %s failed: %v", e.Operation, e.Err)
}

func (e *CryptoError) Unwrap() error {
	return e.Err
}

func EncryptPayload(payload []byte, key []byte, sequence uint32) ([]byte, error) {
	cipher, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, &CryptoError{Operation: "encryption", Err: err}
	}

	nonce := GenerateNonce(sequence)
	encrypted := cipher.Seal(nil, nonce, payload, nil)
	
	return encrypted, nil
}

func DecryptPayload(encryptedPayload []byte, key []byte, sequence uint32) ([]byte, error) {
	cipher, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, &CryptoError{Operation: "decryption", Err: err}
	}

	nonce := GenerateNonce(sequence)
	decrypted, err := cipher.Open(nil, nonce, encryptedPayload, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	
	return decrypted, nil
}

// Generate nonce from sequence number (32 bits + 8 zero bytes = 12 bytes)
func GenerateNonce(sequence uint32) []byte {
	nonce := make([]byte, chacha20poly1305.NonceSize)
	nonce[0] = byte(sequence)
	nonce[1] = byte(sequence >> 8)
	nonce[2] = byte(sequence >> 16)
	nonce[3] = byte(sequence >> 24)
	return nonce
}