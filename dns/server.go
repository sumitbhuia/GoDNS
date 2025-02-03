// File: dns/server.go
package dns

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type DNSServer struct {
	addr          string
	forwarderAddr string
	conn          *net.UDPConn
	running       bool
	mu            sync.Mutex
	wg            sync.WaitGroup
}

func NewDNSServer(addr string, forwarderAddr string) *DNSServer {
	if forwarderAddr == "" {
		forwarderAddr = "8.8.8.8:53" // Default to Google's DNS
	}
	return &DNSServer{
		addr:          addr,
		forwarderAddr: forwarderAddr,
		running:       false,
	}
}

func (s *DNSServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("DNS server already running")
	}

	addr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %v", err)
	}
	s.conn = conn
	s.running = true

	s.wg.Add(1)
	go s.serve()

	log.Printf("DNS Server listening on %s, forwarding to %s", s.addr, s.forwarderAddr)
	return nil
}

func (s *DNSServer) serve() {
	defer s.wg.Done()

	for s.isRunning() {
		buf := make([]byte, 1024)
		s.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, remoteAddr, err := s.conn.ReadFromUDP(buf)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if !s.isRunning() {
				return
			}
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		go s.handleQuery(buf[:n], remoteAddr)
	}
}

func (s *DNSServer) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *DNSServer) ListenAndServe() error {
	if err := s.Start(); err != nil {
		return err
	}
	s.wg.Wait()
	return nil
}

func (s *DNSServer) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	conn := s.conn
	s.mu.Unlock()

	if conn != nil {
		conn.Close()
	}
	s.wg.Wait()
	return nil
}

// // New implementation for createResponse
// func createResponse(msg *DNSMessage) *DNSMessage {
// 	response := &DNSMessage{
// 		Header: DNSHeader{
// 			ID:      msg.Header.ID,
// 			Flags:   0x8180, // Standard response flags
// 			QDCount: msg.Header.QDCount,
// 			ANCount: 0,
// 			NSCount: 0,
// 			ARCount: 0,
// 		},
// 		Question: msg.Question,
// 	}
// 	return response
// }

// New implementation for forwardQuery
func (s *DNSServer) forwardQuery(query []byte) (*DNSMessage, error) {
	forwarderAddr, err := net.ResolveUDPAddr("udp", s.forwarderAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, forwarderAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set timeouts
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Send query to forwarder
	_, err = conn.Write(query)
	if err != nil {
		return nil, err
	}

	// Larger buffer for full DNS responses
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	// Parse forwarded response
	return ParseDNSMessage(buf[:n])
}

func (s *DNSServer) handleQuery(query []byte, remoteAddr *net.UDPAddr) {
	msg, err := ParseDNSMessage(query)
	if err != nil {
		log.Printf("Error parsing DNS message: %v", err)
		return
	}

	// Validate query type
	if msg.Header.Flags&0x8000 != 0 {
		log.Printf("Received a response instead of a query")
		return
	}

	var response *DNSMessage

	// Always forward to external DNS
	forwardedResponse, err := s.forwardQuery(query)
	if err != nil {
		log.Printf("Error forwarding query: %v", err)
		response = createErrorResponse(msg)
	} else {
		response = forwardedResponse
	}

	// Preserve original query ID and add EDNS
	response.Header.ID = msg.Header.ID
	response.EDNS = &EDNSOption{
		Code:   TYPE_OPT,
		Length: 0,
	}
	response.Header.ARCount = 1

	// Pack and send response
	responseBytes, err := response.Pack()
	if err != nil {
		log.Printf("Error packing response: %v", err)
		return
	}

	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()

	if conn != nil {
		_, err = conn.WriteToUDP(responseBytes, remoteAddr)
		if err != nil {
			log.Printf("Error sending response: %v", err)
		}
	}
}

// createErrorResponse generates a DNS error response
func createErrorResponse(msg *DNSMessage) *DNSMessage {
	return &DNSMessage{
		Header: DNSHeader{
			ID:    msg.Header.ID, // Preserve original query ID
			Flags: 0x8183,        // Standard DNS error response flag
			// 0x8183 = Response (0x8000) + Server Failure (0x0002) + Recursion Desired (0x0100)
			QDCount: msg.Header.QDCount, // Preserve original question count
			ANCount: 0,                  // No answers
			NSCount: 0,                  // No name servers
			ARCount: 0,                  // No additional records
		},
		Question: msg.Question, // Preserve original questions
	}
}
