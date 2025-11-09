package server

import (
	"fmt"
	"log"
	"net"

	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/network"
	"github.com/pepalonsocosta/fvp/internal/protocol"
)

type PacketProcessor struct {
	tunInterface  network.TUNInterface
	keyManager    *crypto.KeyManager
	clientManager *ClientManager
	udpConn       *net.UDPConn
}

func NewPacketProcessor(tunInterface network.TUNInterface, keyManager *crypto.KeyManager, clientManager *ClientManager, udpConn *net.UDPConn) *PacketProcessor {
	return &PacketProcessor{
		tunInterface:  tunInterface,
		keyManager:    keyManager,
		clientManager: clientManager,
		udpConn:       udpConn,
	}
}

func (pp *PacketProcessor) ProcessPacket(packetData []byte) error {
	
	packet, err := protocol.DecodePacket(packetData)
	if err != nil {
		return fmt.Errorf("failed to decode packet: %w", err)
	}
	
	client, err := pp.clientManager.GetClient(packet.ClientID)
	if err != nil {
		return fmt.Errorf("failed to get client %d: %w", packet.ClientID, err)
	}

	err = pp.clientManager.UpdateClientActivity(packet.ClientID, packet.Sequence)
	if err != nil {
		return fmt.Errorf("failed to update client activity: %w", err)
	}
	
	decryptedPayload, err := crypto.DecryptPayload(packet.Payload, client.Key, packet.Sequence)
	if err != nil {
		return fmt.Errorf("failed to decrypt payload for client %d: %w", packet.ClientID, err)
	}


	err = pp.tunInterface.WritePacket(decryptedPayload)
	if err != nil {
		return fmt.Errorf("failed to write packet for client %d: %w", packet.ClientID, err)
	}
	

	return nil
}

func (pp *PacketProcessor) ProcessOutgoingPacket() error {
	packetData, err := pp.tunInterface.ReadPacket()
	if err != nil {
		return fmt.Errorf("failed to read from TUN: %w", err)
	}

	clientID, err := pp.clientManager.determineClient(packetData)
	if err != nil {
		return err
	}

	client, err := pp.clientManager.GetClient(clientID)
	if err != nil {
		log.Printf("Unknown client %d: %v", clientID, err)
		return err
	}

	err = pp.createAndSendPacket(client, packetData)
	if err != nil {
		log.Printf("Failed to send packet to client %d: %v", clientID, err)
		return err
	}
	
	return nil
}

func (pp *PacketProcessor) createAndSendPacket(client *Client, ipData []byte) error {
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypeData,
		ClientID: client.ID,
		Sequence: client.LastSeq + 1,
		Length:   uint16(len(ipData)),
		Version:  protocol.ProtocolVersionByte,
		Payload:  ipData,
	}
	
	packetData, err := protocol.EncodePacket(packet)
	if err != nil {
		return fmt.Errorf("failed to encode packet: %w", err)
	}
	
	encrypted, err := crypto.EncryptPayload(packetData, client.Key, packet.Sequence)
	if err != nil {
		return fmt.Errorf("failed to encrypt packet: %w", err)
	}
	
	return pp.sendToClient(client, encrypted)
}

func (pp *PacketProcessor) sendToClient(client *Client, data []byte) error {
	// Parse client address
	addr, err := net.ResolveUDPAddr("udp", client.Address)
	if err != nil {
		return fmt.Errorf("failed to resolve client address: %w", err)
	}
	
	// Send data to client via UDP
	_, err = pp.udpConn.WriteToUDP(data, addr)
	if err != nil {
		return fmt.Errorf("failed to send data to client %d: %w", client.ID, err)
	}
	
	return nil
}
