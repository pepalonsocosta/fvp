package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("127.0.0.1:1194")
	
	if client.serverAddr != "127.0.0.1:1194" {
		t.Errorf("Expected server address 127.0.0.1:1194, got %s", client.serverAddr)
	}
	
	if client.clientID != 0 {
		t.Errorf("Expected client ID 0, got %d", client.clientID)
	}
	
	if client.connected {
		t.Error("Expected client to be disconnected initially")
	}
	
	if client.sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", client.sequence)
	}
}

func TestClientMethods(t *testing.T) {
	client := NewClient("127.0.0.1:1194")
	
	// Test initial state
	if client.IsConnected() {
		t.Error("Expected client to be disconnected initially")
	}
	
	if client.GetClientID() != 0 {
		t.Errorf("Expected client ID 0, got %d", client.GetClientID())
	}
	
	if client.GetAssignedIP() != "" {
		t.Errorf("Expected empty assigned IP, got %s", client.GetAssignedIP())
	}
}
