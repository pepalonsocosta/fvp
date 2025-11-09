package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/network"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Server struct {
		Port           string `yaml:"port"`
		TimeoutMinutes int    `yaml:"timeout_minutes"`
	} `yaml:"server"`
	Clients []crypto.ClientConfig `yaml:"clients"`
}

func (s *Server) LoadConfig(configPath string) error {
	s.keyManager = crypto.NewKeyManager()
	
	err := s.keyManager.LoadKeysFromConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	err = s.loadServerSettings(configPath)
	if err != nil {
		return fmt.Errorf("failed to load server settings: %w", err)
	}
	
	log.Printf("Configuration loaded successfully")
	return nil
}

func (s *Server) loadServerSettings(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config ServerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	if config.Server.TimeoutMinutes > 0 {
		s.timeout = time.Duration(config.Server.TimeoutMinutes) * time.Minute
	}
	
	if config.Server.Port != "" {
		s.port = config.Server.Port
	}

	return nil
}

func (s *Server) CreateTUNInterface() error {
	tunManager := network.NewTunManager()
	
	err := tunManager.Create("fvp0")
	if err != nil {
		return fmt.Errorf("failed to create TUN interface: %w", err)
	}
	
	s.tunInterface = tunManager
	log.Printf("Created TUN interface: %s", tunManager.GetName())
	return nil
}

func (s *Server) CreateClientManager() error {
	if s.keyManager == nil {
		return fmt.Errorf("key manager not initialized")
	}
	s.clientManager = NewClientManager(s.keyManager)
	log.Printf("Created client manager")
	return nil
}

func (s *Server) CreatePacketProcessor() error {
	if s.tunInterface == nil || s.keyManager == nil || s.clientManager == nil || s.udpConn == nil {
		return fmt.Errorf("required components not initialized")
	}
	s.packetProcessor = NewPacketProcessor(s.tunInterface, s.keyManager, s.clientManager, s.udpConn)
	log.Printf("Created packet processor")
	return nil
}

func (s *Server) CreateUDPServer(port string) error {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}
	
	s.udpConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to create UDP server: %w", err)
	}
	
	log.Printf("UDP server listening on %s", port)
	return nil
}
