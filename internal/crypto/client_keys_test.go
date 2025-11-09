package crypto

import (
	"testing"
)

func TestClientKeyManager(t *testing.T) {
	clientID := uint8(42)
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	ckm := NewClientKeyManager(clientID, key)

	// Test GetClientID
	if ckm.GetClientID() != clientID {
		t.Errorf("Expected client ID %d, got %d", clientID, ckm.GetClientID())
	}

	// Test GetKey
	retrievedKey := ckm.GetKey()
	if len(retrievedKey) != len(key) {
		t.Errorf("Expected key length %d, got %d", len(key), len(retrievedKey))
	}
	for i, b := range retrievedKey {
		if b != key[i] {
			t.Errorf("Key byte %d mismatch: expected %d, got %d", i, key[i], b)
		}
	}
}

func TestClientKeyManagerEdgeCases(t *testing.T) {
	// Test with minimum client ID
	ckm := NewClientKeyManager(0, []byte{1, 2, 3, 4})
	if ckm.GetClientID() != 0 {
		t.Errorf("Expected client ID 0, got %d", ckm.GetClientID())
	}

	// Test with maximum client ID
	ckm = NewClientKeyManager(255, []byte{255, 254, 253, 252})
	if ckm.GetClientID() != 255 {
		t.Errorf("Expected client ID 255, got %d", ckm.GetClientID())
	}

	// Test with empty key
	emptyKey := []byte{}
	ckm = NewClientKeyManager(1, emptyKey)
	retrievedKey := ckm.GetKey()
	if len(retrievedKey) != 0 {
		t.Errorf("Expected empty key, got %d bytes", len(retrievedKey))
	}

	// Test with large key
	largeKey := make([]byte, 1000)
	for i := range largeKey {
		largeKey[i] = byte(i % 256)
	}
	ckm = NewClientKeyManager(1, largeKey)
	retrievedKey = ckm.GetKey()
	if len(retrievedKey) != len(largeKey) {
		t.Errorf("Expected key length %d, got %d", len(largeKey), len(retrievedKey))
	}
}

func TestClientKeyManagerKeyModification(t *testing.T) {
	originalKey := []byte{1, 2, 3, 4, 5}
	ckm := NewClientKeyManager(1, originalKey)

	// Modify the original key
	originalKey[0] = 99

	// The ClientKeyManager should still have the original key
	retrievedKey := ckm.GetKey()
	if retrievedKey[0] != 1 {
		t.Errorf("Expected key to be unchanged, got %d", retrievedKey[0])
	}

	// Modify the retrieved key (should not affect the original)
	retrievedKey[0] = 88

	// The ClientKeyManager should still have the original key
	retrievedKey2 := ckm.GetKey()
	if retrievedKey2[0] != 1 {
		t.Errorf("Expected key to be unchanged after modification, got %d", retrievedKey2[0])
	}
}
