package protocol

type Packet struct {
	Magic [3]byte // "FVP"
	Type  uint8   // 1-4
	ClientID uint8 // 0-255
	Sequence uint32 // Sequence number
	Length uint16 // Payload length
	Version uint8 // Protocol version
	Payload []byte
}
