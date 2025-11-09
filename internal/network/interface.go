package network

type TUNInterface interface {
	Create(name string) error
	ReadPacket() ([]byte, error)
	WritePacket(data []byte) error
	Close() error
	GetName() string
	IsCreated() bool
	ConfigureClientInterface(clientIP string) error
}

// Ensure both implementations satisfy the interface
var _ TUNInterface = (*TunManager)(nil)
var _ TUNInterface = (*MockTunManager)(nil)
