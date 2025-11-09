package protocol

import "fmt"

func ValidateMagic(packet *Packet) error {
	if string(packet.Magic[:]) != MagicBytes {
		return fmt.Errorf("invalid magic: got %s, want %s", string(packet.Magic[:]), MagicBytes)
	}
	return nil
}

func ValidateVersion(packet *Packet) error {
	major, _, _ := parseVersion(packet.Version)
	if major != ProtocolVersionMajor {
		return fmt.Errorf("unsupported version: got %d, want %d", major, ProtocolVersionMajor)
	}
	return nil
}

func ValidateType(packet *Packet) error {
	if packet.Type < PacketTypeData || packet.Type > PacketTypePong {
		return fmt.Errorf("invalid packet type: %d", packet.Type)
	}
	return nil
}

func ValidateLength(packet *Packet) error {
	if packet.Length != uint16(len(packet.Payload)) {
		return fmt.Errorf("length mismatch: header says %d, payload is %d", packet.Length, len(packet.Payload))
	}
	return nil
}

func ValidatePacket(packet *Packet) error {
	validators := []func(*Packet) error{
		ValidateMagic,
		ValidateVersion,
		ValidateType,
		ValidateLength,
	}

	for _, validate := range validators {
		if err := validate(packet); err != nil {
			return err
		}
	}
	return nil
}

func parseVersion(version byte) (major int, minor int, patch int) {

	major = 1
	minor = int(version >> 3)
	patch = int(version & 0x07)

	return major, minor, patch
}