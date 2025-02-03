// File: dns/message.go
package dns

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	// "log"
	"strings"
)

const (
	QTYPE_A       = 1  // Host address
	QCLASS_IN     = 1  // Internet
	TYPE_OPT      = 41 // EDNS0 OPT pseudo-RR
	CLASS_IN      = 1
	EDNS0_VERSION = 0
)

// Add EDNS0 specific type
type EDNSOption struct {
	Code   uint16
	Length uint16
	Data   []byte
}

type DNSHeader struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

type DNSQuestion struct {
	Name  string
	Type  uint16
	Class uint16
}

type DNSAnswer struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []byte
}

// Extend DNSMessage to include EDNS0 details
type DNSMessage struct {
	Header     DNSHeader
	Question   []DNSQuestion
	Answers    []DNSAnswer
	Authority  []DNSAnswer
	Additional []DNSAnswer
	EDNS       *EDNSOption
}

// encodeDomainName converts a domain name string to DNS wire format
func encodeDomainName(domain string) []byte {
	var encoded bytes.Buffer
	if domain == "" {
		encoded.WriteByte(0)
		return encoded.Bytes()
	}

	labels := strings.Split(strings.Trim(domain, "."), ".")
	for _, label := range labels {
		encoded.WriteByte(byte(len(label)))
		encoded.WriteString(label)
	}
	encoded.WriteByte(0) // Root label
	return encoded.Bytes()
}

// Enhanced decodeDomainName to handle compression
func decodeDomainName(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", offset, errors.New("offset out of bounds")
	}

	var name strings.Builder

	visited := make(map[int]bool)

	for {
		if offset >= len(data) {
			return "", offset, errors.New("truncated domain name")
		}

		// Check for compression pointer
		if data[offset]&0xC0 == 0xC0 {
			if offset+1 >= len(data) {
				return "", offset, errors.New("incomplete compression pointer")
			}

			// Extract compression pointer
			pointerOffset := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)

			// Prevent infinite loops
			if visited[pointerOffset] {
				return "", offset, errors.New("circular compression pointer")
			}
			visited[pointerOffset] = true

			// Temporarily change offset and continue parsing
			offset = pointerOffset
			continue
		}

		// Length of current label
		length := int(data[offset])

		// Null terminator
		if length == 0 {
			offset++
			break
		}

		// Ensure we don't go out of bounds
		if offset+1+length > len(data) {
			return "", offset, errors.New("label exceeds message length")
		}

		// Add dot for subsequent labels
		if name.Len() > 0 {
			name.WriteRune('.')
		}

		// Write label
		name.Write(data[offset+1 : offset+1+length])

		// Move offset
		offset += 1 + length
	}

	return name.String(), offset, nil
}

func (msg *DNSMessage) Pack() ([]byte, error) {
	var buf bytes.Buffer

	// Recalculate counts
	msg.Header.QDCount = uint16(len(msg.Question))
	msg.Header.ANCount = uint16(len(msg.Answers))
	msg.Header.NSCount = uint16(len(msg.Authority))
	msg.Header.ARCount = uint16(len(msg.Additional))

	// Ensure EDNS is handled correctly
	if msg.EDNS == nil {
		msg.EDNS = &EDNSOption{
			Code:   TYPE_OPT,
			Length: 0,
		}
		msg.Header.ARCount = 1
	}

	// Write header
	binary.Write(&buf, binary.BigEndian, msg.Header)

	// Write questions
	for _, q := range msg.Question {
		buf.Write(encodeDomainName(q.Name))
		binary.Write(&buf, binary.BigEndian, q.Type)
		binary.Write(&buf, binary.BigEndian, q.Class)
	}

	// Write answers
	for _, a := range msg.Answers {
		buf.Write(encodeDomainName(a.Name))
		binary.Write(&buf, binary.BigEndian, a.Type)
		binary.Write(&buf, binary.BigEndian, a.Class)
		binary.Write(&buf, binary.BigEndian, a.TTL)
		binary.Write(&buf, binary.BigEndian, a.RDLength)
		buf.Write(a.RData)
	}

	// Write authority records
	for _, ns := range msg.Authority {
		buf.Write(encodeDomainName(ns.Name))
		binary.Write(&buf, binary.BigEndian, ns.Type)
		binary.Write(&buf, binary.BigEndian, ns.Class)
		binary.Write(&buf, binary.BigEndian, ns.TTL)
		binary.Write(&buf, binary.BigEndian, ns.RDLength)
		buf.Write(ns.RData)
	}

	// Write EDNS pseudosection explicitly
	if msg.EDNS != nil {
		buf.WriteByte(0) // Root name
		binary.Write(&buf, binary.BigEndian, uint16(TYPE_OPT))
		binary.Write(&buf, binary.BigEndian, uint16(4096)) // UDP payload size
		buf.WriteByte(EDNS0_VERSION)
		buf.WriteByte(0)                                // No extended flags
		binary.Write(&buf, binary.BigEndian, uint16(0)) // No RDATA
	}

	return buf.Bytes(), nil
}

