package server

import (
	"net"
	"testing"

	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/network"
	"github.com/pepalonsocosta/fvp/internal/protocol"
)

func TestPacketProcessor_ProcessPacket(t *testing.T) {
	// Create mock TUN interface
	mockTUN := network.NewMockTunManager()
	
	// Create the mock TUN interface
	err := mockTUN.Create("test0")
	if err != nil {
		t.Fatalf("Failed to create mock TUN: %v", err)
	}
	
	// Create key manager
	keyManager := crypto.NewKeyManager()
	
	// Create client manager
	clientManager := NewClientManager(keyManager)
	
	// Create mock UDP connection
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	mockUDPConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer mockUDPConn.Close()
	
	// Create packet processor
	processor := NewPacketProcessor(mockTUN, keyManager, clientManager, mockUDPConn)
	
	// Add a client
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	
	client, err := clientManager.AddClient(key, "192.168.1.100:12345")
	if err != nil {
		t.Fatalf("Failed to add client: %v", err)
	}
	
	// Create a test packet with encrypted payload
	testPayload := []byte("Hello, World!")
	
	// Encrypt only the payload
	encryptedPayload, err := crypto.EncryptPayload(testPayload, client.Key, 1)
	if err != nil {
		t.Fatalf("Failed to encrypt payload: %v", err)
	}
	
	// Create the final packet with encrypted payload
	finalPacket := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypeData,
		ClientID: client.ID,
		Sequence: 1,
		Length:   uint16(len(encryptedPayload)),
		Version:  1,
		Payload:  encryptedPayload,
	}
	
	// Encode the final packet
	finalPacketData, err := protocol.EncodePacket(finalPacket)
	if err != nil {
		t.Fatalf("Failed to encode final packet: %v", err)
	}
	
	// Process the packet
	err = processor.ProcessPacket(finalPacketData)
	if err != nil {
		t.Fatalf("ProcessPacket failed: %v", err)
	}
	
	// Check that packet was written to TUN
	writeQueue := mockTUN.GetWriteQueue()
	if len(writeQueue) != 1 {
		t.Errorf("Expected 1 packet in TUN write queue, got %d", len(writeQueue))
	}
	
	// Check that the decrypted payload matches
	if string(writeQueue[0]) != string(testPayload) {
		t.Errorf("Expected payload %s, got %s", string(testPayload), string(writeQueue[0]))
	}
}

func TestPacketProcessor_ProcessPacket_InvalidPacket(t *testing.T) {
	// Create mock TUN interface
	mockTUN := network.NewMockTunManager()
	
	// Create key manager
	keyManager := crypto.NewKeyManager()
	
	// Create client manager
	clientManager := NewClientManager(keyManager)
	
	// Create mock UDP connection
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	mockUDPConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer mockUDPConn.Close()
	
	// Create packet processor
	processor := NewPacketProcessor(mockTUN, keyManager, clientManager, mockUDPConn)
	
	// Test with invalid packet
	err = processor.ProcessPacket([]byte("invalid packet"))
	if err == nil {
		t.Error("Expected error for invalid packet")
	}
}

func TestPacketProcessor_ProcessPacket_UnknownClient(t *testing.T) {
	// Create mock TUN interface
	mockTUN := network.NewMockTunManager()
	
	// Create key manager
	keyManager := crypto.NewKeyManager()
	
	// Create client manager
	clientManager := NewClientManager(keyManager)
	
	// Create mock UDP connection
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	mockUDPConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer mockUDPConn.Close()
	
	// Create packet processor
	processor := NewPacketProcessor(mockTUN, keyManager, clientManager, mockUDPConn)
	
	// Create a test packet with unknown client ID
	testPayload := []byte("Hello, World!")
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypeData,
		ClientID: 99, // Unknown client
		Sequence: 1,
		Length:   uint16(len(testPayload)),
		Version:  1,
		Payload:  testPayload,
	}
	
	// Encode packet
	packetData, err := protocol.EncodePacket(packet)
	if err != nil {
		t.Fatalf("Failed to encode packet: %v", err)
	}
	
	// Process the packet
	err = processor.ProcessPacket(packetData)
	if err == nil {
		t.Error("Expected error for unknown client")
	}
}

func TestPacketProcessor_ProcessOutgoingPacket(t *testing.T) {
	// Create mock TUN interface
	mockTUN := network.NewMockTunManager()
	
	// Create the mock TUN interface
	err := mockTUN.Create("test0")
	if err != nil {
		t.Fatalf("Failed to create mock TUN: %v", err)
	}
	
	// Create key manager
	keyManager := crypto.NewKeyManager()
	
	// Create client manager
	clientManager := NewClientManager(keyManager)
	
	// Create mock UDP connection
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	mockUDPConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer mockUDPConn.Close()
	
	// Create packet processor
	processor := NewPacketProcessor(mockTUN, keyManager, clientManager, mockUDPConn)
	
	// Add a client
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	
	_, err = clientManager.AddClient(key, "192.168.1.100:12345")
	if err != nil {
		t.Fatalf("Failed to add client: %v", err)
	}
	
	// Create a mock IP packet (destination IP = client's IP)
	ipPacket := createMockIPPacket("10.0.0.2", "8.8.8.8", []byte("test data"))
	
	// Queue the packet in TUN
	mockTUN.QueueReadPacket(ipPacket)
	
	// Process outgoing packet
	err = processor.ProcessOutgoingPacket()
	if err != nil {
		t.Fatalf("ProcessOutgoingPacket failed: %v", err)
	}
}

func TestPacketProcessor_ProcessOutgoingPacket_NoPackets(t *testing.T) {
	// Create mock TUN interface
	mockTUN := network.NewMockTunManager()
	
	// Create the mock TUN interface
	err := mockTUN.Create("test0")
	if err != nil {
		t.Fatalf("Failed to create mock TUN: %v", err)
	}
	
	// Create key manager
	keyManager := crypto.NewKeyManager()
	
	// Create client manager
	clientManager := NewClientManager(keyManager)
	
	// Create mock UDP connection
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	mockUDPConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer mockUDPConn.Close()
	
	// Create packet processor
	processor := NewPacketProcessor(mockTUN, keyManager, clientManager, mockUDPConn)
	
	// Process outgoing packet with no packets in TUN
	err = processor.ProcessOutgoingPacket()
	if err == nil {
		t.Error("Expected error when no packets available")
	}
}

// createMockIPPacket creates a mock IP packet for testing
func createMockIPPacket(srcIP, dstIP string, payload []byte) []byte {
	// Simple IP header (20 bytes) + payload
	packet := make([]byte, 20+len(payload))
	
	// IP version 4, header length 5
	packet[0] = 0x45
	
	// Total length
	length := uint16(20 + len(payload))
	packet[2] = byte(length >> 8)
	packet[3] = byte(length & 0xFF)
	
	// Protocol (UDP = 17)
	packet[9] = 17
	
	// Source IP (simplified - just use last octet)
	packet[12] = 192
	packet[13] = 168
	packet[14] = 1
	packet[15] = 100
	
	// Destination IP (simplified - just use last octet)
	packet[16] = 10
	packet[17] = 0
	packet[18] = 0
	packet[19] = 2
	
	// Copy payload
	copy(packet[20:], payload)
	
	return packet
}
