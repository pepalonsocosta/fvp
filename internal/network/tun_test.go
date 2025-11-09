package network

import (
	"testing"
)

func TestMockTUNManager_Create(t *testing.T) {
	mtm := NewMockTunManager()

	// Test successful creation
	err := mtm.Create("fvp0")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if !mtm.IsCreated() {
		t.Error("Interface should be created")
	}

	if mtm.GetName() != "fvp0" {
		t.Errorf("Expected name 'fvp0', got '%s'", mtm.GetName())
	}

	// Test double creation
	err = mtm.Create("fvp1")
	if err == nil {
		t.Error("Expected error for double creation")
	}
}

func TestMockTUNManager_ReadPacket(t *testing.T) {
	mtm := NewMockTunManager()

	// Test read without creation
	_, err := mtm.ReadPacket()
	if err == nil {
		t.Error("Expected error when interface not created")
	}

	// Create interface
	err = mtm.Create("fvp0")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test read with empty queue
	_, err = mtm.ReadPacket()
	if err == nil {
		t.Error("Expected error when no packets available")
	}

	// Queue a packet
	testData := []byte("test packet data")
	mtm.QueueReadPacket(testData)

	// Test successful read
	data, err := mtm.ReadPacket()
	if err != nil {
		t.Errorf("ReadPacket failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(data))
	}
}

func TestMockTUNManager_WritePacket(t *testing.T) {
	mtm := NewMockTunManager()

	// Test write without creation
	err := mtm.WritePacket([]byte("test"))
	if err == nil {
		t.Error("Expected error when interface not created")
	}

	// Create interface
	err = mtm.Create("fvp0")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test successful write
	testData := []byte("test packet data")
	err = mtm.WritePacket(testData)
	if err != nil {
		t.Errorf("WritePacket failed: %v", err)
	}

	// Check write queue
	queue := mtm.GetWriteQueue()
	if len(queue) != 1 {
		t.Errorf("Expected 1 packet in queue, got %d", len(queue))
	}

	if string(queue[0]) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(queue[0]))
	}
}

func TestMockTUNManager_Close(t *testing.T) {
	mtm := NewMockTunManager()

	// Test close without creation
	err := mtm.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Create interface
	err = mtm.Create("fvp0")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test close after creation
	err = mtm.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if mtm.IsCreated() {
		t.Error("Interface should not be created after close")
	}

	if mtm.GetName() != "" {
		t.Error("Name should be empty after close")
	}
}

func TestMockTUNManager_QueueReadPacket(t *testing.T) {
	mtm := NewMockTunManager()
	mtm.Create("fvp0")

	// Queue multiple packets
	packet1 := []byte("packet 1")
	packet2 := []byte("packet 2")
	packet3 := []byte("packet 3")

	mtm.QueueReadPacket(packet1)
	mtm.QueueReadPacket(packet2)
	mtm.QueueReadPacket(packet3)

	// Read them back
	read1, err := mtm.ReadPacket()
	if err != nil {
		t.Errorf("ReadPacket 1 failed: %v", err)
	}
	if string(read1) != string(packet1) {
		t.Errorf("Expected %s, got %s", string(packet1), string(read1))
	}

	read2, err := mtm.ReadPacket()
	if err != nil {
		t.Errorf("ReadPacket 2 failed: %v", err)
	}
	if string(read2) != string(packet2) {
		t.Errorf("Expected %s, got %s", string(packet2), string(read2))
	}

	read3, err := mtm.ReadPacket()
	if err != nil {
		t.Errorf("ReadPacket 3 failed: %v", err)
	}
	if string(read3) != string(packet3) {
		t.Errorf("Expected %s, got %s", string(packet3), string(read3))
	}

	// Queue should be empty now
	_, err = mtm.ReadPacket()
	if err == nil {
		t.Error("Expected error when queue is empty")
	}
}

func TestMockTUNManager_GetWriteQueue(t *testing.T) {
	mtm := NewMockTunManager()
	mtm.Create("fvp0")

	// Write multiple packets
	packet1 := []byte("packet 1")
	packet2 := []byte("packet 2")

	mtm.WritePacket(packet1)
	mtm.WritePacket(packet2)

	// Get write queue
	queue := mtm.GetWriteQueue()
	if len(queue) != 2 {
		t.Errorf("Expected 2 packets in queue, got %d", len(queue))
	}

	if string(queue[0]) != string(packet1) {
		t.Errorf("Expected %s, got %s", string(packet1), string(queue[0]))
	}

	if string(queue[1]) != string(packet2) {
		t.Errorf("Expected %s, got %s", string(packet2), string(queue[1]))
	}

	// Clear queue
	mtm.ClearWriteQueue()
	queue = mtm.GetWriteQueue()
	if len(queue) != 0 {
		t.Errorf("Expected empty queue, got %d packets", len(queue))
	}
}

func TestMockTUNManager_ConcurrentAccess(t *testing.T) {
	mtm := NewMockTunManager()
	mtm.Create("fvp0")

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			packet := []byte{byte(i)}
			mtm.WritePacket(packet)
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check queue
	queue := mtm.GetWriteQueue()
	if len(queue) != 10 {
		t.Errorf("Expected 10 packets in queue, got %d", len(queue))
	}
}
