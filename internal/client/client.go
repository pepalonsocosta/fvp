package client

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/network"
	"github.com/pepalonsocosta/fvp/internal/protocol"
)

// Client represents a VPN client
type Client struct {
	serverAddr     string
	clientID       uint8
	key            []byte
	assignedIP     string
	tunInterface   network.TUNInterface
	udpConn        *net.UDPConn
	sequence       uint32
	connected      bool
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewClient creates a new VPN client
func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr:   serverAddr,
		clientID:     0, // Will be assigned by server
		key:          nil, // Will be assigned by server
		assignedIP:   "", // Will be assigned by server
		tunInterface: network.NewTunManager(),
		sequence:     1,
		connected:    false,
		stopChan:     make(chan struct{}),
	}
}

func (c *Client) Connect() error {
	log.Printf("Connecting to VPN server at %s", c.serverAddr)

	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve server address: %w", err)
	}

	c.udpConn, err = net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	err = c.sendAuthRequest()
	if err != nil {
		c.udpConn.Close()
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	err = c.waitForAuthResponse()
	if err != nil {
		c.udpConn.Close()
		return fmt.Errorf("authentication failed: %w", err)
	}

	err = c.tunInterface.Create("fvp-client0")
	if err != nil {
		c.udpConn.Close()
		return fmt.Errorf("failed to create TUN interface: %w", err)
	}

	// Step 5: Configure TUN interface with assigned IP
	err = c.tunInterface.ConfigureClientInterface(c.assignedIP)
	if err != nil {
		c.tunInterface.Close()
		c.udpConn.Close()
		return fmt.Errorf("failed to configure TUN interface: %w", err)
	}
	
	log.Printf("TUN interface configured with IP %s", c.assignedIP)

	// Step 6: Start packet processing
	c.connected = true
	c.startPacketProcessing()

	log.Printf("Successfully connected to VPN server. Client ID: %d, IP: %s", c.clientID, c.assignedIP)
	return nil
}

// Disconnect closes the VPN connection
func (c *Client) Disconnect() error {
	log.Printf("Disconnecting from VPN server")

	c.connected = false

	// Signal all goroutines to stop
	close(c.stopChan)

	// Wait for all goroutines to finish
	c.wg.Wait()

	// Close connections
	if c.udpConn != nil {
		c.udpConn.Close()
	}
	if c.tunInterface != nil {
		c.tunInterface.Close()
	}

	log.Printf("Disconnected from VPN server")
	return nil
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) GetClientID() uint8 {
	return c.clientID
}

func (c *Client) GetAssignedIP() string {
	return c.assignedIP
}

func (c *Client) sendAuthRequest() error {
	authPacket := protocol.CreateAuthPacket(c.clientID, c.sequence, []byte{})
	
	packetData, err := protocol.EncodePacket(authPacket)
	if err != nil {
		return fmt.Errorf("failed to encode auth packet: %w", err)
	}

	_, err = c.udpConn.Write(packetData)
	if err != nil {
		return fmt.Errorf("failed to send auth packet: %w", err)
	}

	log.Printf("Sent authentication request to server")
	return nil
}

func (c *Client) waitForAuthResponse() error {
	c.udpConn.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 1500)
	n, err := c.udpConn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	packet, err := protocol.DecodePacket(buffer[:n])
	if err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	if packet.Type != protocol.PacketTypeAuth {
		return fmt.Errorf("expected auth response, got packet type %d", packet.Type)
	}

	if len(packet.Payload) < 32 {
		return fmt.Errorf("invalid auth response payload length")
	}

	c.clientID = packet.ClientID
	c.key = make([]byte, 32)
	copy(c.key, packet.Payload[:32])
	c.assignedIP = string(packet.Payload[32:])

	log.Printf("Received authentication response: Client ID %d, IP %s", c.clientID, c.assignedIP)
	return nil
}

func (c *Client) startPacketProcessing() {
	c.wg.Add(1)
	go c.handleServerPackets()

	c.wg.Add(1)
	go c.handleTUNPackets()

	c.wg.Add(1)
	go c.sendKeepAlive()

	log.Printf("Started packet processing goroutines")
}

func (c *Client) handleServerPackets() {
	defer c.wg.Done()

	buffer := make([]byte, 1500)
	for {
		select {
		case <-c.stopChan:
			log.Printf("Server packet handler stopped")
			return
		default:
			c.udpConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, err := c.udpConn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("Error reading from server: %v", err)
				continue
			}

			c.processServerPacket(buffer[:n])
		}
	}
}

func (c *Client) handleTUNPackets() {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopChan:
			log.Printf("TUN packet handler stopped")
			return
		default:
			packetData, err := c.tunInterface.ReadPacket()
			if err != nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			c.processTUNPacket(packetData)
		}
	}
}

func (c *Client) processServerPacket(data []byte) {
	packet, err := protocol.DecodePacket(data)
	if err != nil {
		log.Printf("Failed to decode server packet: %v", err)
		return
	}

	switch packet.Type {
	case protocol.PacketTypeData:
		c.handleDataPacket(packet)
	case protocol.PacketTypePong:
		c.handlePongPacket(packet)
	default:
		log.Printf("Unknown packet type %d from server", packet.Type)
	}
}

func (c *Client) processTUNPacket(data []byte) {
	encryptedData, err := crypto.EncryptPayload(data, c.key, c.sequence)
	if err != nil {
		log.Printf("Failed to encrypt packet: %v", err)
		return
	}

	dataPacket := protocol.CreateDataPacket(c.clientID, c.sequence, encryptedData)
	
	packetData, err := protocol.EncodePacket(dataPacket)
	if err != nil {
		log.Printf("Failed to encode data packet: %v", err)
		return
	}

	_, err = c.udpConn.Write(packetData)
	if err != nil {
		log.Printf("Failed to send data packet to server: %v", err)
		return
	}

	c.sequence++
}

func (c *Client) handleDataPacket(packet *protocol.Packet) {
	decryptedData, err := crypto.DecryptPayload(packet.Payload, c.key, packet.Sequence)
	if err != nil {
		log.Printf("Failed to decrypt data packet: %v", err)
		return
	}

	err = c.tunInterface.WritePacket(decryptedData)
	if err != nil {
		log.Printf("Failed to write packet to TUN interface: %v", err)
		return
	}
}

func (c *Client) handlePongPacket(packet *protocol.Packet) {
	log.Printf("Received pong from server (sequence %d)", packet.Sequence)
}

func (c *Client) sendKeepAlive() {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.sendPing()
		}
	}
}

func (c *Client) sendPing() {
	pingPacket := protocol.CreatePingPacket(c.clientID, c.sequence)
	
	packetData, err := protocol.EncodePacket(pingPacket)
	if err != nil {
		log.Printf("Failed to encode ping packet: %v", err)
		return
	}

	_, err = c.udpConn.Write(packetData)
	if err != nil {
		log.Printf("Failed to send ping packet: %v", err)
		return
	}

	c.sequence++
}
