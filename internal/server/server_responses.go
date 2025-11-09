package server

import (
	"fmt"
	"log"
	"net"

	"github.com/pepalonsocosta/fvp/internal/protocol"
)

func (s *Server) sendAuthResponse(clientID uint8, clientIP string, key []byte, clientAddr *net.UDPAddr) error {
	// Create response payload with key and IP
	// Format: [32-byte key][IP string]
	payload := make([]byte, 32+len(clientIP))
	copy(payload[:32], key)
	copy(payload[32:], []byte(clientIP))
	
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypeAuth,
		ClientID: clientID,
		Sequence: 0, // Auth response uses sequence 0
		Length:   uint16(len(payload)),
		Version:  protocol.ProtocolVersionByte,
		Payload:  payload,
	}
	
	packetData, err := protocol.EncodePacket(packet)
	if err != nil {
		return fmt.Errorf("failed to encode auth response: %w", err)
	}
	
	_, err = s.udpConn.WriteToUDP(packetData, clientAddr)
	if err != nil {
		return fmt.Errorf("failed to send auth response: %w", err)
	}
	
	log.Printf("Sent auth response to client %d with IP %s", clientID, clientIP)
	return nil
}

func (s *Server) sendPongResponse(clientID uint8, sequence uint32) error {
	client, err := s.clientManager.GetClient(clientID)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}
	
	clientAddr, err := net.ResolveUDPAddr("udp", client.Address)
	if err != nil {
		return fmt.Errorf("invalid client address: %w", err)
	}
	
	packet := &protocol.Packet{
		Magic:    [3]byte{'F', 'V', 'P'},
		Type:     protocol.PacketTypePong,
		ClientID: clientID,
		Sequence: sequence, // Echo back the same sequence
		Length:   0,        // Empty payload
		Version:  protocol.ProtocolVersionByte,
		Payload:  []byte{}, // Empty payload for pong response
	}
	
	packetData, err := protocol.EncodePacket(packet)
	if err != nil {
		return fmt.Errorf("failed to encode pong response: %w", err)
	}
	
	_, err = s.udpConn.WriteToUDP(packetData, clientAddr)
	if err != nil {
		return fmt.Errorf("failed to send pong response: %w", err)
	}
	
	log.Printf("Sent pong response to client %d", clientID)
	return nil
}
