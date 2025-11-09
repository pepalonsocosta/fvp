package server

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pepalonsocosta/fvp/internal/crypto"
)

type Client struct {
	ID       uint8
	IP       string
	Key      []byte
	Address  string
	Connected bool
	LastSeen  time.Time
	LastSeq   uint32
}

type ClientManager struct {
	clients     map[uint8]*Client
	ipToClient  map[string]uint8
	keyToClient map[string]uint8
	mutex       sync.RWMutex
	timeout     time.Duration
	keyManager  *crypto.KeyManager
}

var (
	ErrClientNotFound      = errors.New("client not found")
	ErrClientAlreadyExists = errors.New("client already exists")
	ErrMaxClientsReached   = errors.New("maximum clients reached (256)")
	ErrInvalidKey          = errors.New("invalid client key")
	ErrClientTimeout       = errors.New("client timeout")
	ErrInvalidSequence     = errors.New("invalid sequence number")
	ErrClientDisconnected  = errors.New("client disconnected")
)

func NewClientManager(keyManager *crypto.KeyManager) *ClientManager {
	cm := &ClientManager{
		clients:     make(map[uint8]*Client),
		ipToClient:  make(map[string]uint8),
		keyToClient: make(map[string]uint8),
		timeout:     30 * time.Minute,
		keyManager:  keyManager,
	}
	
	go cm.startTimeoutChecker()
	
	return cm
}

func (cm *ClientManager) AddClient(key []byte, address string) (*Client, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if len(cm.clients) >= 256 {
		return nil, ErrMaxClientsReached
	}
	
	keyHash := fmt.Sprintf("%x", key)
	if _, exists := cm.keyToClient[keyHash]; exists {
		return nil, ErrClientAlreadyExists
	}
	
	clientID := cm.findNextClientID()
	if clientID == 0 {
		return nil, ErrMaxClientsReached
	}
	
	ip := cm.assignNextIP()
	if ip == "" {
		return nil, fmt.Errorf("no IP addresses available")
	}
	
	client := &Client{
		ID:        clientID,
		IP:        ip,
		Key:       key,
		Address:   address,
		Connected: true,
		LastSeen:  time.Now(),
		LastSeq:   0,
	}
	
	cm.clients[clientID] = client
	cm.ipToClient[ip] = clientID
	cm.keyToClient[keyHash] = clientID
	
	log.Printf("Added client %d with IP %s from %s", clientID, ip, address)
	return client, nil
}

func (cm *ClientManager) RemoveClient(clientID uint8) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	client, exists := cm.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}
	
	delete(cm.clients, clientID)
	delete(cm.ipToClient, client.IP)
	keyHash := fmt.Sprintf("%x", client.Key)
	delete(cm.keyToClient, keyHash)
	
	log.Printf("Removed client %d with IP %s", clientID, client.IP)
	return nil
}

func (cm *ClientManager) GetClient(clientID uint8) (*Client, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	client, exists := cm.clients[clientID]
	if !exists {
		return nil, ErrClientNotFound
	}
	
	return client, nil
}

func (cm *ClientManager) GetClientByIP(ip string) (*Client, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	clientID, exists := cm.ipToClient[ip]
	if !exists {
		return nil, ErrClientNotFound
	}
	
	client, exists := cm.clients[clientID]
	if !exists {
		return nil, ErrClientNotFound
	}
	
	return client, nil
}

func (cm *ClientManager) ListClients() []*Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	clients := make([]*Client, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	
	return clients
}

func (cm *ClientManager) UpdateClientActivity(clientID uint8, sequence uint32) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	client, exists := cm.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}
	
	if sequence <= client.LastSeq {
		return ErrInvalidSequence
	}
	
	client.LastSeen = time.Now()
	client.LastSeq = sequence
	
	return nil
}

func (cm *ClientManager) CheckTimeouts() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	now := time.Now()
	var toRemove []uint8
	
	for clientID, client := range cm.clients {
		if now.Sub(client.LastSeen) > cm.timeout {
			toRemove = append(toRemove, clientID)
		}
	}
	
	for _, clientID := range toRemove {
		client := cm.clients[clientID]
		delete(cm.clients, clientID)
		delete(cm.ipToClient, client.IP)
		keyHash := fmt.Sprintf("%x", client.Key)
		delete(cm.keyToClient, keyHash)
		log.Printf("Removed timed-out client %d with IP %s", clientID, client.IP)
	}
}

func (cm *ClientManager) startTimeoutChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		cm.CheckTimeouts()
	}
}

func (cm *ClientManager) findNextClientID() uint8 {
	for i := uint8(1); i != 0; i++ {
		if _, exists := cm.clients[i]; !exists {
			return i
		}
	}
	return 0
}

func (cm *ClientManager) assignNextIP() string {
	for i := 2; i <= 255; i++ {
		ip := fmt.Sprintf("10.0.0.%d", i)
		if _, exists := cm.ipToClient[ip]; !exists {
			return ip
		}
	}
	return ""
}

func (cm *ClientManager) determineClient(packetData []byte) (uint8, error) {
	if len(packetData) < 20 {
		return 0, fmt.Errorf("packet too short for IP header")
	}

	sourceIP := fmt.Sprintf("%d.%d.%d.%d", packetData[12], packetData[13], packetData[14], packetData[15])
	destinationIP := fmt.Sprintf("%d.%d.%d.%d", packetData[16], packetData[17], packetData[18], packetData[19])

	if destinationIP == "10.0.0.1" {
		client, err := cm.GetClientByIP(sourceIP)
		if err != nil {
			return 0, fmt.Errorf("no client found for IP %s: %w", sourceIP, err)
		}
		return client.ID, nil
	}

	client, err := cm.GetClientByIP(destinationIP)
	if err != nil {
		return 0, fmt.Errorf("no client found for IP %s: %w", destinationIP, err)
	}
	return client.ID, nil
}
