package crypto

type ClientKeyManager struct {
	clientID uint8
	key      []byte
}

func NewClientKeyManager(clientID uint8, key []byte) *ClientKeyManager {
	// Create a copy of the key to prevent external modification
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	return &ClientKeyManager{
		clientID: clientID,
		key:      keyCopy,
	}
}

func (ckm *ClientKeyManager) GetKey() []byte {
	// Return a copy to prevent external modification
	keyCopy := make([]byte, len(ckm.key))
	copy(keyCopy, ckm.key)
	return keyCopy
}

func (ckm *ClientKeyManager) GetClientID() uint8 {
	return ckm.clientID
}
