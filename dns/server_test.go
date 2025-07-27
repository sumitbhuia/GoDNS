package dns

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockForwarder listens on a UDP port and sends a fixed response.
func mockForwarder(t *testing.T, responseToSend []byte) *net.UDPAddr {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	require.NoError(t, err)

	t.Cleanup(func() { conn.Close() }) // Ensure connection is closed after test

	go func() {
		buf := make([]byte, 512)
		_, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return // Can happen when conn is closed
		}
		conn.WriteToUDP(responseToSend, remoteAddr)
	}()

	return conn.LocalAddr().(*net.UDPAddr)
}

func TestServerForwarding(t *testing.T) {
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	// 1. Prepare a fake query and response
	query := &DNSMessage{Header: DNSHeader{ID: 111, QDCount: 1}, Questions: []DNSQuestion{{Name: "test.com", Type: QTypeA, Class: QClassIN}}}
	queryBytes, err := query.Pack()
	require.NoError(t, err)

	response := &DNSMessage{Header: DNSHeader{ID: 111, ANCount: 1, Flags: FlagResponse}, Answers: []DNSRecord{{Name: "test.com", RData: []byte{1, 2, 3, 4}}}}
	responseBytes, err := response.Pack()
	require.NoError(t, err)

	// 2. Start the mock forwarder
	mockAddr := mockForwarder(t, responseBytes)

	// 3. Start your DNS server, configured to use the mock forwarder
	server := NewDNSServer("127.0.0.1:0", mockAddr.String()) // Port 0 asks OS for a free port
	err = server.Start()
	require.NoError(t, err)
	t.Cleanup(func() { server.Stop() }) // Ensure server is stopped after test

	// 4. Act as a client: send a query to YOUR server
	clientConn, err := net.Dial("udp", server.conn.LocalAddr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	_, err = clientConn.Write(queryBytes)
	require.NoError(t, err)

	// 5. Receive the response from YOUR server
	respBuf := make([]byte, 512)
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := clientConn.Read(respBuf)
	require.NoError(t, err)

	// 6. Assert that the response is the one from our mock forwarder
	require.True(t, bytes.Equal(responseBytes, respBuf[:n]), "The response from the server should match the mock forwarder's response")
}

func BenchmarkServerE2E(b *testing.B) {
	log.SetOutput(io.Discard)
	b.Cleanup(func() { log.SetOutput(os.Stderr) })

	query := &DNSMessage{Header: DNSHeader{ID: 111, QDCount: 1}, Questions: []DNSQuestion{{Name: "test.com", Type: QTypeA, Class: QClassIN}}}
	queryBytes, err := query.Pack()
	if err != nil {
		b.Fatalf("Failed to pack query: %v", err)
	}

	response := &DNSMessage{Header: DNSHeader{ID: 111, ANCount: 1, Flags: FlagResponse}, Answers: []DNSRecord{{Name: "test.com", RData: []byte{1, 2, 3, 4}}}}
	responseBytes, err := response.Pack()
	if err != nil {
		b.Fatalf("Failed to pack response: %v", err)
	}

	mockAddr := mockForwarderForBenchmark(b, responseBytes)

	server := NewDNSServer("127.0.0.1:0", mockAddr.String())
	err = server.Start()
	if err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	clientConn, err := net.Dial("udp", server.conn.LocalAddr().String())
	if err != nil {
		b.Fatalf("Failed to dial server: %v", err)
	}
	defer clientConn.Close()

	respBuf := make([]byte, 512)

	// --- Benchmark Loop ---
	b.ResetTimer() // Start timing now
	for i := 0; i < b.N; i++ {
		clientConn.Write(queryBytes)
		clientConn.Read(respBuf)
	}
}

func mockForwarderForBenchmark(b *testing.B, responseToSend []byte) *net.UDPAddr {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		b.Fatalf("mock setup failed: %v", err)
	}
	b.Cleanup(func() { conn.Close() })
	go func() {
		buf := make([]byte, 512)
		for {
			_, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			conn.WriteToUDP(responseToSend, remoteAddr)
		}
	}()
	return conn.LocalAddr().(*net.UDPAddr)
}

func BenchmarkServerThroughput(b *testing.B) {
	log.SetOutput(io.Discard)
	b.Cleanup(func() { log.SetOutput(os.Stderr) }) // Restore the logger after the benchmark

	query := &DNSMessage{Header: DNSHeader{ID: 111, QDCount: 1}, Questions: []DNSQuestion{{Name: "test.com", Type: QTypeA, Class: QClassIN}}}
	queryBytes, err := query.Pack()
	if err != nil {
		b.Fatalf("Failed to pack query: %v", err)
	}

	response := &DNSMessage{Header: DNSHeader{ID: 111, ANCount: 1, Flags: FlagResponse}, Answers: []DNSRecord{{Name: "test.com", RData: []byte{1, 2, 3, 4}}}}
	responseBytes, err := response.Pack()
	if err != nil {
		b.Fatalf("Failed to pack response: %v", err)
	}

	mockAddr := mockForwarderForBenchmark(b, responseBytes)

	server := NewDNSServer("127.0.0.1:0", mockAddr.String())
	err = server.Start()
	if err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	clientConn, err := net.Dial("udp", server.conn.LocalAddr().String())
	if err != nil {
		b.Fatalf("Failed to dial server: %v", err)
	}
	defer clientConn.Close()

	respBuf := make([]byte, 512)

	b.ResetTimer()
	// run the query logic in parallel across multiple goroutines.
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			clientConn.Write(queryBytes)
			clientConn.Read(respBuf)
		}
	})
}
