package client

import (
	"testing"
)

// Test key: 64 hex characters (32 bytes)
const testKey = "0000000000000000000000000000000000000000000000000000000000000000"

func TestNewClient(t *testing.T) {
	client, err := NewClient("127.0.0.1:1194", 1, testKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	if client.serverAddr != "127.0.0.1:1194" {
		t.Errorf("Expected server address 127.0.0.1:1194, got %s", client.serverAddr)
	}
	
	if client.clientID != 1 {
		t.Errorf("Expected client ID 1, got %d", client.clientID)
	}
	
	if client.connected {
		t.Error("Expected client to be disconnected initially")
	}
	
	if client.sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", client.sequence)
	}
}

func TestClientMethods(t *testing.T) {
	client, err := NewClient("127.0.0.1:1194", 1, testKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Test initial state
	if client.IsConnected() {
		t.Error("Expected client to be disconnected initially")
	}
	
	if client.GetClientID() != 1 {
		t.Errorf("Expected client ID 1, got %d", client.GetClientID())
	}
	
	if client.GetAssignedIP() != "" {
		t.Errorf("Expected empty assigned IP, got %s", client.GetAssignedIP())
	}
}