func ParseDNSMessage(data []byte) (*DNSMessage, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("message too short: %d bytes", len(data))
	}

	msg := &DNSMessage{}

	// Parse Header
	msg.Header.ID = binary.BigEndian.Uint16(data[0:2])
	msg.Header.Flags = binary.BigEndian.Uint16(data[2:4])
	msg.Header.QDCount = binary.BigEndian.Uint16(data[4:6])
	msg.Header.ANCount = binary.BigEndian.Uint16(data[6:8])
	msg.Header.NSCount = binary.BigEndian.Uint16(data[8:10])
	msg.Header.ARCount = binary.BigEndian.Uint16(data[10:12])

	offset := 12

	// Parse Questions
	for i := uint16(0); i < msg.Header.QDCount; i++ {
		name, newOffset, err := decodeDomainName(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing question name: %v", err)
		}

		if newOffset+4 > len(data) {
			return nil, errors.New("truncated question section")
		}

		question := DNSQuestion{
			Name:  name,
			Type:  binary.BigEndian.Uint16(data[newOffset : newOffset+2]),
			Class: binary.BigEndian.Uint16(data[newOffset+2 : newOffset+4]),
		}
		msg.Question = append(msg.Question, question)
		offset = newOffset + 4
	}

	// Sections to parse
	sections := []struct {
		count   uint16
		records *[]DNSAnswer
		name    string
	}{
		{msg.Header.ANCount, &msg.Answers, "answer"},
		{msg.Header.NSCount, &msg.Authority, "authority"},
		{msg.Header.ARCount, &msg.Additional, "additional"},
	}

	for _, section := range sections {
		for i := uint16(0); i < section.count; i++ {
			if offset >= len(data) {
				break
			}

			name, newOffset, err := decodeDomainName(data, offset)
			if err != nil {
				return nil, fmt.Errorf("error parsing %s section name: %v", section.name, err)
			}

			if newOffset+10 > len(data) {
				return nil, fmt.Errorf("truncated %s section", section.name)
			}

			recType := binary.BigEndian.Uint16(data[newOffset : newOffset+2])
			recClass := binary.BigEndian.Uint16(data[newOffset+2 : newOffset+4])

			// EDNS handling
			if recType == TYPE_OPT {
				rdLength := binary.BigEndian.Uint16(data[newOffset+8 : newOffset+10])

				// Ensure we don't go out of bounds
				if newOffset+10+int(rdLength) > len(data) {
					return nil, fmt.Errorf("truncated EDNS record")
				}

				msg.EDNS = &EDNSOption{
					Code:   recType,
					Length: rdLength,
					Data:   data[newOffset+10 : newOffset+10+int(rdLength)],
				}
				offset = newOffset + 10 + int(rdLength)
				continue
			}

			rdLength := binary.BigEndian.Uint16(data[newOffset+8 : newOffset+10])

			// Ensure we don't go out of bounds
			if newOffset+10+int(rdLength) > len(data) {
				return nil, fmt.Errorf("truncated rdata in %s section", section.name)
			}

			answer := DNSAnswer{
				Name:     name,
				Type:     recType,
				Class:    recClass,
				TTL:      binary.BigEndian.Uint32(data[newOffset+4 : newOffset+8]),
				RDLength: rdLength,
				RData:    data[newOffset+10 : newOffset+10+int(rdLength)],
			}

			*section.records = append(*section.records, answer)
			offset = newOffset + 10 + int(rdLength)
		}
	}

	// Handle extra bytes
	if offset < len(data) {
		fmt.Printf("Warning: %d extra bytes at end of DNS message\n", len(data)-offset)
	}

	return msg, nil
}
