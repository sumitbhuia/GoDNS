// File: dns/message.go
package dns

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

const (
	QTYPE_A   = 1  // Host address
	QCLASS_IN = 1  // Internet
	TYPE_OPT  = 41 // EDNS0 OPT record
	CLASS_IN  = 1
)

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

type DNSRecord struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []byte
}

type DNSMessage struct {
	Header     DNSHeader
	Questions  []DNSQuestion
	Answers    []DNSRecord
	Authority  []DNSRecord
	Additional []DNSRecord
}

// encodeDomainName converts a domain name string to DNS wire format
func encodeDomainName(domain string) []byte {
	var encoded bytes.Buffer

	// Special case for empty domain
	if domain == "" {
		encoded.WriteByte(0)
		return encoded.Bytes()
	}

	// Special case for root domain
	if domain == "." {
		encoded.WriteByte(0)
		return encoded.Bytes()
	}

	domain = strings.TrimSuffix(domain, ".")

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		encoded.WriteByte(byte(len(label)))
		encoded.WriteString(label)
	}
	encoded.WriteByte(0)
	return encoded.Bytes()
}

// Enhanced decodeDomainName to handle compression
func decodeDomainName(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", offset, errors.New("offset out of bounds")
	}

	var name strings.Builder
	originalOffset := offset
	visited := make(map[int]bool)

	// Special case: if first byte is 0, return empty string for empty domain
	if data[offset] == 0 {
		return "", offset + 1, nil
	}

	for {
		if offset >= len(data) {
			return "", originalOffset, errors.New("truncated domain name")
		}

		length := int(data[offset])

		// Check for compression pointer
		if length&0xC0 == 0xC0 {
			if offset+2 > len(data) {
				return "", originalOffset, errors.New("invalid compression pointer")
			}

			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			if visited[pointer] {
				return "", originalOffset, errors.New("compression loop detected")
			}
			visited[pointer] = true

			if pointer >= offset {
				return "", originalOffset, errors.New("forward compression pointer")
			}

			suffixName, _, err := decodeDomainName(data, pointer)
			if err != nil {
				return "", originalOffset, err
			}

			if name.Len() > 0 {
				name.WriteString(".")
			}
			name.WriteString(suffixName)
			return name.String(), offset + 2, nil
		}

		// Regular label
		if length == 0 {
			if name.Len() == 0 {
				return ".", offset + 1, nil
			}
			return name.String(), offset + 1, nil
		}

		if offset+1+length > len(data) {
			return "", originalOffset, errors.New("label exceeds message")
		}

		if name.Len() > 0 {
			name.WriteString(".")
		}
		name.Write(data[offset+1 : offset+1+length])
		offset += 1 + length
	}
}

func (msg *DNSMessage) Pack() ([]byte, error) {
	var buf bytes.Buffer

	// Write header
	binary.Write(&buf, binary.BigEndian, msg.Header)

	// Write questions
	for _, q := range msg.Questions {
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

	// Write authority
	for _, a := range msg.Authority {
		buf.Write(encodeDomainName(a.Name))
		binary.Write(&buf, binary.BigEndian, a.Type)
		binary.Write(&buf, binary.BigEndian, a.Class)
		binary.Write(&buf, binary.BigEndian, a.TTL)
		binary.Write(&buf, binary.BigEndian, a.RDLength)
		buf.Write(a.RData)
	}

	// Write additional
	for _, a := range msg.Additional {
		buf.Write(encodeDomainName(a.Name))
		binary.Write(&buf, binary.BigEndian, a.Type)
		binary.Write(&buf, binary.BigEndian, a.Class)
		binary.Write(&buf, binary.BigEndian, a.TTL)
		binary.Write(&buf, binary.BigEndian, a.RDLength)
		buf.Write(a.RData)
	}

	return buf.Bytes(), nil
}

func ParseDNSMessage(data []byte) (*DNSMessage, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("message too short")
	}

	msg := &DNSMessage{}

	// Parse header
	msg.Header.ID = binary.BigEndian.Uint16(data[0:2])
	msg.Header.Flags = binary.BigEndian.Uint16(data[2:4])
	msg.Header.QDCount = binary.BigEndian.Uint16(data[4:6])
	msg.Header.ANCount = binary.BigEndian.Uint16(data[6:8])
	msg.Header.NSCount = binary.BigEndian.Uint16(data[8:10])
	msg.Header.ARCount = binary.BigEndian.Uint16(data[10:12])

	offset := 12

	// Parse questions
	for i := uint16(0); i < msg.Header.QDCount; i++ {
		name, newOffset, err := decodeDomainName(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing question: %v", err)
		}
		offset = newOffset

		if offset+4 > len(data) {
			return nil, errors.New("truncated question section")
		}

		question := DNSQuestion{
			Name:  name,
			Type:  binary.BigEndian.Uint16(data[offset : offset+2]),
			Class: binary.BigEndian.Uint16(data[offset+2 : offset+4]),
		}
		msg.Questions = append(msg.Questions, question)
		offset += 4
	}

	// Parse answers
	for i := uint16(0); i < msg.Header.ANCount; i++ {
		record, newOffset, err := parseRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing answer: %v", err)
		}
		msg.Answers = append(msg.Answers, record)
		offset = newOffset
	}

	// Parse authority
	for i := uint16(0); i < msg.Header.NSCount; i++ {
		record, newOffset, err := parseRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing authority: %v", err)
		}
		msg.Authority = append(msg.Authority, record)
		offset = newOffset
	}

	// Parse additional
	for i := uint16(0); i < msg.Header.ARCount; i++ {
		record, newOffset, err := parseRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing additional: %v", err)
		}
		msg.Additional = append(msg.Additional, record)
		offset = newOffset
	}

	return msg, nil
}

func parseRecord(data []byte, offset int) (DNSRecord, int, error) {
	var record DNSRecord
	var err error

	record.Name, offset, err = decodeDomainName(data, offset)
	if err != nil {
		return record, offset, err
	}

	if offset+10 > len(data) {
		return record, offset, errors.New("truncated record")
	}

	record.Type = binary.BigEndian.Uint16(data[offset : offset+2])
	record.Class = binary.BigEndian.Uint16(data[offset+2 : offset+4])
	record.TTL = binary.BigEndian.Uint32(data[offset+4 : offset+8])
	record.RDLength = binary.BigEndian.Uint16(data[offset+8 : offset+10])
	offset += 10

	if offset+int(record.RDLength) > len(data) {
		return record, offset, errors.New("truncated rdata")
	}

	record.RData = make([]byte, record.RDLength)
	copy(record.RData, data[offset:offset+int(record.RDLength)])
	offset += int(record.RDLength)

	return record, offset, nil
}
