package config

import (
	"github.com/pepalonsocosta/fvp/internal/crypto"
)

// ServerConfig represents the server configuration file structure
type ServerConfig struct {
	Server struct {
		Port           string `yaml:"port"`
		TimeoutMinutes int    `yaml:"timeout_minutes"`
	} `yaml:"server"`
	Clients []crypto.ClientConfig `yaml:"clients"`
}

