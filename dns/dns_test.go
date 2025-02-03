package dns

import (
	"fmt"
	"testing"
)

// Helper function to create a basic DNS message
func createTestDNSMessage() *DNSMessage {
	return &DNSMessage{
		Header: DNSHeader{
			ID:      12345,
			Flags:   0x0100, // Standard query
			QDCount: 1,
			ANCount: 0,
			NSCount: 0,
			ARCount: 0,
		},
		Question: []DNSQuestion{
			{
				Name:  "yahoo.com",
				Type:  QTYPE_A,
				Class: QCLASS_IN,
			},
		},
	}
}

// Test DNS message packing and unpacking
func TestDNSMessagePackUnpack(t *testing.T) {

	// Create test scenarios
	testCases := []struct {
		name        string
		message     *DNSMessage
		expectError bool
	}{
		{
			name:        "Basic Query",
			message:     createTestDNSMessage(),
			expectError: false,
		},
		{
			name: "Multiple Answers",
			message: &DNSMessage{
				Header: DNSHeader{
					ID:      54321,
					Flags:   0x8180, // Response
					QDCount: 1,
					ANCount: 2,
					NSCount: 0,
					ARCount: 0,
				},
				Question: []DNSQuestion{
					{
						Name:  "yahoo.com",
						Type:  QTYPE_A,
						Class: QCLASS_IN,
					},
				},
				Answers: []DNSAnswer{
					{
						Name:     "yahoo.com",
						Type:     QTYPE_A,
						Class:    QCLASS_IN,
						TTL:      300,
						RDLength: 4,
						RData:    []byte{98, 137, 11, 163},
					},
					{
						Name:     "yahoo.com",
						Type:     QTYPE_A,
						Class:    QCLASS_IN,
						TTL:      300,
						RDLength: 4,
						RData:    []byte{74, 6, 143, 25},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Pack the message
			packedData, err := tc.message.Pack()
			if err != nil {
				t.Fatalf("Failed to pack message: %v", err)
			}

			// Log packed message details
			t.Logf("Packed message length: %d bytes", len(packedData))
			t.Logf("Packed message data: %v", packedData)

			// Unpack the message
			unpackedMsg, err := ParseDNSMessage(packedData)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to parse packed message: %v", err)
			}

			// Compare key message attributes
			if unpackedMsg.Header.ID != tc.message.Header.ID {
				t.Errorf("ID mismatch: got %d, want %d", unpackedMsg.Header.ID, tc.message.Header.ID)
			}

			if len(unpackedMsg.Question) != len(tc.message.Question) {
				t.Errorf("Question count mismatch: got %d, want %d",
					len(unpackedMsg.Question), len(tc.message.Question))
			}

			if len(unpackedMsg.Answers) != len(tc.message.Answers) {
				t.Errorf("Answer count mismatch: got %d, want %d",
					len(unpackedMsg.Answers), len(tc.message.Answers))
			}
		})
	}
}

// Test domain name encoding and decoding
func TestDomainNameEncoding(t *testing.T) {
	testCases := []string{
		"yahoo.com",
		"www.example.com",
		"test.domain.net",
		"",
	}

	for _, domain := range testCases {
		t.Run(fmt.Sprintf("Domain: %s", domain), func(t *testing.T) {
			// Encode domain
			encoded := encodeDomainName(domain)

			// Decode domain
			decoded, _, err := decodeDomainName(encoded, 0)
			if err != nil {
				t.Fatalf("Failed to decode domain: %v", err)
			}

			// Compare original and decoded
			if decoded != domain {
				t.Errorf("Domain mismatch: got %s, want %s", decoded, domain)
			}

			t.Logf("Encoded %s: %v", domain, encoded)
		})
	}
}

// Detailed byte-level inspection test
func TestMessageByteLevelInspection(t *testing.T) {
	// Create a sample DNS message
	msg := &DNSMessage{
		Header: DNSHeader{
			ID:      12345,
			Flags:   0x0100, // Standard query
			QDCount: 1,
			ANCount: 2,
			NSCount: 0,
			ARCount: 0,
		},
		Question: []DNSQuestion{
			{
				Name:  "yahoo.com",
				Type:  QTYPE_A,
				Class: QCLASS_IN,
			},
		},
		Answers: []DNSAnswer{
			{
				Name:     "yahoo.com",
				Type:     QTYPE_A,
				Class:    QCLASS_IN,
				TTL:      300,
				RDLength: 4,
				RData:    []byte{98, 137, 11, 163},
			},
			{
				Name:     "yahoo.com",
				Type:     QTYPE_A,
				Class:    QCLASS_IN,
				TTL:      300,
				RDLength: 4,
				RData:    []byte{74, 6, 143, 25},
			},
		},
	}

	// Pack the message
	packedData, err := msg.Pack()
	if err != nil {
		t.Fatalf("Failed to pack message: %v", err)
	}

	// Detailed byte-level logging
	t.Logf("Total packed message length: %d bytes", len(packedData))

	// Breakdown of message sections
	t.Logf("Header bytes: %v", packedData[:12])

	// Identify and log each section
	offset := 12
	t.Logf("Question section starts at offset %d", offset)

	// Log extra bytes if present
	if len(packedData) > offset {
		t.Logf("Extra bytes at end: %v", packedData[offset:])
		t.Logf("Number of extra bytes: %d", len(packedData)-offset)
	}
}
