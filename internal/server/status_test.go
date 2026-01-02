package server

import (
	"testing"
	"time"
)

// TestServerStatus tests the server status functionality
func TestServerStatus(t *testing.T) {
	server := NewServer()

	// Test 1: Status when server is stopped
	t.Run("StatusWhenStopped", func(t *testing.T) {
		// Simulate server being stopped by closing the stopChan
		close(server.stopChan)
		
		status := server.GetServerStatus()
		
		if status.Status != "stopped" {
			t.Errorf("Expected status 'stopped', got '%s'", status.Status)
		}
		
		if status.Uptime != 0 {
			t.Errorf("Expected uptime 0, got %v", status.Uptime)
		}
		
		if status.TotalClients != 0 {
			t.Errorf("Expected 0 total clients, got %d", status.TotalClients)
		}
		
		if status.ConnectedClients != 0 {
			t.Errorf("Expected 0 connected clients, got %d", status.ConnectedClients)
		}
	})

	// Test 2: Status when server is running (simulated)
	t.Run("StatusWhenRunning", func(t *testing.T) {
		// Create a fresh server instance for this test
		runningServer := NewServer()
		
		// Simulate server start
		runningServer.startTime = time.Now().Add(-5 * time.Minute)
		runningServer.port = ":1194"
		runningServer.serverIP = "10.0.0.1"
		
		// Create a mock client manager with some clients
		runningServer.clientManager = &ClientManager{
			clients: make(map[uint8]*Client),
		}
		
		// Add a mock connected client
		runningServer.clientManager.clients[1] = &Client{
			ID:        1,
			IP:        "10.0.0.2",
			Connected: true,
			LastSeen:  time.Now().Add(-1 * time.Minute),
		}
		
		// Add a mock disconnected client
		runningServer.clientManager.clients[2] = &Client{
			ID:        2,
			IP:        "10.0.0.3",
			Connected: false,
			LastSeen:  time.Now().Add(-10 * time.Minute),
		}
		
		status := runningServer.GetServerStatus()
		
		if status.Status != "running" {
			t.Errorf("Expected status 'running', got '%s'", status.Status)
		}
		
		if status.Uptime < 4*time.Minute || status.Uptime > 6*time.Minute {
			t.Errorf("Expected uptime around 5 minutes, got %v", status.Uptime)
		}
		
		if status.Port != ":1194" {
			t.Errorf("Expected port ':1194', got '%s'", status.Port)
		}
		
		if status.TUNInterface != "fvp0" {
			t.Errorf("Expected TUN interface 'fvp0', got '%s'", status.TUNInterface)
		}
		
		if status.TotalClients != 2 {
			t.Errorf("Expected 2 total clients, got %d", status.TotalClients)
		}
		
		if status.ConnectedClients != 1 {
			t.Errorf("Expected 1 connected client, got %d", status.ConnectedClients)
		}
	})
}

// TestClientStatus tests the client status functionality
func TestClientStatus(t *testing.T) {
	// Test 1: No client manager
	t.Run("NoClientManager", func(t *testing.T) {
		server := NewServer()
		clients := server.GetClientStatus()
		
		if len(clients) != 0 {
			t.Errorf("Expected 0 clients, got %d", len(clients))
		}
	})

	// Test 2: With client manager and clients
	t.Run("WithClients", func(t *testing.T) {
		server := NewServer()
		
		// Create a mock client manager
		server.clientManager = &ClientManager{
			clients: make(map[uint8]*Client),
		}
		
		// Add mock clients
		server.clientManager.clients[1] = &Client{
			ID:        1,
			IP:        "10.0.0.2",
			Connected: true,
			LastSeen:  time.Now().Add(-1 * time.Minute),
		}
		
		server.clientManager.clients[2] = &Client{
			ID:        2,
			IP:        "10.0.0.3",
			Connected: false,
			LastSeen:  time.Now().Add(-5 * time.Minute),
		}
		
		clients := server.GetClientStatus()
		
		if len(clients) != 2 {
			t.Errorf("Expected 2 clients, got %d", len(clients))
		}
		
		// Check clients (order-independent)
		clientMap := make(map[uint8]ClientStatus)
		for _, client := range clients {
			clientMap[client.ID] = client
		}
		
		// Check client 1
		client1, exists := clientMap[1]
		if !exists {
			t.Errorf("Expected client 1 to exist")
		} else {
			if client1.IP != "10.0.0.2" {
				t.Errorf("Expected client 1 IP '10.0.0.2', got '%s'", client1.IP)
			}
			if !client1.Connected {
				t.Errorf("Expected client 1 to be connected")
			}
		}
		
		// Check client 2
		client2, exists := clientMap[2]
		if !exists {
			t.Errorf("Expected client 2 to exist")
		} else {
			if client2.IP != "10.0.0.3" {
				t.Errorf("Expected client 2 IP '10.0.0.3', got '%s'", client2.IP)
			}
			if client2.Connected {
				t.Errorf("Expected client 2 to be disconnected")
			}
		}
	})
}
