package crypto

import (
	"testing"
)

func TestEncryptDecryptPayload(t *testing.T) {
	// Test data
	payload := []byte("Hello, World!")
	key := make([]byte, 32) // 32-byte key
	for i := range key {
		key[i] = byte(i)
	}
	sequence := uint32(12345)

	// Encrypt
	encrypted, err := EncryptPayload(payload, key, sequence)
	if err != nil {
		t.Fatalf("EncryptPayload failed: %v", err)
	}

	// Verify encrypted data is different from original
	if len(encrypted) == len(payload) {
		t.Error("Encrypted data should be longer than original due to authentication tag")
	}

	// Decrypt
	decrypted, err := DecryptPayload(encrypted, key, sequence)
	if err != nil {
		t.Fatalf("DecryptPayload failed: %v", err)
	}

	// Verify decrypted data matches original
	if string(decrypted) != string(payload) {
		t.Errorf("Decrypted data doesn't match original: got %s, want %s", string(decrypted), string(payload))
	}
}

func TestEncryptPayloadInvalidKey(t *testing.T) {
	payload := []byte("test")
	invalidKey := []byte("short") // Too short
	sequence := uint32(1)

	// With invalid key length, chacha20poly1305.New should fail
	_, err := EncryptPayload(payload, invalidKey, sequence)
	if err == nil {
		t.Error("Expected error for invalid key length")
	}
}

func TestDecryptPayloadInvalidKey(t *testing.T) {
	encrypted := []byte("some encrypted data")
	invalidKey := []byte("short") // Too short
	sequence := uint32(1)

	// With invalid key length, chacha20poly1305.New should fail
	_, err := DecryptPayload(encrypted, invalidKey, sequence)
	if err == nil {
		t.Error("Expected error for invalid key length")
	}
}

func TestDecryptPayloadWrongKey(t *testing.T) {
	// First encrypt with one key
	payload := []byte("test data")
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
		key2[i] = byte(i + 1) // Different key
	}
	sequence := uint32(1)

	encrypted, err := EncryptPayload(payload, key1, sequence)
	if err != nil {
		t.Fatalf("EncryptPayload failed: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = DecryptPayload(encrypted, key2, sequence)
	if err != ErrDecryptionFailed {
		t.Errorf("Expected ErrDecryptionFailed, got %v", err)
	}
}

func TestDecryptPayloadWrongSequence(t *testing.T) {
	// Encrypt with one sequence
	payload := []byte("test data")
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	sequence1 := uint32(1)
	sequence2 := uint32(2) // Different sequence

	encrypted, err := EncryptPayload(payload, key, sequence1)
	if err != nil {
		t.Fatalf("EncryptPayload failed: %v", err)
	}

	// Try to decrypt with wrong sequence
	_, err = DecryptPayload(encrypted, key, sequence2)
	if err != ErrDecryptionFailed {
		t.Errorf("Expected ErrDecryptionFailed, got %v", err)
	}
}

func TestGenerateNonce(t *testing.T) {
	tests := []struct {
		name     string
		sequence uint32
		expected []byte
	}{
		{
			name:     "sequence 0",
			sequence: 0,
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "sequence 1",
			sequence: 1,
			expected: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "sequence 0x12345678",
			sequence: 0x12345678,
			expected: []byte{0x78, 0x56, 0x34, 0x12, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "sequence 0xFFFFFFFF",
			sequence: 0xFFFFFFFF,
			expected: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateNonce(tt.sequence)
			
			if len(result) != 12 {
				t.Errorf("Expected nonce length 12, got %d", len(result))
			}
			
			for i, b := range result {
				if b != tt.expected[i] {
					t.Errorf("Byte %d mismatch: got 0x%02X, want 0x%02X", i, b, tt.expected[i])
				}
			}
		})
	}
}
