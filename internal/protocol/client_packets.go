package protocol

func CreateAuthPacket(clientID uint8, sequence uint32, payload []byte) *Packet {
	return &Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     PacketTypeAuth,
		ClientID: clientID,
		Sequence: sequence,
		Length:   uint16(len(payload)),
		Version:  ProtocolVersionByte,
		Payload:  payload,
	}
}

func CreateDataPacket(clientID uint8, sequence uint32, payload []byte) *Packet {
	return &Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     PacketTypeData,
		ClientID: clientID,
		Sequence: sequence,
		Length:   uint16(len(payload)),
		Version:  ProtocolVersionByte,
		Payload:  payload,
	}
}

func CreatePingPacket(clientID uint8, sequence uint32) *Packet {
	return &Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     PacketTypePing,
		ClientID: clientID,
		Sequence: sequence,
		Length:   0,
		Version:  ProtocolVersionByte,
		Payload:  []byte{},
	}
}
