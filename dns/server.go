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
	forwarderConn *net.UDPConn
	running       bool
	mu            sync.Mutex
	wg            sync.WaitGroup
}

func NewDNSServer(addr, forwarderAddr string) *DNSServer {
	if forwarderAddr == "" {
		forwarderAddr = "8.8.8.8:53"
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
		return fmt.Errorf("server already running")
	}

	// Resolve and listen on the server address
	addr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.conn = conn

	// Resolve and dial the forwarder address ONCE
	forwarderAddr, err := net.ResolveUDPAddr("udp", s.forwarderAddr)
	if err != nil {
		s.conn.Close() // Clean up listener
		return fmt.Errorf("failed to resolve forwarder address: %v", err)
	}

	forwarderConn, err := net.DialUDP("udp", nil, forwarderAddr)
	if err != nil {
		s.conn.Close()
		return fmt.Errorf("failed to dial forwarder: %v", err)
	}

	s.forwarderConn = forwarderConn

	s.running = true
	s.wg.Add(1)
	go s.serve()

	log.Printf("DNS Server listening on %s, forwarding to %s", s.addr, s.forwarderAddr)
	return nil
}

func (s *DNSServer) serve() {
	defer s.wg.Done()

	for s.isRunning() {
		buf := make([]byte, 512)
		s.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, remoteAddr, err := s.conn.ReadFromUDP(buf)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if !s.isRunning() {
				return
			}
			log.Printf("Read error: %v", err)
			continue
		}

		go s.handleQuery(buf[:n], remoteAddr)
	}
}

func (s *DNSServer) forwardQuery(query []byte) ([]byte, error) {
	conn := s.forwarderConn
	if conn == nil {
		return nil, fmt.Errorf("forwarder connection is not available")
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err := conn.Write(query)
	if err != nil {
		return nil, err
	}

	response := make([]byte, 512)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}

	return response[:n], nil
}

func (s *DNSServer) handleQuery(query []byte, remoteAddr *net.UDPAddr) {
	response, err := s.forwardQuery(query)
	if err != nil {
		log.Printf("Forward error: %v", err)
		return
	}

	if s.conn != nil {
		_, err = s.conn.WriteToUDP(response, remoteAddr)
		if err != nil {
			if s.isRunning() {
				log.Printf("Response error: %v", err)
			}
		}
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
	forwarderConn := s.forwarderConn // Get the forwarder connection
	s.mu.Unlock()

	if conn != nil {
		conn.Close()
	}
	if forwarderConn != nil {
		forwarderConn.Close() // Close the forwarder connection
	}

	s.wg.Wait()
	return nil
}
