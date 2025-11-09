package network

import (
	"errors"
	"sync"
)

// MockTunManager is a mock implementation for testing
type MockTunManager struct {
	name       string
	created    bool
	readQueue  [][]byte
	writeQueue [][]byte
	mu         sync.Mutex
}

// NewMockTunManager creates a new mock TUN manager
func NewMockTunManager() *MockTunManager {
	return &MockTunManager{
		readQueue:  make([][]byte, 0),
		writeQueue: make([][]byte, 0),
	}
}

// Create creates a mock TUN interface
func (mtm *MockTunManager) Create(name string) error {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()

	if mtm.created {
		return errors.New("interface already created")
	}

	mtm.name = name
	mtm.created = true
	return nil
}

// ReadPacket reads a packet from the mock interface
func (mtm *MockTunManager) ReadPacket() ([]byte, error) {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()

	if !mtm.created {
		return nil, errors.New("interface not created")
	}

	if len(mtm.readQueue) == 0 {
		return nil, errors.New("no packets available")
	}

	packet := mtm.readQueue[0]
	mtm.readQueue = mtm.readQueue[1:]
	return packet, nil
}

// WritePacket writes a packet to the mock interface
func (mtm *MockTunManager) WritePacket(data []byte) error {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()

	if !mtm.created {
		return errors.New("interface not created")
	}

	// Copy the data to avoid issues with slice references
	packet := make([]byte, len(data))
	copy(packet, data)
	mtm.writeQueue = append(mtm.writeQueue, packet)
	return nil
}

// Close closes the mock interface
func (mtm *MockTunManager) Close() error {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()

	mtm.created = false
	mtm.name = ""
	mtm.readQueue = nil
	mtm.writeQueue = nil
	return nil
}

// GetName returns the interface name
func (mtm *MockTunManager) GetName() string {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()
	return mtm.name
}

// IsCreated returns true if the interface is created
func (mtm *MockTunManager) IsCreated() bool {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()
	return mtm.created
}

// ConfigureClientInterface configures the mock TUN interface for a client
func (mtm *MockTunManager) ConfigureClientInterface(clientIP string) error {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()
	
	if !mtm.created {
		return errors.New("interface not created")
	}
	
	// In mock mode, we just log the configuration
	// In real mode, this would configure the TUN interface with the client IP
	return nil
}

// QueueReadPacket queues a packet for reading (testing helper)
func (mtm *MockTunManager) QueueReadPacket(data []byte) {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()
	
	packet := make([]byte, len(data))
	copy(packet, data)
	mtm.readQueue = append(mtm.readQueue, packet)
}

// GetWriteQueue returns the write queue (testing helper)
func (mtm *MockTunManager) GetWriteQueue() [][]byte {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()
	
	// Return a copy to avoid race conditions
	result := make([][]byte, len(mtm.writeQueue))
	for i, packet := range mtm.writeQueue {
		result[i] = make([]byte, len(packet))
		copy(result[i], packet)
	}
	return result
}

// ClearWriteQueue clears the write queue (testing helper)
func (mtm *MockTunManager) ClearWriteQueue() {
	mtm.mu.Lock()
	defer mtm.mu.Unlock()
	mtm.writeQueue = nil
}
