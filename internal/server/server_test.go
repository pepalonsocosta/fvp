package server

import (
	"net"
	"testing"
	"time"

	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/network"
	"github.com/pepalonsocosta/fvp/internal/protocol"
)

// TestNewServer tests server creation
func TestNewServer(t *testing.T) {
	server := NewServer()
	
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}
	
	if server.stopChan == nil {
		t.Error("Expected stopChan to be initialized")
	}
	
	if server.timeout != 30*time.Minute {
		t.Errorf("Expected default timeout to be 30 minutes, got %v", server.timeout)
	}
}

// TestLoadConfig tests configuration loading
func TestLoadConfig(t *testing.T) {
	server := NewServer()
	
	// Test with non-existent config file
	err := server.LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
	
	// Test with valid config file
	err = server.LoadConfig("../../server.example.yaml")
	if err != nil {
		t.Errorf("Expected no error for valid config file, got: %v", err)
	}
	
	if server.keyManager == nil {
		t.Error("Expected keyManager to be initialized")
	}
}

// TestCreateTUNInterface tests TUN interface creation
func TestCreateTUNInterface(t *testing.T) {
	server := NewServer()
	
	// Test with mock TUN interface (skip real TUN creation)
	// Just test that the function doesn't panic
	err := server.CreateTUNInterface()
	if err != nil {
		// Expected to fail due to permissions, that's OK for testing
		t.Logf("TUN creation failed as expected: %v", err)
	}
}

// TestCreateClientManager tests client manager creation
func TestCreateClientManager(t *testing.T) {
	server := NewServer()
	
	// Create key manager first
	server.keyManager = crypto.NewKeyManager()
	
	err := server.CreateClientManager()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if server.clientManager == nil {
		t.Error("Expected client manager to be created")
	}
}

// TestCreatePacketProcessor tests packet processor creation
func TestCreatePacketProcessor(t *testing.T) {
	server := NewServer()
	
	// Set up dependencies
	server.tunInterface = network.NewMockTunManager()
	server.keyManager = crypto.NewKeyManager()
	server.clientManager = NewClientManager(server.keyManager)
	
	// Create UDP server first
	err := server.CreateUDPServer(":0")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server.udpConn.Close()
	
	err = server.CreatePacketProcessor()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if server.packetProcessor == nil {
		t.Error("Expected packet processor to be created")
	}
}

// TestCreateUDPServer tests UDP server creation
func TestCreateUDPServer(t *testing.T) {
	server := NewServer()
	
	// Test with valid port
	err := server.CreateUDPServer(":0") // Use port 0 for testing
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if server.udpConn == nil {
		t.Error("Expected UDP connection to be created")
	}
	
	// Clean up
	server.udpConn.Close()
}

// TestSendAuthResponse tests auth response sending
func TestSendAuthResponse(t *testing.T) {
	server := NewServer()
	
	// Create UDP server
	err := server.CreateUDPServer(":0")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server.udpConn.Close()
	
	// Test sending auth response
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to resolve test address: %v", err)
	}
	
	err = server.sendAuthResponse(1, "10.0.0.2", []byte("test-key-32-bytes-long-key-here"), clientAddr)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestSendPongResponse tests pong response sending
func TestSendPongResponse(t *testing.T) {
	server := NewServer()
	
	// Set up dependencies
	server.keyManager = crypto.NewKeyManager()
	server.clientManager = NewClientManager(server.keyManager)
	
	// Create UDP server
	err := server.CreateUDPServer(":0")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server.udpConn.Close()
	
	// Add a test client
	key := make([]byte, 32)
	client, err := server.clientManager.AddClient(key, "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to add test client: %v", err)
	}
	
	// Test sending pong response
	err = server.sendPongResponse(client.ID, 123)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestProcessClientPacket tests client packet processing
func TestProcessClientPacket(t *testing.T) {
	server := NewServer()
	
	// Set up dependencies
	server.keyManager = crypto.NewKeyManager()
	
	// Load test keys
	key1 := make([]byte, 32)
	copy(key1, "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456")
	server.keyManager.SetTestKey(1, key1)
	
	server.clientManager = NewClientManager(server.keyManager)
	server.tunInterface = network.NewMockTunManager()
	server.packetProcessor = NewPacketProcessor(server.tunInterface, server.keyManager, server.clientManager, server.udpConn)
	
	// Create UDP server for sending responses
	err := server.CreateUDPServer(":0")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server.udpConn.Close()
	
	// Create test packet
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypeAuth,
		ClientID: 1,
		Sequence: 0,
		Length:   0,
		Version:  1,
		Payload:  []byte{},
	}
	
	// Encode packet
	packetData, err := protocol.EncodePacket(packet)
	if err != nil {
		t.Fatalf("Failed to encode test packet: %v", err)
	}
	
	// Create test address
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to resolve test address: %v", err)
	}
	
	// Test processing packet
	server.processClientPacket(packetData, clientAddr)
	
	// Verify client was added (auth packet should add client)
	client, err := server.clientManager.GetClient(1)
	if err != nil {
		t.Errorf("Expected client to be added, got error: %v", err)
	}
	
	if client == nil {
		t.Error("Expected client to exist after auth packet")
	}
}

