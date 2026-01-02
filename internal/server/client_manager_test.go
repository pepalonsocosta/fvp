package server

import (
	"fmt"
	"testing"

	"github.com/pepalonsocosta/fvp/internal/crypto"
)

func TestClientManager_AddClient(t *testing.T) {
	keyManager := crypto.NewKeyManager()
	cm := NewClientManager(keyManager)

	// Test adding first client
	key1 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
	}

	client1, err := cm.AddClient(key1, "192.168.1.100:12345")
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	if client1.ID != 1 {
		t.Errorf("Expected client ID 1, got %d", client1.ID)
	}

	if client1.IP != "10.0.0.2" {
		t.Errorf("Expected IP 10.0.0.2, got %s", client1.IP)
	}

	if !client1.Connected {
		t.Error("Client should be connected")
	}

	// Test adding second client
	key2 := make([]byte, 32)
	for i := range key2 {
		key2[i] = byte(i + 1)
	}

	client2, err := cm.AddClient(key2, "192.168.1.101:12346")
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	if client2.ID != 2 {
		t.Errorf("Expected client ID 2, got %d", client2.ID)
	}

	if client2.IP != "10.0.0.3" {
		t.Errorf("Expected IP 10.0.0.3, got %s", client2.IP)
	}

	// Test adding duplicate key
	_, err = cm.AddClient(key1, "192.168.1.102:12347")
	if err != ErrClientAlreadyExists {
		t.Errorf("Expected ErrClientAlreadyExists, got %v", err)
	}
}

func TestClientManager_RemoveClient(t *testing.T) {
	keyManager := crypto.NewKeyManager()
	cm := NewClientManager(keyManager)

	// Add a client
	key := make([]byte, 32)
	client, err := cm.AddClient(key, "192.168.1.100:12345")
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	// Remove the client
	err = cm.RemoveClient(client.ID)
	if err != nil {
		t.Fatalf("RemoveClient failed: %v", err)
	}

	// Try to get the client
	_, err = cm.GetClient(client.ID)
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}

	// Try to remove non-existent client
	err = cm.RemoveClient(99)
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestClientManager_GetClient(t *testing.T) {
	keyManager := crypto.NewKeyManager()
	cm := NewClientManager(keyManager)

	// Add a client
	key := make([]byte, 32)
	client, err := cm.AddClient(key, "192.168.1.100:12345")
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	// Get the client
	retrievedClient, err := cm.GetClient(client.ID)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}

	if retrievedClient.ID != client.ID {
		t.Errorf("Expected client ID %d, got %d", client.ID, retrievedClient.ID)
	}

	// Try to get non-existent client
	_, err = cm.GetClient(99)
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestClientManager_GetClientByIP(t *testing.T) {
	keyManager := crypto.NewKeyManager()
	cm := NewClientManager(keyManager)

	// Add a client
	key := make([]byte, 32)
	client, err := cm.AddClient(key, "192.168.1.100:12345")
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	// Get client by IP
	retrievedClient, err := cm.GetClientByIP(client.IP)
	if err != nil {
		t.Fatalf("GetClientByIP failed: %v", err)
	}

	if retrievedClient.ID != client.ID {
		t.Errorf("Expected client ID %d, got %d", client.ID, retrievedClient.ID)
	}

	// Try to get client by non-existent IP
	_, err = cm.GetClientByIP("10.0.0.99")
	if err != ErrClientNotFound {
		t.Errorf("Expected ErrClientNotFound, got %v", err)
	}
}

func TestClientManager_UpdateClientActivity(t *testing.T) {
	keyManager := crypto.NewKeyManager()
	cm := NewClientManager(keyManager)

	// Add a client
	key := make([]byte, 32)
	client, err := cm.AddClient(key, "192.168.1.100:12345")
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	// Update activity with valid sequence
	err = cm.UpdateClientActivity(client.ID, 1)
	if err != nil {
		t.Fatalf("UpdateClientActivity failed: %v", err)
	}

	// Try to update with same sequence (should fail)
	err = cm.UpdateClientActivity(client.ID, 1)
	if err != ErrInvalidSequence {
		t.Errorf("Expected ErrInvalidSequence, got %v", err)
	}

	// Try to update with lower sequence (should fail)
	err = cm.UpdateClientActivity(client.ID, 0)
	if err != ErrInvalidSequence {
		t.Errorf("Expected ErrInvalidSequence, got %v", err)
	}

	// Update with higher sequence (should succeed)
	err = cm.UpdateClientActivity(client.ID, 5)
	if err != nil {
		t.Fatalf("UpdateClientActivity failed: %v", err)
	}
}

func TestClientManager_ListClients(t *testing.T) {
	keyManager := crypto.NewKeyManager()
	cm := NewClientManager(keyManager)

	// Initially no clients
	clients := cm.ListClients()
	if len(clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(clients))
	}

	// Add some clients
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
		key2[i] = byte(i + 1)
	}

	_, err := cm.AddClient(key1, "192.168.1.100:12345")
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	_, err = cm.AddClient(key2, "192.168.1.101:12346")
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	// Check client count
	clients = cm.ListClients()
	if len(clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients))
	}
}

func TestClientManager_IPAssignment(t *testing.T) {
	keyManager := crypto.NewKeyManager()
	cm := NewClientManager(keyManager)

	// Add multiple clients and check IP assignment
	expectedIPs := []string{"10.0.0.2", "10.0.0.3", "10.0.0.4"}
	
	for i := 0; i < 3; i++ {
		key := make([]byte, 32)
		for j := range key {
			key[j] = byte(i*10 + j)
		}

		client, err := cm.AddClient(key, fmt.Sprintf("192.168.1.%d:12345", 100+i))
		if err != nil {
			t.Fatalf("AddClient failed: %v", err)
		}

		if client.IP != expectedIPs[i] {
			t.Errorf("Expected IP %s, got %s", expectedIPs[i], client.IP)
		}
	}
}

func TestClientManager_ConcurrentAccess(t *testing.T) {
	keyManager := crypto.NewKeyManager()
	cm := NewClientManager(keyManager)

	// Test concurrent client additions
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(i int) {
			key := make([]byte, 32)
			for j := range key {
				key[j] = byte(i*10 + j)
			}
			
			_, err := cm.AddClient(key, fmt.Sprintf("192.168.1.%d:12345", 100+i))
			if err != nil {
				t.Errorf("AddClient failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all additions
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check that all clients were added
	clients := cm.ListClients()
	if len(clients) != 10 {
		t.Errorf("Expected 10 clients, got %d", len(clients))
	}
}
