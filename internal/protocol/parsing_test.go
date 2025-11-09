package protocol

import (
	"testing"
)

func TestParsePacket(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		expected    *Packet
	}{
		{
			name:        "valid data packet",
			data:        []byte{'F', 'V', 'P', PacketTypeData, 1, 0, 0, 0, 0, 5, 0, 1, 'h', 'e', 'l', 'l', 'o'},
			expectError: false,
			expected: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeData,
				ClientID: 1,
				Sequence: 0,
				Length:   5,
				Version:  1,
				Payload:  []byte{'h', 'e', 'l', 'l', 'o'},
			},
		},
		{
			name:        "valid auth packet",
			data:        []byte{'F', 'V', 'P', PacketTypeAuth, 42, 0x12, 0x34, 0x56, 0x78, 3, 0, 1, 'a', 'b', 'c'},
			expectError: false,
			expected: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeAuth,
				ClientID: 42,
				Sequence: 0x78563412,
				Length:   3,
				Version:  1,
				Payload:  []byte{'a', 'b', 'c'},
			},
		},
		{
			name:        "valid ping packet with empty payload",
			data:        []byte{'F', 'V', 'P', PacketTypePing, 0, 0, 0, 0, 0, 0, 0, 1},
			expectError: false,
			expected: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypePing,
				ClientID: 0,
				Sequence: 0,
				Length:   0,
				Version:  1,
				Payload:  []byte{},
			},
		},
		{
			name:        "packet too short",
			data:        []byte{'F', 'V', 'P', PacketTypeData},
			expectError: true,
			expected:    nil,
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
			expected:    nil,
		},
		{
			name:        "nil data",
			data:        nil,
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePacket(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("expected packet but got nil")
				return
			}

			// Compare all fields
			if result.Magic != tt.expected.Magic {
				t.Errorf("Magic mismatch: got %v, want %v", result.Magic, tt.expected.Magic)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Type mismatch: got %d, want %d", result.Type, tt.expected.Type)
			}
			if result.ClientID != tt.expected.ClientID {
				t.Errorf("ClientID mismatch: got %d, want %d", result.ClientID, tt.expected.ClientID)
			}
			if result.Sequence != tt.expected.Sequence {
				t.Errorf("Sequence mismatch: got %d, want %d", result.Sequence, tt.expected.Sequence)
			}
			if result.Length != tt.expected.Length {
				t.Errorf("Length mismatch: got %d, want %d", result.Length, tt.expected.Length)
			}
			if result.Version != tt.expected.Version {
				t.Errorf("Version mismatch: got %d, want %d", result.Version, tt.expected.Version)
			}
			if len(result.Payload) != len(tt.expected.Payload) {
				t.Errorf("Payload length mismatch: got %d, want %d", len(result.Payload), len(tt.expected.Payload))
			} else {
				for i, b := range result.Payload {
					if b != tt.expected.Payload[i] {
						t.Errorf("Payload byte %d mismatch: got %d, want %d", i, b, tt.expected.Payload[i])
					}
				}
			}
		})
	}
}

func TestDecodePacket(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "valid packet",
			data:        []byte{'F', 'V', 'P', PacketTypeData, 1, 0, 0, 0, 0, 5, 0, 1, 'h', 'e', 'l', 'l', 'o'},
			expectError: false,
		},
		{
			name:        "invalid magic",
			data:        []byte{'X', 'Y', 'Z', PacketTypeData, 1, 0, 0, 0, 0, 5, 0, 1, 'h', 'e', 'l', 'l', 'o'},
			expectError: true,
		},
		{
			name:        "invalid packet type",
			data:        []byte{'F', 'V', 'P', 99, 1, 0, 0, 0, 0, 5, 0, 1, 'h', 'e', 'l', 'l', 'o'},
			expectError: true,
		},
		{
			name:        "length mismatch",
			data:        []byte{'F', 'V', 'P', PacketTypeData, 1, 0, 0, 0, 0, 10, 0, 1, 'h', 'e', 'l', 'l', 'o'},
			expectError: true,
		},
		{
			name:        "packet too short",
			data:        []byte{'F', 'V', 'P'},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodePacket(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("expected packet but got nil")
			}
		})
	}
}

