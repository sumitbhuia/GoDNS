# ⚡ GoDNS - High-Performance, RFC-Compliant Recursive DNS Resolver

A concurrent DNS protocol implementation in Go, featuring non-blocking I/O architecture, RFC1035-compliant message processing, and upstream resolver integration. The implementation demonstrates advanced protocol engineering patterns and distributed systems principles.

## Protocol Architecture

### Core Implementation Components
- **UDP Transport Layer**: Non-blocking datagram processing with concurrent goroutine dispatch
- **DNS Protocol Layer**: RFC1035-compliant message encoding/decoding with compression support
- **Resource Record Handler**: Polymorphic RR type processing with binary-safe implementations
- **Query Propagation**: Asynchronous upstream resolver integration with timeout semantics

### Protocol Specifications
```go
type DNSHeader struct {
    ID      uint16 // Transaction identifier
    Flags   uint16 // Control flags (QR|Opcode|AA|TC|RD|RA|Z|RCODE)
    QDCount uint16 // Question section cardinality
    ANCount uint16 // Answer section RR count
    NSCount uint16 // Authority section RR count
    ARCount uint16 // Additional section RR count
}

type DNSQuestion struct {
    Name  string  // Domain name sequence
    Type  uint16  // RR type identifier
    Class uint16  // Class identifier
}

type DNSRecord struct {
    Name     string  // Domain name sequence
    Type     uint16  // RR type identifier
    Class    uint16  // Class identifier
    TTL      uint32  // Time-to-live
    RDLength uint16  // RDATA length
    RData    []byte  // Resource data
}
```

## Implementation Architecture

### Protocol Processing Pipeline

#### Message Parser Implementation
- Binary-safe buffer management for DNS wire format
- Domain name label compression/decompression with pointer traversal
- Resource record serialization with length-prefixed encoding
- Transaction ID correlation for asynchronous responses

#### Network Stack Integration
- UDP socket multiplexing with Go runtime scheduler
- Concurrent query handling via goroutine dispatch
- Configurable upstream resolver interface
- Structured error propagation and logging

### Binary Wire Format Specification

#### DNS Header Structure (96 bits)
| Field      | Bit Offset | Length | Semantic Definition |
|------------|------------|--------|-------------------|
| ID         | 0          | 16     | Query identifier for transaction correlation |
| Flags      | 16         | 16     | Protocol control bits |
| QDCOUNT    | 32         | 16     | Question section cardinality |
| ANCOUNT    | 48         | 16     | Answer section RR count |
| NSCOUNT    | 64         | 16     | Authority section RR count |
| ARCOUNT    | 80         | 16     | Additional section RR count |

#### Message Compression Algorithm
- Label pointer detection with 2-bit discrimination
- Offset-based compression with 14-bit pointer space
- Recursive decompression with cycle detection
- Length-prefixed label encoding

## Implementation Capabilities

### Protocol Features
- RFC1035-compliant message processing
- Concurrent query handling
- Upstream resolver integration
- Error propagation and recovery
- Domain name compression support

### Technical Architecture
- Non-blocking I/O operations
- Concurrent goroutine dispatch
- Binary-safe buffer handling
- Resource lifecycle management

## Deployment Configuration

### Prerequisites
- Go runtime environment (≥1.18)
- Privileged port binding capabilities
- Network stack access permissions

### Binary Compilation
```sh
git clone https://github.com/sumitbhuia/GoDNS.git
cd GoDNS
go build -o godns main.go
```

### Process Execution
```sh
sudo ./godns  # Requires privileged port binding
```

Default configuration establishes UDP listener on port 53 with upstream resolver at 8.8.8.8:53.

## Protocol Enhancement Specifications

### Planned Implementation Extensions
- TCP fallback for truncated responses
- EDNS0 (RFC6891) implementation
- Response cache with LRU eviction
- Extended RR type support
- Security protocol integration

## RFC Specifications

### Primary Protocol Documentation
- [RFC 1035: DNS Implementation and Specification](https://datatracker.ietf.org/doc/html/rfc1035)
- [RFC 6891: Extension Mechanisms for DNS (EDNS(0))](https://datatracker.ietf.org/doc/html/rfc6891)

## Technical Proficiencies Demonstrated
- Protocol Engineering
- Distributed Systems Architecture
- Concurrent Programming Patterns
- Binary Protocol Implementation
- Network Stack Integration
- Resource Management
- Error Handling Methodologies

## Implementation Contact Vector

📧 **Electronic Mail:** sumitbhuia100@gmail.com  
🐙 **Version Control:** [sumitbhuia](https://github.com/sumitbhuia)  
💬 **Communication Channel:** [@bhuia_sumit](https://twitter.com/bhuia_sumit)

---

⭐ **Repository attribution appreciated** ⭐