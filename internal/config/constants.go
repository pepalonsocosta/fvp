package config

import (
	"os"
	"path/filepath"
	"time"
)

const (
	// ConfigDirName is the name of the config directory in user's home
	ConfigDirName = ".fvp"
	// ServerConfigFileName is the server config filename
	ServerConfigFileName = "server.yaml"
	// ClientConfigFileName is the client config filename
	ClientConfigFileName = "client.yaml"

	// Network constants
	DefaultMTU          = 1500 // Standard Ethernet MTU
	DefaultTUNInterface = "fvp0"
	DefaultServerPort  = ":1194"

	// IP address range for VPN clients
	VPNNetworkBase = "10.0.0"
	VPNServerIP   = "10.0.0.1"
	VPNClientIPStart = 2  // 10.0.0.2
	VPNClientIPEnd   = 255 // 10.0.0.255

	// Timeouts
	DefaultClientTimeout      = 30 * time.Minute
	DefaultUDPReadTimeout     = 1 * time.Second
	DefaultClientUDPTimeout   = 10 * time.Second
	DefaultTUNReadDelay       = 10 * time.Millisecond
	DefaultKeepAliveInterval  = 30 * time.Second
	DefaultTimeoutCheckInterval = 1 * time.Minute

	// Client limits
	MaxClients = 256 // Maximum number of concurrent clients (0-255)
)

// GetConfigDir returns the path to the .fvp config directory in user's home
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigDirName), nil
}

// GetServerConfigPath returns the default server config path
func GetServerConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ServerConfigFileName), nil
}

// GetClientConfigPath returns the default client config path
func GetClientConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ClientConfigFileName), nil
}

// EnsureConfigDir creates the .fvp directory if it doesn't exist with 0700 permissions
func EnsureConfigDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return "", err
	}
	
	return configDir, nil
}

