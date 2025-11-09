package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/network"
	"github.com/pepalonsocosta/fvp/internal/server"
	"gopkg.in/yaml.v3"
)

type CLIServer struct {
	server *server.Server
}

func NewCLIServer() *CLIServer {
	return &CLIServer{
		server: server.NewServer(),
	}
}

type ServerConfig struct {
	Server struct {
		Port           string `yaml:"port"`
		TimeoutMinutes int    `yaml:"timeout_minutes"`
	} `yaml:"server"`
	Clients []crypto.ClientConfig `yaml:"clients"`
}

type ClientInfo struct {
	ID         uint8     `json:"id"`
	IP         string    `json:"ip"`
	LastSeen   time.Time `json:"last_seen"`
	Connected  bool      `json:"connected"`
}

func (s *CLIServer) Setup(port string, timeoutMinutes int) error {
	if _, err := os.Stat("server.yaml"); err == nil {
		return fmt.Errorf("configuration file already exists")
	}

	config := ServerConfig{}
	config.Server.Port = port
	config.Server.TimeoutMinutes = timeoutMinutes
	config.Clients = []crypto.ClientConfig{}

	err := s.writeConfig("server.yaml", &config)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Test TUN interface creation (requires root)
	err = s.testTUNInterface()
	if err != nil {
		fmt.Printf("Warning: TUN interface test failed: %v\n", err)
		fmt.Println("You may need to run with sudo for TUN interface creation")
	}

	return nil
}

func (s *CLIServer) AddClient() (uint8, string, error) {
	config, err := s.loadConfig("server.yaml")
	if err != nil {
		return 0, "", fmt.Errorf("no configuration found, run 'fvps setup' first")
	}

	key, err := s.generateKey()
	if err != nil {
		return 0, "", fmt.Errorf("failed to generate key: %w", err)
	}

	nextID := s.findNextClientID(config.Clients)
	if nextID == 0 {
		return 0, "", fmt.Errorf("maximum clients reached (255)")
	}

	client := crypto.ClientConfig{
		ID:  nextID,
		Key: key,
	}
	config.Clients = append(config.Clients, client)

	err = s.writeConfig("server.yaml", config)
	if err != nil {
		return 0, "", fmt.Errorf("failed to update config: %w", err)
	}

	return nextID, key, nil
}

func (s *CLIServer) ListClients() ([]ClientInfo, error) {
	config, err := s.loadConfig("server.yaml")
	if err != nil {
		return nil, fmt.Errorf("no configuration found, run 'fvps setup' first")
	}

	clients := make([]ClientInfo, len(config.Clients))
	for i, client := range config.Clients {
		clients[i] = ClientInfo{
			ID:        client.ID,
			IP:        s.getClientIP(client.ID),
			LastSeen:  time.Time{}, // Not available from config
			Connected: false,        // Not available from config
		}
	}

	return clients, nil
}

func (s *CLIServer) RemoveClient(clientID uint8) error {
	config, err := s.loadConfig("server.yaml")
	if err != nil {
		return fmt.Errorf("no configuration found, run 'fvps setup' first")
	}

	found := false
	for i, client := range config.Clients {
		if client.ID == clientID {
			config.Clients = append(config.Clients[:i], config.Clients[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("client %d not found", clientID)
	}

	err = s.writeConfig("server.yaml", config)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}


func (s *CLIServer) loadConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *CLIServer) writeConfig(path string, config *ServerConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *CLIServer) generateKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func (s *CLIServer) findNextClientID(clients []crypto.ClientConfig) uint8 {
	used := make(map[uint8]bool)
	for _, client := range clients {
		used[client.ID] = true
	}

	for i := uint8(1); i != 0; i++ {
		if !used[i] {
			return i
		}
	}
	return 0
}

func (s *CLIServer) getClientIP(clientID uint8) string {
	return fmt.Sprintf("10.0.0.%d", clientID+1)
}

func (s *CLIServer) testTUNInterface() error {
	tunManager := network.NewTunManager()
	err := tunManager.Create("fvp-test")
	if err != nil {
		return err
	}
	
	tunManager.Close()
	return nil
}

func (s *CLIServer) Status() error {
	status := s.server.GetServerStatus()
	
	fmt.Println("Server Status:")
	fmt.Printf("  Status: %s\n", status.Status)
	if status.Status == "running" {
		fmt.Printf("  Uptime: %v\n", status.Uptime.Round(time.Second))
		fmt.Printf("  Port: %s\n", status.Port)
		fmt.Printf("  TUN Interface: %s\n", status.TUNInterface)
		fmt.Printf("  Total Clients: %d\n", status.TotalClients)
		fmt.Printf("  Connected Clients: %d\n", status.ConnectedClients)
	}
	
	return nil
}

func (s *CLIServer) ListClientsRealtime() ([]server.ClientStatus, error) {
	clients := s.server.GetClientStatus()
	if len(clients) > 0 {
		return clients, nil
	}
	
	// Fallback to config file if server not running
	configClients, err := s.ListClients()
	if err != nil {
		return nil, err
	}
	
	realtimeClients := make([]server.ClientStatus, len(configClients))
	for i, client := range configClients {
		realtimeClients[i] = server.ClientStatus{
			ID:        client.ID,
			IP:        client.IP,
			Connected: client.Connected,
			LastSeen:  client.LastSeen,
		}
	}
	
	return realtimeClients, nil
}
