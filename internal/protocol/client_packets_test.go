package protocol

import (
	"testing"
)

func init() {
	// Initialize protocol version for tests (default 1.0.0)
	if err := InitProtocolVersion("1.0.0"); err != nil {
		panic(err)
	}
}

func TestCreateAuthPacket(t *testing.T) {
	clientID := uint8(42)
	sequence := uint32(12345)
	payload := []byte("auth data")

	packet := CreateAuthPacket(clientID, sequence, payload)

	// Verify packet structure
	if string(packet.Magic[:]) != MagicBytes {
		t.Errorf("Expected magic %s, got %s", MagicBytes, string(packet.Magic[:]))
	}
	if packet.Type != PacketTypeAuth {
		t.Errorf("Expected type %d, got %d", PacketTypeAuth, packet.Type)
	}
	if packet.ClientID != clientID {
		t.Errorf("Expected client ID %d, got %d", clientID, packet.ClientID)
	}
	if packet.Sequence != sequence {
		t.Errorf("Expected sequence %d, got %d", sequence, packet.Sequence)
	}
	if packet.Length != uint16(len(payload)) {
		t.Errorf("Expected length %d, got %d", len(payload), packet.Length)
	}
	if packet.Version != ProtocolVersionByte {
		t.Errorf("Expected version %d, got %d", ProtocolVersionByte, packet.Version)
	}
	if string(packet.Payload) != string(payload) {
		t.Errorf("Expected payload %s, got %s", string(payload), string(packet.Payload))
	}
}

func TestCreateDataPacket(t *testing.T) {
	clientID := uint8(1)
	sequence := uint32(54321)
	payload := []byte("data payload")

	packet := CreateDataPacket(clientID, sequence, payload)

	// Verify packet structure
	if string(packet.Magic[:]) != MagicBytes {
		t.Errorf("Expected magic %s, got %s", MagicBytes, string(packet.Magic[:]))
	}
	if packet.Type != PacketTypeData {
		t.Errorf("Expected type %d, got %d", PacketTypeData, packet.Type)
	}
	if packet.ClientID != clientID {
		t.Errorf("Expected client ID %d, got %d", clientID, packet.ClientID)
	}
	if packet.Sequence != sequence {
		t.Errorf("Expected sequence %d, got %d", sequence, packet.Sequence)
	}
	if packet.Length != uint16(len(payload)) {
		t.Errorf("Expected length %d, got %d", len(payload), packet.Length)
	}
	if packet.Version != ProtocolVersionByte {
		t.Errorf("Expected version %d, got %d", ProtocolVersionByte, packet.Version)
	}
	if string(packet.Payload) != string(payload) {
		t.Errorf("Expected payload %s, got %s", string(payload), string(packet.Payload))
	}
}

func TestCreatePingPacket(t *testing.T) {
	clientID := uint8(255)
	sequence := uint32(0xFFFFFFFF)

	packet := CreatePingPacket(clientID, sequence)

	// Verify packet structure
	if string(packet.Magic[:]) != MagicBytes {
		t.Errorf("Expected magic %s, got %s", MagicBytes, string(packet.Magic[:]))
	}
	if packet.Type != PacketTypePing {
		t.Errorf("Expected type %d, got %d", PacketTypePing, packet.Type)
	}
	if packet.ClientID != clientID {
		t.Errorf("Expected client ID %d, got %d", clientID, packet.ClientID)
	}
	if packet.Sequence != sequence {
		t.Errorf("Expected sequence %d, got %d", sequence, packet.Sequence)
	}
	if packet.Length != 0 {
		t.Errorf("Expected length 0, got %d", packet.Length)
	}
	if packet.Version != ProtocolVersionByte {
		t.Errorf("Expected version %d, got %d", ProtocolVersionByte, packet.Version)
	}
	if len(packet.Payload) != 0 {
		t.Errorf("Expected empty payload, got %d bytes", len(packet.Payload))
	}
}

func TestCreatePacketsEdgeCases(t *testing.T) {
	// Test with empty payload
	packet := CreateAuthPacket(1, 0, []byte{})
	if packet.Length != 0 {
		t.Errorf("Expected length 0 for empty payload, got %d", packet.Length)
	}
	if len(packet.Payload) != 0 {
		t.Errorf("Expected empty payload, got %d bytes", len(packet.Payload))
	}

	// Test with maximum client ID
	packet = CreateDataPacket(255, 0, []byte("test"))
	if packet.ClientID != 255 {
		t.Errorf("Expected client ID 255, got %d", packet.ClientID)
	}

	// Test with maximum sequence
	packet = CreatePingPacket(1, 0xFFFFFFFF)
	if packet.Sequence != 0xFFFFFFFF {
		t.Errorf("Expected sequence 0xFFFFFFFF, got 0x%X", packet.Sequence)
	}
}
