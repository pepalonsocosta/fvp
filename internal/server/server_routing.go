package server

import (
	"log"
)

func (s *Server) routePackets() {
	defer s.wg.Done()
	
	for {
		select {
		case <-s.stopChan:
			return
		default:
			_, err := s.tunInterface.ReadPacket()
			if err != nil {
				log.Printf("TUN read error: %v", err)
				continue
			}
			
			s.processOutgoingPacket()
		}
	}
}

func (s *Server) processOutgoingPacket() {
	err := s.packetProcessor.ProcessOutgoingPacket()
	if err != nil {
		log.Printf("Packet processing error: %v", err)
	}
}
