package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manager handles loading and saving server configuration files
type Manager struct {
	configPath string
}

// NewManager creates a new config manager with the default server config path
func NewManager() (*Manager, error) {
	path, err := GetServerConfigPath()
	if err != nil {
		return nil, err
	}
	return NewManagerWithPath(path), nil
}

// NewManagerWithPath creates a new config manager with a custom config path
func NewManagerWithPath(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Load loads the server configuration from file
func (m *Manager) Load() (*ServerConfig, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ServerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the server configuration to file
func (m *Manager) Save(config *ServerConfig) error {
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

// Exists checks if the configuration file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}

// GetPath returns the current config file path
func (m *Manager) GetPath() string {
	return m.configPath
}

