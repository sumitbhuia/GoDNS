package dns

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRoundTripAQuery tests if we can pack and parse a simple A record query.
func TestRoundTripAQuery(t *testing.T) {
	// 1. Create the original message
	originalMsg := &DNSMessage{
		Header: DNSHeader{
			ID:      1234,
			Flags:   FlagRecursionDesired, // A common flag for queries
			QDCount: 1,
		},
		Questions: []DNSQuestion{
			{
				Name:  "google.com",
				Type:  QTypeA,
				Class: QClassIN,
			},
		},
	}

	// 2. Pack the message
	packedBytes, err := originalMsg.Pack()
	require.NoError(t, err)
	require.NotEmpty(t, packedBytes)

	// 3. Parse the message
	parsedMsg, err := ParseDNSMessage(packedBytes)
	require.NoError(t, err)

	// 4. Assert that the parsed message is identical to the original
	require.Equal(t, originalMsg.Header, parsedMsg.Header)
	require.Equal(t, originalMsg.Questions, parsedMsg.Questions)
	require.Empty(t, parsedMsg.Answers) // Ensure no stray answers
}

// TestRoundTripAResponse tests packing and parsing a response with an answer.
func TestRoundTripAResponse(t *testing.T) {
	// 1. Create the original message
	originalMsg := &DNSMessage{
		Header: DNSHeader{
			ID:      5678,
			Flags:   FlagResponse | FlagRecursionAvailable,
			QDCount: 1,
			ANCount: 1,
		},
		Questions: []DNSQuestion{
			{
				Name:  "home.arpa",
				Type:  QTypeA,
				Class: QClassIN,
			},
		},
		Answers: []DNSRecord{
			{
				Name:     "home.arpa",
				Type:     TypeA,
				Class:    ClassIN,
				TTL:      300,
				RDLength: 4,
				RData:    net.ParseIP("192.168.1.1").To4(),
			},
		},
	}

	// 2. Pack it
	packedBytes, err := originalMsg.Pack()
	require.NoError(t, err)

	// 3. Parse it
	parsedMsg, err := ParseDNSMessage(packedBytes)
	require.NoError(t, err)

	// 4. Assert deep equality
	require.Equal(t, originalMsg.Header, parsedMsg.Header)
	require.Equal(t, originalMsg.Questions, parsedMsg.Questions)
	require.Equal(t, len(originalMsg.Answers), len(parsedMsg.Answers))
	require.Equal(t, originalMsg.Answers[0].Name, parsedMsg.Answers[0].Name)
	require.True(t, bytes.Equal(originalMsg.Answers[0].RData, parsedMsg.Answers[0].RData))
}
