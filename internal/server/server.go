package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/network"
)

// ServerStatus represents the current server status
type ServerStatus struct {
	Uptime           time.Duration `json:"uptime"`
	TotalClients     int           `json:"total_clients"`
	ConnectedClients int           `json:"connected_clients"`
	ServerIP         string        `json:"server_ip"`
	TUNInterface     string        `json:"tun_interface"`
	Port             string        `json:"port"`
	Status           string        `json:"status"` // "running", "stopped", "error"
}

// ClientStatus represents real-time client information
type ClientStatus struct {
	ID        uint8     `json:"id"`
	IP        string    `json:"ip"`
	Connected bool      `json:"connected"`
	LastSeen  time.Time `json:"last_seen"`
}

// Server represents the VPN server
type Server struct {
	tunInterface   network.TUNInterface
	keyManager     *crypto.KeyManager
	clientManager  *ClientManager
	packetProcessor *PacketProcessor
	udpConn        *net.UDPConn
	stopChan       chan struct{}
	wg             sync.WaitGroup
	timeout        time.Duration
	startTime      time.Time
	serverIP       string
	port           string
}

// NewServer creates a new VPN server
func NewServer() *Server {
	return &Server{
		stopChan: make(chan struct{}),
		timeout:  30 * time.Minute, // Default timeout
	}
}

// Start starts the VPN server
func (s *Server) Start(configPath, port string) error {
	log.Printf("Starting VPN server...")
	
	// Set server status tracking
	s.startTime = time.Now()
	s.port = port
	
	// Step 1: Load configuration
	err := s.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Step 2: Create TUN interface
	err = s.CreateTUNInterface()
	if err != nil {
		return fmt.Errorf("failed to create TUN interface: %w", err)
	}
	
	// Step 3: Create client manager
	err = s.CreateClientManager()
	if err != nil {
		return fmt.Errorf("failed to create client manager: %w", err)
	}
	
	// Step 4: Create UDP server
	err = s.CreateUDPServer(port)
	if err != nil {
		return fmt.Errorf("failed to create UDP server: %w", err)
	}
	
	// Step 5: Create packet processor
	err = s.CreatePacketProcessor()
	if err != nil {
		return fmt.Errorf("failed to create packet processor: %w", err)
	}
	
	// Step 6: Start packet processing goroutines
	s.startPacketProcessing()
	
	log.Printf("VPN server started on port %s", s.port)
	return nil
}

// startPacketProcessing starts the packet processing goroutines
func (s *Server) startPacketProcessing() {
	// Start client packet handling goroutine
	s.wg.Add(1)
	go s.handleClients()
	
	// Start TUN packet routing goroutine
	s.wg.Add(1)
	go s.routePackets()
	
}

// Stop stops the VPN server
func (s *Server) Stop() error {
	log.Printf("Stopping VPN server...")
	
	// Only close stopChan if it's not already closed
	select {
	case <-s.stopChan:
		// Already closed, do nothing
	default:
		close(s.stopChan)
	}
	
	// Wait for all goroutines to finish
	s.wg.Wait()
	
	// Close UDP connection
	if s.udpConn != nil {
		s.udpConn.Close()
	}
	
	// Close TUN interface
	if s.tunInterface != nil {
		s.tunInterface.Close()
	}
	
	log.Printf("VPN server stopped")
	return nil
}

func (s *Server) GetServerStatus() ServerStatus {
	status := ServerStatus{
		Status: "stopped",
	}
	
	if s.startTime.IsZero() {
		status.Status = "stopped"
		return status
	}
	
	select {
	case <-s.stopChan:
		status.Status = "stopped"
		return status
	default:
		status.Status = "running"
	}
	
	if !s.startTime.IsZero() {
		status.Uptime = time.Since(s.startTime)
	}
	
	if s.clientManager != nil {
		clients := s.clientManager.ListClients()
		status.TotalClients = len(clients)
		
		connectedCount := 0
		for _, client := range clients {
			if client.Connected {
				connectedCount++
			}
		}
		status.ConnectedClients = connectedCount
	}
	
	status.ServerIP = s.serverIP
	status.Port = s.port
	status.TUNInterface = "fvp0"
	
	return status
}

func (s *Server) GetClientStatus() []ClientStatus {
	if s.clientManager == nil {
		return []ClientStatus{}
	}
	
	clients := s.clientManager.ListClients()
	status := make([]ClientStatus, len(clients))
	
	for i, client := range clients {
		status[i] = ClientStatus{
			ID:        client.ID,
			IP:        client.IP,
			Connected: client.Connected,
			LastSeen:  client.LastSeen,
		}
	}
	
	return status
}

func (s *Server) GetPort() string {
	return s.port
}