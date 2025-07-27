package dns

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

const (
	QTypeA uint16 = 1
	TypeA         = QTypeA

	QClassIN uint16 = 1
	ClassIN         = QClassIN

	FlagResponse           uint16 = 1 << 15
	FlagRecursionDesired   uint16 = 1 << 8
	FlagRecursionAvailable uint16 = 1 << 7

	DNSHeaderSize        = 12
	DefaultUDPBufferSize = 512
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

func (msg *DNSMessage) Pack() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, msg.Header)

	for _, q := range msg.Questions {
		buf.Write(encodeDomainName(q.Name))
		binary.Write(&buf, binary.BigEndian, q.Type)
		binary.Write(&buf, binary.BigEndian, q.Class)
	}

	for _, a := range msg.Answers {
		packRecord(&buf, a)
	}
	for _, a := range msg.Authority {
		packRecord(&buf, a)
	}
	for _, a := range msg.Additional {
		packRecord(&buf, a)
	}

	return buf.Bytes(), nil
}

func packRecord(buf *bytes.Buffer, record DNSRecord) {
	buf.Write(encodeDomainName(record.Name))
	binary.Write(buf, binary.BigEndian, record.Type)
	binary.Write(buf, binary.BigEndian, record.Class)
	binary.Write(buf, binary.BigEndian, record.TTL)
	binary.Write(buf, binary.BigEndian, record.RDLength)
	buf.Write(record.RData)
}

func ParseDNSMessage(data []byte) (*DNSMessage, error) {
	if len(data) < DNSHeaderSize {
		return nil, fmt.Errorf("message too short")
	}

	msg := &DNSMessage{}
	var err error

	msg.Header.ID = binary.BigEndian.Uint16(data[0:2])
	msg.Header.Flags = binary.BigEndian.Uint16(data[2:4])
	msg.Header.QDCount = binary.BigEndian.Uint16(data[4:6])
	msg.Header.ANCount = binary.BigEndian.Uint16(data[6:8])
	msg.Header.NSCount = binary.BigEndian.Uint16(data[8:10])
	msg.Header.ARCount = binary.BigEndian.Uint16(data[10:12])

	offset := DNSHeaderSize

	for i := uint16(0); i < msg.Header.QDCount; i++ {
		name, newOffset, err := decodeDomainName(data, offset)
		if err != nil {
			return nil, fmt.Errorf("parsing question name: %v", err)
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

	msg.Answers, offset, err = parseRecordSection(data, offset, msg.Header.ANCount)
	if err != nil {
		return nil, fmt.Errorf("parsing answers: %v", err)
	}

	msg.Authority, offset, err = parseRecordSection(data, offset, msg.Header.NSCount)
	if err != nil {
		return nil, fmt.Errorf("parsing authority records: %v", err)
	}

	msg.Additional, _, err = parseRecordSection(data, offset, msg.Header.ARCount)
	if err != nil {
		return nil, fmt.Errorf("parsing additional records: %v", err)
	}

	return msg, nil
}

func parseRecordSection(data []byte, offset int, count uint16) ([]DNSRecord, int, error) {
	records := make([]DNSRecord, 0, count)
	for i := uint16(0); i < count; i++ {
		record, newOffset, err := parseRecord(data, offset)
		if err != nil {
			return nil, offset, err
		}
		records = append(records, record)
		offset = newOffset
	}
	return records, offset, nil
}

func parseRecord(data []byte, offset int) (DNSRecord, int, error) {
	var record DNSRecord
	var err error

	record.Name, offset, err = decodeDomainName(data, offset)
	if err != nil {
		return record, offset, err
	}

	if offset+10 > len(data) {
		return record, offset, errors.New("truncated record header")
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

func encodeDomainName(domain string) []byte {
	var encoded bytes.Buffer
	if domain == "" || domain == "." {
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

func decodeDomainName(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", offset, errors.New("offset out of bounds")
	}

	var name strings.Builder
	originalOffset := offset
	visitedPointers := make(map[int]bool)

	for {
		if offset >= len(data) {
			return "", originalOffset, errors.New("truncated domain name")
		}

		length := int(data[offset])

		// Check for compression pointer (first two bits are 11).
		if length&0xC0 == 0xC0 {
			if offset+2 > len(data) {
				return "", originalOffset, errors.New("invalid compression pointer")
			}

			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			if visitedPointers[pointer] {
				return "", originalOffset, errors.New("compression loop detected")
			}
			visitedPointers[pointer] = true

			suffixName, _, err := decodeDomainName(data, pointer)
			if err != nil {
				return "", originalOffset, err
			}

			if name.Len() > 0 {
				name.WriteString(".")
			}
			name.WriteString(suffixName)

			return name.String(), originalOffset + 2, nil
		}

		offset++
		if length == 0 {
			if name.Len() == 0 {
				return ".", offset, nil
			}
			return name.String(), offset, nil
		}

		if offset+length > len(data) {
			return "", originalOffset, errors.New("label length exceeds message boundary")
		}

		if name.Len() > 0 {
			name.WriteString(".")
		}
		name.Write(data[offset : offset+length])
		offset += length
	}
}
