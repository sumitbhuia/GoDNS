# GoDNS - A Lightweight DNS Server in Go

## Overview
GoDNS is a high-performance, customizable DNS server written in Golang. It supports recursive DNS resolution, caching, and various DNS record types. Additionally, it includes security features like DNS over HTTPS (DoH) and DNS over TLS (DoT) for encrypted queries.

## Features
- **Recursive DNS Resolution** – Resolves domain names by querying root, TLD, and authoritative servers.
- **Caching** – Stores resolved queries to improve performance.
- **Support for Multiple DNS Record Types**:
  - A (IPv4 Address)
  - AAAA (IPv6 Address)
  - CNAME (Canonical Name)
  - MX (Mail Exchanger)
  - NS (Name Server)
  - TXT (Text Records)
  - PTR (Reverse DNS Lookup)
- **Security Enhancements**:
  - DNS over HTTPS (DoH)
  - DNS over TLS (DoT)
  - DNSSEC (Domain Name System Security Extensions)
- **Custom Configuration** – Define custom domain mappings.

## Installation

### Prerequisites
- Golang installed (Go 1.18+ recommended)

### Clone the Repository
```sh
git clone https://github.com/yourusername/GoDNS.git
cd GoDNS
```

### Install Dependencies
```sh
go mod tidy
```

## Usage

### Start the DNS Server
```sh
go run main.go
```
By default, the server listens on **UDP port 53**.

### Configure Custom DNS Records
Modify `config.json` to define custom mappings.

Example:
```json
{
  "A": {
    "example.com": "192.168.1.1"
  },
  "CNAME": {
    "www.example.com": "example.com"
  },
  "MX": {
    "example.com": "mail.example.com"
  }
}
```

## API Endpoints (For DoH Support)
| Method | Endpoint | Description |
|--------|---------|-------------|
| GET | `/dns-query` | Resolves a domain over HTTPS |
| POST | `/resolve` | Performs a DNS lookup via JSON request |

## Learning
During the development of GoDNS, the following key concepts and technologies were explored:
- **DNS Protocol** – Understanding how domain name resolution works, including recursive and iterative queries.
- **Go Networking** – Utilizing Go's `net` package and `miekg/dns` library for DNS query handling.
- **Caching Mechanisms** – Implementing in-memory caching strategies to improve response times.
- **Security in DNS** – Exploring DNSSEC, DoH, and DoT for encrypted and secure DNS resolution.
- **Configuration Management** – Creating JSON-based configurations for easy customization of DNS records.
- **Optimizing Performance** – Reducing latency with efficient query handling and caching strategies.

## Roadmap
- Implement full DNSSEC validation.
- Add Web UI for managing DNS records.
- Enhance caching with TTL-based eviction policies.

## Contributing
Pull requests are welcome! Please open an issue first to discuss proposed changes.

## License
This project is licensed under the MIT License.

## Acknowledgments
- Inspired by existing DNS resolver implementations.
- Uses `miekg/dns` for handling DNS queries efficiently.

