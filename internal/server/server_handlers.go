package server

import (
	"log"
	"net"
	"time"

	"github.com/pepalonsocosta/fvp/internal/config"
	"github.com/pepalonsocosta/fvp/internal/crypto"
	"github.com/pepalonsocosta/fvp/internal/protocol"
)

func (s *Server) handleClients() {
	defer s.wg.Done()
	
	buffer := make([]byte, config.DefaultMTU)
	
	for {
		select {
		case <-s.stopChan:
			return
		default:
			s.udpConn.SetReadDeadline(time.Now().Add(config.DefaultUDPReadTimeout))
			
			n, clientAddr, err := s.udpConn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				// Only log non-timeout errors
				log.Printf("UDP read error: %v", err)
				continue
			}
			
			s.processClientPacket(buffer[:n], clientAddr)
		}
	}
}

func (s *Server) processClientPacket(data []byte, clientAddr *net.UDPAddr) {
	packet, err := protocol.DecodePacket(data)
	if err != nil {
		log.Printf("Failed to decode packet from %s: %v", clientAddr, err)
		return
	}
	
	switch packet.Type {
	case protocol.PacketTypeAuth:
		s.handleAuthPacket(packet, clientAddr)
	case protocol.PacketTypeData:
		s.handleDataPacket(packet, clientAddr)
	case protocol.PacketTypePing:
		s.handlePingPacket(packet, clientAddr)
	case protocol.PacketTypePong:
		s.handlePongPacket(packet, clientAddr)
	default:
		// Silently drop unknown packet types (common for malformed packets)
	}
}

func (s *Server) handleAuthPacket(packet *protocol.Packet, clientAddr *net.UDPAddr) {
	var clientID uint8
	var key []byte
	var err error
	
	if packet.ClientID == 0 {
		// Request assignment - server generates key and assigns ID
		key = s.generateRandomKey()
		clientID = s.clientManager.findNextClientID()
		if clientID == 0 {
			log.Printf("Authentication failed: no available client IDs from %s", clientAddr)
			return
		}
		log.Printf("New client requesting assignment from %s, assigned ID %d", clientAddr, clientID)
	} else {
		// Pre-shared key - use existing key
		if !s.keyManager.HasClient(packet.ClientID) {
			log.Printf("Authentication failed: unknown client ID %d from %s", packet.ClientID, clientAddr)
			return
		}
		
		key, err = s.keyManager.GetClientKey(packet.ClientID)
		if err != nil {
			log.Printf("Authentication failed: could not get key for client %d from %s: %v", packet.ClientID, clientAddr, err)
			return
		}
		clientID = packet.ClientID
		log.Printf("Existing client %d authenticating from %s", clientID, clientAddr)
	}
	
	client, err := s.clientManager.AddClient(key, clientAddr.String())
	if err != nil {
		log.Printf("Authentication failed: could not add client %d from %s: %v", clientID, clientAddr, err)
		return
	}
	
	log.Printf("Client %d connected from %s, assigned IP %s", client.ID, clientAddr, client.IP)
	
	err = s.sendAuthResponse(client.ID, client.IP, key, clientAddr)
	if err != nil {
		log.Printf("Failed to send auth response to client %d: %v", client.ID, err)
	}
}

func (s *Server) handleDataPacket(packet *protocol.Packet, clientAddr *net.UDPAddr) {
	packetData, err := protocol.EncodePacket(packet)
	if err != nil {
		log.Printf("Failed to encode packet from client %d: %v", packet.ClientID, err)
		return
	}
	
	err = s.packetProcessor.ProcessPacket(packetData)
	if err != nil {
		log.Printf("Failed to process data packet from client %d: %v", packet.ClientID, err)
		return
	}
}

func (s *Server) handlePingPacket(packet *protocol.Packet, clientAddr *net.UDPAddr) {
	err := s.clientManager.UpdateClientActivity(packet.ClientID, packet.Sequence)
	if err != nil {
		log.Printf("Failed to update client activity for ping from client %d: %v", packet.ClientID, err)
		return
	}
	
	err = s.sendPongResponse(packet.ClientID, packet.Sequence)
	if err != nil {
		log.Printf("Failed to send pong response to client %d: %v", packet.ClientID, err)
	}
	
	log.Printf("Received ping from client %d", packet.ClientID)
}

func (s *Server) handlePongPacket(packet *protocol.Packet, clientAddr *net.UDPAddr) {
	err := s.clientManager.UpdateClientActivity(packet.ClientID, packet.Sequence)
	if err != nil {
		log.Printf("Failed to update client activity for pong from client %d: %v", packet.ClientID, err)
		return
	}
	
	log.Printf("Received pong from client %d", packet.ClientID)
}

// generateRandomKey generates a random 32-byte key for new clients
func (s *Server) generateRandomKey() []byte {
	key, err := crypto.GenerateKeyBytes()
	if err != nil {
		// This should never happen, but log if it does
		log.Printf("Warning: failed to generate random key: %v", err)
		// Return zero key as fallback (will fail validation)
		return make([]byte, 32)
	}
	return key
}
