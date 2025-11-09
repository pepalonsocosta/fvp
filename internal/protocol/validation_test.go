package protocol

import (
	"testing"
)

func TestValidateMagic(t *testing.T) {
	tests := []struct {
		name        string
		packet      *Packet
		expectError bool
	}{
		{
			name: "valid magic",
			packet: &Packet{
				Magic: [3]byte{'F', 'V', 'P'},
			},
			expectError: false,
		},
		{
			name: "invalid magic - wrong first byte",
			packet: &Packet{
				Magic: [3]byte{'X', 'V', 'P'},
			},
			expectError: true,
		},
		{
			name: "invalid magic - wrong second byte",
			packet: &Packet{
				Magic: [3]byte{'F', 'X', 'P'},
			},
			expectError: true,
		},
		{
			name: "invalid magic - wrong third byte",
			packet: &Packet{
				Magic: [3]byte{'F', 'V', 'X'},
			},
			expectError: true,
		},
		{
			name: "invalid magic - all wrong",
			packet: &Packet{
				Magic: [3]byte{'X', 'Y', 'Z'},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMagic(tt.packet)

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

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name        string
		packet      *Packet
		expectError bool
	}{
		{
			name: "valid version - major 1, minor 0, patch 0",
			packet: &Packet{
				Version: 0x00, // 0 << 3 | 0 = 0
			},
			expectError: false,
		},
		{
			name: "valid version - major 1, minor 0, patch 1",
			packet: &Packet{
				Version: 0x01, // 0 << 3 | 1 = 1
			},
			expectError: false,
		},
		{
			name: "valid version - major 1, minor 1, patch 0",
			packet: &Packet{
				Version: 0x08, // 1 << 3 | 0 = 8
			},
			expectError: false,
		},
		{
			name: "valid version - major 1, minor 1, patch 1",
			packet: &Packet{
				Version: 0x09, // 1 << 3 | 1 = 9
			},
			expectError: false,
		},
		{
			name: "valid version - major 1, minor 7, patch 7",
			packet: &Packet{
				Version: 0x3F, // 7 << 3 | 7 = 63
			},
			expectError: false,
		},
		{
			name: "valid version - major 1, minor 31, patch 7",
			packet: &Packet{
				Version: 0xFF, // 31 << 3 | 7 = 255
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.packet)

			if tt.expectError {
				if err != nil {
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

func TestValidateType(t *testing.T) {
	tests := []struct {
		name        string
		packet      *Packet
		expectError bool
	}{
		{
			name: "valid type - Data",
			packet: &Packet{
				Type: PacketTypeData,
			},
			expectError: false,
		},
		{
			name: "valid type - Auth",
			packet: &Packet{
				Type: PacketTypeAuth,
			},
			expectError: false,
		},
		{
			name: "valid type - Ping",
			packet: &Packet{
				Type: PacketTypePing,
			},
			expectError: false,
		},
		{
			name: "valid type - Pong",
			packet: &Packet{
				Type: PacketTypePong,
			},
			expectError: false,
		},
		{
			name: "invalid type - too low",
			packet: &Packet{
				Type: 0,
			},
			expectError: true,
		},
		{
			name: "invalid type - too high",
			packet: &Packet{
				Type: 5,
			},
			expectError: true,
		},
		{
			name: "invalid type - very high",
			packet: &Packet{
				Type: 255,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateType(tt.packet)

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

func TestValidateLength(t *testing.T) {
	tests := []struct {
		name        string
		packet      *Packet
		expectError bool
	}{
		{
			name: "valid length - empty payload",
			packet: &Packet{
				Length:  0,
				Payload: []byte{},
			},
			expectError: false,
		},
		{
			name: "valid length - small payload",
			packet: &Packet{
				Length:  3,
				Payload: []byte{'a', 'b', 'c'},
			},
			expectError: false,
		},
		{
			name: "valid length - large payload",
			packet: &Packet{
				Length:  1000,
				Payload: make([]byte, 1000),
			},
			expectError: false,
		},
		{
			name: "invalid length - too short",
			packet: &Packet{
				Length:  5,
				Payload: []byte{'a', 'b', 'c'},
			},
			expectError: true,
		},
		{
			name: "invalid length - too long",
			packet: &Packet{
				Length:  3,
				Payload: []byte{'a', 'b', 'c', 'd', 'e'},
			},
			expectError: true,
		},
		{
			name: "invalid length - zero length with payload",
			packet: &Packet{
				Length:  0,
				Payload: []byte{'a'},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLength(tt.packet)

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

func TestValidatePacket(t *testing.T) {
	tests := []struct {
		name        string
		packet      *Packet
		expectError bool
	}{
		{
			name: "valid packet",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeData,
				ClientID: 1,
				Sequence: 0,
				Length:   5,
				Version:  0x00, // major 1, minor 0, patch 0
				Payload:  []byte{'h', 'e', 'l', 'l', 'o'},
			},
			expectError: false,
		},
		{
			name: "invalid packet - wrong magic",
			packet: &Packet{
				Magic:    [3]byte{'X', 'Y', 'Z'},
				Type:     PacketTypeData,
				ClientID: 1,
				Sequence: 0,
				Length:   5,
				Version:  0x00,
				Payload:  []byte{'h', 'e', 'l', 'l', 'o'},
			},
			expectError: true,
		},
		{
			name: "invalid packet - wrong type",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     99,
				ClientID: 1,
				Sequence: 0,
				Length:   5,
				Version:  0x00,
				Payload:  []byte{'h', 'e', 'l', 'l', 'o'},
			},
			expectError: true,
		},
		{
			name: "invalid packet - length mismatch",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeData,
				ClientID: 1,
				Sequence: 0,
				Length:   10,
				Version:  0x00,
				Payload:  []byte{'h', 'e', 'l', 'l', 'o'},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePacket(tt.packet)

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

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name           string
		version        byte
		expectedMajor  int
		expectedMinor  int
		expectedPatch  int
	}{
		{
			name:          "version 1.0.0",
			version:       0x00, // 0 << 3 | 0 = 0
			expectedMajor: 1,
			expectedMinor: 0,
			expectedPatch: 0,
		},
		{
			name:          "version 1.0.1",
			version:       0x01, // 0 << 3 | 1 = 1
			expectedMajor: 1,
			expectedMinor: 0,
			expectedPatch: 1,
		},
		{
			name:          "version 1.1.0",
			version:       0x08, // 1 << 3 | 0 = 8
			expectedMajor: 1,
			expectedMinor: 1,
			expectedPatch: 0,
		},
		{
			name:          "version 1.1.1",
			version:       0x09, // 1 << 3 | 1 = 9
			expectedMajor: 1,
			expectedMinor: 1,
			expectedPatch: 1,
		},
		{
			name:          "version 1.7.7",
			version:       0x3F, // 7 << 3 | 7 = 63
			expectedMajor: 1,
			expectedMinor: 7,
			expectedPatch: 7,
		},
		{
			name:          "version 1.31.7",
			version:       0xFF, // 31 << 3 | 7 = 255
			expectedMajor: 1,
			expectedMinor: 31,
			expectedPatch: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch := parseVersion(tt.version)

			if major != tt.expectedMajor {
				t.Errorf("major version mismatch: got %d, want %d", major, tt.expectedMajor)
			}
			if minor != tt.expectedMinor {
				t.Errorf("minor version mismatch: got %d, want %d", minor, tt.expectedMinor)
			}
			if patch != tt.expectedPatch {
				t.Errorf("patch version mismatch: got %d, want %d", patch, tt.expectedPatch)
			}
		})
	}
}

func TestValidatePacketEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		packet      *Packet
		expectError bool
	}{
		{
			name: "maximum client ID",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeData,
				ClientID: 255,
				Sequence: 0,
				Length:   0,
				Version:  0x00, // major 1, minor 0, patch 0
				Payload:  []byte{},
			},
			expectError: false,
		},
		{
			name: "maximum sequence number",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeData,
				ClientID: 0,
				Sequence: 0xFFFFFFFF,
				Length:   0,
				Version:  0x00, // major 1, minor 0, patch 0
				Payload:  []byte{},
			},
			expectError: false,
		},
		{
			name: "maximum payload length",
			packet: &Packet{
				Magic:    [3]byte{'F', 'V', 'P'},
				Type:     PacketTypeData,
				ClientID: 0,
				Sequence: 0,
				Length:   0xFFFF,
				Version:  0x00, // major 1, minor 0, patch 0
				Payload:  make([]byte, 0xFFFF),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePacket(tt.packet)

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