// TestHandleAuthPacket tests auth packet handling
func TestHandleAuthPacket(t *testing.T) {
	server := NewServer()
	
	// Set up dependencies
	server.keyManager = crypto.NewKeyManager()
	
	// Load test keys
	key1 := make([]byte, 32)
	copy(key1, "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456")
	server.keyManager.SetTestKey(1, key1)
	
	server.clientManager = NewClientManager(server.keyManager)
	
	// Create UDP server
	err := server.CreateUDPServer(":0")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server.udpConn.Close()
	
	// Create test packet
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypeAuth,
		ClientID: 1,
		Sequence: 0,
		Length:   0,
		Version:  1,
		Payload:  []byte{},
	}
	
	// Create test address
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to resolve test address: %v", err)
	}
	
	// Test handling auth packet
	server.handleAuthPacket(packet, clientAddr)
	
	// Verify client was added
	client, err := server.clientManager.GetClient(1)
	if err != nil {
		t.Errorf("Expected client to be added, got error: %v", err)
	}
	
	if client == nil {
		t.Error("Expected client to exist after auth packet")
	}
}

// TestHandleDataPacket tests data packet handling
func TestHandleDataPacket(t *testing.T) {
	server := NewServer()
	
	// Set up dependencies
	server.keyManager = crypto.NewKeyManager()
	server.clientManager = NewClientManager(server.keyManager)
	server.tunInterface = network.NewMockTunManager()
	server.packetProcessor = NewPacketProcessor(server.tunInterface, server.keyManager, server.clientManager, server.udpConn)
	
	// Create test packet
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypeData,
		ClientID: 1,
		Sequence: 1,
		Length:   0,
		Version:  1,
		Payload:  []byte{},
	}
	
	// Create test address
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to resolve test address: %v", err)
	}
	
	// Test handling data packet
	server.handleDataPacket(packet, clientAddr)
	
	// Note: This test just verifies the function doesn't panic
	// The actual packet processing is tested in packet_processor_test.go
}

// TestHandlePingPacket tests ping packet handling
func TestHandlePingPacket(t *testing.T) {
	server := NewServer()
	
	// Set up dependencies
	server.keyManager = crypto.NewKeyManager()
	server.clientManager = NewClientManager(server.keyManager)
	
	// Create UDP server
	err := server.CreateUDPServer(":0")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server.udpConn.Close()
	
	// Add a test client first
	key := make([]byte, 32)
	client, err := server.clientManager.AddClient(key, "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to add test client: %v", err)
	}
	
	// Create test packet
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypePing,
		ClientID: client.ID,
		Sequence: 123,
		Length:   0,
		Version:  1,
		Payload:  []byte{},
	}
	
	// Create test address
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to resolve test address: %v", err)
	}
	
	// Test handling ping packet
	server.handlePingPacket(packet, clientAddr)
	
	// Verify client activity was updated
	updatedClient, err := server.clientManager.GetClient(client.ID)
	if err != nil {
		t.Errorf("Expected client to exist, got error: %v", err)
	}
	
	if updatedClient.LastSeq != 123 {
		t.Errorf("Expected LastSeq to be 123, got %d", updatedClient.LastSeq)
	}
}

// TestHandlePongPacket tests pong packet handling
func TestHandlePongPacket(t *testing.T) {
	server := NewServer()
	
	// Set up dependencies
	server.keyManager = crypto.NewKeyManager()
	server.clientManager = NewClientManager(server.keyManager)
	
	// Add a test client first
	key := make([]byte, 32)
	client, err := server.clientManager.AddClient(key, "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to add test client: %v", err)
	}
	
	// Create test packet
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypePong,
		ClientID: client.ID,
		Sequence: 456,
		Length:   0,
		Version:  1,
		Payload:  []byte{},
	}
	
	// Create test address
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		t.Fatalf("Failed to resolve test address: %v", err)
	}
	
	// Test handling pong packet
	server.handlePongPacket(packet, clientAddr)
	
	// Verify client activity was updated
	updatedClient, err := server.clientManager.GetClient(client.ID)
	if err != nil {
		t.Errorf("Expected client to exist, got error: %v", err)
	}
	
	if updatedClient.LastSeq != 456 {
		t.Errorf("Expected LastSeq to be 456, got %d", updatedClient.LastSeq)
	}
}

// TestProcessOutgoingPacket tests outgoing packet processing
func TestProcessOutgoingPacket(t *testing.T) {
	server := NewServer()
	
	// Set up dependencies
	mockTUN := network.NewMockTunManager()
	// Create the mock TUN interface first
	err := mockTUN.Create("test0")
	if err != nil {
		t.Fatalf("Failed to create mock TUN: %v", err)
	}
	
	server.tunInterface = mockTUN
	server.keyManager = crypto.NewKeyManager()
	server.clientManager = NewClientManager(server.keyManager)
	server.packetProcessor = NewPacketProcessor(server.tunInterface, server.keyManager, server.clientManager, server.udpConn)
	
	// Test processing outgoing packet
	server.processOutgoingPacket()
	
	// Note: This test just verifies the function doesn't panic
	// The actual packet processing is tested in packet_processor_test.go
}

// TestStop tests server shutdown
func TestStop(t *testing.T) {
	// Test stopping server without connections
	server1 := NewServer()
	err := server1.Stop()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	// Test stopping server with UDP connection
	server2 := NewServer()
	err = server2.CreateUDPServer(":0")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server2.udpConn.Close()
	
	err = server2.Stop()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
