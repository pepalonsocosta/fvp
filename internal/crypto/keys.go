package crypto

import (
	"encoding/hex"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	ID  uint8  `yaml:"id"`
	Key string `yaml:"key"`
}

type Config struct {
	Clients []ClientConfig `yaml:"clients"`
}

type KeyManager struct {
	keys map[uint8][]byte
}

func NewKeyManager() *KeyManager {
	return &KeyManager{
		keys: make(map[uint8][]byte),
	}
}

func (km *KeyManager) LoadKeysFromConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	km.keys = make(map[uint8][]byte)

	for _, client := range config.Clients {
		key, err := hex.DecodeString(client.Key)
		if err != nil {
			return fmt.Errorf("invalid hex key for client %d: %w", client.ID, err)
		}

		if len(key) != 32 {
			return fmt.Errorf("key for client %d must be exactly 32 bytes (64 hex chars), got %d bytes", client.ID, len(key))
		}

		km.keys[client.ID] = key
	}

	return nil
}

func (km *KeyManager) GetClientKey(clientID uint8) ([]byte, error) {
	key, exists := km.keys[clientID]
	if !exists {
		return nil, ErrKeyNotFound
	}
	return key, nil
}

func (km *KeyManager) HasClient(clientID uint8) bool {
	_, exists := km.keys[clientID]
	return exists
}

// SetTestKey sets a test key for testing purposes
func (km *KeyManager) SetTestKey(clientID uint8, key []byte) {
	if km.keys == nil {
		km.keys = make(map[uint8][]byte)
	}
	km.keys[clientID] = key
}
