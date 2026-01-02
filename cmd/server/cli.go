package main

import (
	"fmt"
	"time"

	"github.com/pepalonsocosta/fvp/internal/config"
	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/network"
	"github.com/pepalonsocosta/fvp/internal/server"
)

type CLIServer struct {
	server     *server.Server
	configMgr  *config.Manager
}

func NewCLIServer() (*CLIServer, error) {
	configMgr := config.NewManagerWithPath("server.yaml")

	return &CLIServer{
		server:    server.NewServer(),
		configMgr: configMgr,
	}, nil
}

type ClientInfo struct {
	ID         uint8     `json:"id"`
	IP         string    `json:"ip"`
	LastSeen   time.Time `json:"last_seen"`
	Connected  bool      `json:"connected"`
}

func (s *CLIServer) Setup(port string, timeoutMinutes int) error {
	if s.configMgr.Exists() {
		return fmt.Errorf("configuration file already exists")
	}

	cfg := &config.ServerConfig{}
	cfg.Server.Port = port
	cfg.Server.TimeoutMinutes = timeoutMinutes
	cfg.Clients = []crypto.ClientConfig{}

	err := s.configMgr.Save(cfg)
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
	cfg, err := s.configMgr.Load()
	if err != nil {
		return 0, "", fmt.Errorf("no configuration found, run 'fvps setup' first: %w", err)
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		return 0, "", fmt.Errorf("failed to generate key: %w", err)
	}

	nextID := s.findNextClientID(cfg.Clients)
	if nextID == 0 {
		return 0, "", fmt.Errorf("maximum clients reached (%d)", config.MaxClients-1)
	}

	client := crypto.ClientConfig{
		ID:  nextID,
		Key: key,
	}
	cfg.Clients = append(cfg.Clients, client)

	err = s.configMgr.Save(cfg)
	if err != nil {
		return 0, "", fmt.Errorf("failed to update config: %w", err)
	}

	return nextID, key, nil
}

func (s *CLIServer) ListClients() ([]ClientInfo, error) {
	cfg, err := s.configMgr.Load()
	if err != nil {
		return nil, fmt.Errorf("no configuration found, run 'fvps setup' first: %w", err)
	}

	clients := make([]ClientInfo, len(cfg.Clients))
	for i, client := range cfg.Clients {
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
	cfg, err := s.configMgr.Load()
	if err != nil {
		return fmt.Errorf("no configuration found, run 'fvps setup' first: %w", err)
	}

	found := false
	for i, client := range cfg.Clients {
		if client.ID == clientID {
			cfg.Clients = append(cfg.Clients[:i], cfg.Clients[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("client %d not found", clientID)
	}

	err = s.configMgr.Save(cfg)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}



func (s *CLIServer) findNextClientID(clients []crypto.ClientConfig) uint8 {
	used := make(map[uint8]bool)
	for _, client := range clients {
		used[client.ID] = true
	}

	// Start from 1, go up to MaxClients-1 (255)
	// MaxClients is 256, so we check up to 255 (uint8 max)
	for i := uint8(1); i != 0; i++ {
		if !used[i] {
			return i
		}
	}
	return 0
}

func (s *CLIServer) getClientIP(clientID uint8) string {
	return fmt.Sprintf("%s.%d", config.VPNNetworkBase, clientID+1)
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
