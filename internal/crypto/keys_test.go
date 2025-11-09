package crypto

import (
	"os"
	"testing"
)

func TestKeyManager(t *testing.T) {
	// Create a temporary config file
	configContent := `clients:
  - id: 1
    key: "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
  - id: 2
    key: "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
  - id: 255
    key: "1111111111111111111111111111111111111111111111111111111111111111"
`

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Test key manager
	km := NewKeyManager()

	// Load keys
	err = km.LoadKeysFromConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadKeysFromConfig failed: %v", err)
	}

	// Test getting existing keys
	key1, err := km.GetClientKey(1)
	if err != nil {
		t.Fatalf("GetClientKey(1) failed: %v", err)
	}
	if len(key1) != 32 {
		t.Errorf("Key length should be 32, got %d", len(key1))
	}

	key2, err := km.GetClientKey(2)
	if err != nil {
		t.Fatalf("GetClientKey(2) failed: %v", err)
	}
	if len(key2) != 32 {
		t.Errorf("Key length should be 32, got %d", len(key2))
	}

	// Test getting non-existent key
	_, err = km.GetClientKey(99)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	// Test HasClient
	if !km.HasClient(1) {
		t.Error("HasClient(1) should return true")
	}
	if !km.HasClient(2) {
		t.Error("HasClient(2) should return true")
	}
	if km.HasClient(99) {
		t.Error("HasClient(99) should return false")
	}
}

func TestKeyManagerInvalidConfig(t *testing.T) {
	km := NewKeyManager()

	// Test with non-existent file
	err := km.LoadKeysFromConfig("non_existent_file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with invalid YAML
	tmpFile, err := os.CreateTemp("", "test_invalid_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("invalid: yaml: content: ["); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	err = km.LoadKeysFromConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestKeyManagerInvalidKey(t *testing.T) {
	// Test with invalid hex key
	configContent := `clients:
  - id: 1
    key: "invalid_hex_key"
`

	tmpFile, err := os.CreateTemp("", "test_invalid_key_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	km := NewKeyManager()
	err = km.LoadKeysFromConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid hex key")
	}
}

func TestKeyManagerWrongKeyLength(t *testing.T) {
	// Test with wrong key length
	configContent := `clients:
  - id: 1
    key: "a1b2c3d4e5f6"  # Too short
`

	tmpFile, err := os.CreateTemp("", "test_wrong_length_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	km := NewKeyManager()
	err = km.LoadKeysFromConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for wrong key length")
	}
}