func TestEncodePacket(t *testing.T) {
	tests := []struct {
		name     string
		packet   *Packet
		expected []byte
	}{
		{
			name: "data packet",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeData,
				ClientID: 1,
				Sequence: 0,
				Length:   5,
				Version:  1,
				Payload:  []byte{'h', 'e', 'l', 'l', 'o'},
			},
			expected: []byte{'F', 'V', 'P', PacketTypeData, 1, 0, 0, 0, 0, 5, 0, 1, 'h', 'e', 'l', 'l', 'o'},
		},
		{
			name: "auth packet with sequence",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeAuth,
				ClientID: 42,
				Sequence: 0x78563412,
				Length:   3,
				Version:  1,
				Payload:  []byte{'a', 'b', 'c'},
			},
			expected: []byte{'F', 'V', 'P', PacketTypeAuth, 42, 0x12, 0x34, 0x56, 0x78, 3, 0, 1, 'a', 'b', 'c'},
		},
		{
			name: "ping packet with empty payload",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypePing,
				ClientID: 0,
				Sequence: 0,
				Length:   0,
				Version:  1,
				Payload:  []byte{},
			},
			expected: []byte{'F', 'V', 'P', PacketTypePing, 0, 0, 0, 0, 0, 0, 0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodePacket(tt.packet)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("length mismatch: got %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, b := range result {
				if b != tt.expected[i] {
					t.Errorf("byte %d mismatch: got %d, want %d", i, b, tt.expected[i])
				}
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := &Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     PacketTypeData,
		ClientID: 123,
		Sequence: 0x12345678,
		Length:   7,
		Version:  1,
		Payload:  []byte{'t', 'e', 's', 't', 'i', 'n', 'g'},
	}

	// Encode
	encoded, err := EncodePacket(original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Decode
	decoded, err := DecodePacket(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// Compare
	if decoded.Magic != original.Magic {
		t.Errorf("Magic mismatch: got %v, want %v", decoded.Magic, original.Magic)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type mismatch: got %d, want %d", decoded.Type, original.Type)
	}
	if decoded.ClientID != original.ClientID {
		t.Errorf("ClientID mismatch: got %d, want %d", decoded.ClientID, original.ClientID)
	}
	if decoded.Sequence != original.Sequence {
		t.Errorf("Sequence mismatch: got %d, want %d", decoded.Sequence, original.Sequence)
	}
	if decoded.Length != original.Length {
		t.Errorf("Length mismatch: got %d, want %d", decoded.Length, original.Length)
	}
	if decoded.Version != original.Version {
		t.Errorf("Version mismatch: got %d, want %d", decoded.Version, original.Version)
	}
	if len(decoded.Payload) != len(original.Payload) {
		t.Errorf("Payload length mismatch: got %d, want %d", len(decoded.Payload), len(original.Payload))
	} else {
		for i, b := range decoded.Payload {
			if b != original.Payload[i] {
				t.Errorf("Payload byte %d mismatch: got %d, want %d", i, b, original.Payload[i])
			}
		}
	}
}

func TestParsePacketEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "exactly header size",
			data:        []byte{'F', 'V', 'P', PacketTypePing, 0, 0, 0, 0, 0, 0, 0, 1},
			expectError: false,
		},
		{
			name:        "one byte short of header",
			data:        []byte{'F', 'V', 'P', PacketTypePing, 0, 0, 0, 0, 0, 0, 0},
			expectError: true,
		},
		{
			name:        "maximum client ID",
			data:        []byte{'F', 'V', 'P', PacketTypeData, 255, 0, 0, 0, 0, 0, 0, 1},
			expectError: false,
		},
		{
			name:        "maximum sequence number",
			data:        []byte{'F', 'V', 'P', PacketTypeData, 0, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 1},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePacket(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
} 