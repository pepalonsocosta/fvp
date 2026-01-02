package protocol

import (
	"encoding/binary"
	"errors"
)

func ParsePacket(data []byte) (*Packet, error) {
	if len(data) < HeaderSize {
		return nil, errors.New("packet too short")
	}

	return &Packet{
		Magic:    [3]byte{data[0], data[1], data[2]},
		Type:     data[3],
		ClientID: data[4],
		Sequence: binary.LittleEndian.Uint32(data[5:9]),
		Length:   binary.LittleEndian.Uint16(data[9:11]),
		Version:  data[11],
		Payload:  data[HeaderSize:],
	}, nil
}

func DecodePacket(data []byte) (*Packet, error) {
	packet, err := ParsePacket(data)
	if err != nil {
		return nil, err
	}

	if err := ValidatePacket(packet); err != nil {
		return nil, err
	}

	return packet, nil
}

func EncodePacket(packet *Packet) ([]byte, error) {
	data := make([]byte, HeaderSize+len(packet.Payload))

	copy(data[0:3], packet.Magic[:])
	data[3] = packet.Type
	data[4] = packet.ClientID
	binary.LittleEndian.PutUint32(data[5:9], packet.Sequence)
	binary.LittleEndian.PutUint16(data[9:11], packet.Length)
	data[11] = packet.Version
	copy(data[HeaderSize:], packet.Payload)

	return data, nil
}

