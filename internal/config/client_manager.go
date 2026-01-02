package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ClientManager struct {
	configPath string
}

func NewClientManager() (*ClientManager, error) {
	path, err := GetClientConfigPath()
	if err != nil {
		return nil, err
	}
	return NewClientManagerWithPath(path), nil
}

func NewClientManagerWithPath(configPath string) *ClientManager {
	return &ClientManager{
		configPath: configPath,
	}
}

func (m *ClientManager) Load() (*ClientConfig, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ClientConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (m *ClientManager) Save(config *ClientConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(m.configPath, data, 0600)
}

func (m *ClientManager) Exists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}

func (m *ClientManager) GetPath() string {
	return m.configPath
}

func (m *ClientManager) CreateTemplate() error {
	template := `# FVP Client Configuration
# Fill in the values below and run 'fvpc connect' again

client_id: 0  # Client ID from server (run 'fvps add-client' on server)
key: ""       # 64-character hex key from server
server: ""    # Server address (e.g., 192.168.1.100:1194)
`

	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(m.configPath, []byte(template), 0600)
}

