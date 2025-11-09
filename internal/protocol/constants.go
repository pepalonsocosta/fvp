package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	MagicBytes = "FVP"
	HeaderSize = 12

	PacketTypeData = 1
	PacketTypeAuth = 2
	PacketTypePing = 3
	PacketTypePong = 4
)

var (
	// Note: Packet encoding only supports major=1, so this will always be 1
	ProtocolVersionMajor = 1
	// ProtocolVersionMinor is set at runtime from injected app version (0-31)
	ProtocolVersionMinor = 0
	// ProtocolVersionPatch is set at runtime from injected app version (0-7)
	ProtocolVersionPatch = 0
	// ProtocolVersionByte is the encoded version byte for packets
	ProtocolVersionByte uint8 = 0
)

// Packet encoding limitation: major must be 1, minor 0-31, patch 0-7
func InitProtocolVersion(version string) error {
	if version == "" {
		// Default to 1.0.0 if no version provided
		ProtocolVersionMajor = 1
		ProtocolVersionMinor = 0
		ProtocolVersionPatch = 0
		ProtocolVersionByte = encodeVersion(1, 0, 0)
		return nil
	}

	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return fmt.Errorf("invalid version format: expected major.minor.patch, got %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid patch version: %w", err)
	}

	// Packet encoding only supports major=1
	if major != 1 {
		return fmt.Errorf("protocol version major must be 1 (got %d), protocol encoding limitation", major)
	}

	// Packet encoding limits: minor 0-31, patch 0-7
	if minor > 31 {
		return fmt.Errorf("protocol version minor must be 0-31 (got %d), protocol encoding limitation", minor)
	}

	if patch > 7 {
		return fmt.Errorf("protocol version patch must be 0-7 (got %d), protocol encoding limitation", patch)
	}

	ProtocolVersionMajor = major
	ProtocolVersionMinor = minor
	ProtocolVersionPatch = patch
	ProtocolVersionByte = encodeVersion(major, minor, patch)

	return nil
}

func encodeVersion(major, minor, patch int) uint8 {
	// major is always 1, not encoded in the byte
	// minor: 5 bits (0-31), shifted left by 3
	// patch: 3 bits (0-7), in lower 3 bits
	return uint8((minor << 3) | patch)
}

