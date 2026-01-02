package config

// ClientConfig represents the client configuration file structure
type ClientConfig struct {
	ClientID uint8  `yaml:"client_id"`
	Key      string `yaml:"key"`
	Server   string `yaml:"server"`
}

