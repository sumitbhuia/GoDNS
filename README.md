# GoDNS - A Lightweight Custom DNS Server in Go

## Overview
This project is a high-performance, fully functional DNS server written in Go. Designed with efficiency and scalability in mind, it provides fast domain resolution while maintaining strict compliance with the DNS protocol. By leveraging Go's concurrency model, the server ensures low-latency query resolution, making it suitable for both personal and enterprise-level deployments.

## Features
- **Custom DNS Resolution**: Implements core DNS query handling without relying on third-party resolvers.
- **Efficient Parsing and Encoding**: Optimized request parsing and response encoding to minimize processing overhead.
- **Concurrency-Driven Performance**: Uses Goroutines to handle multiple requests concurrently.
- **Logging and Debugging**: Provides detailed logs for each query, allowing easy debugging and monitoring.
- **Configurable Query Handling**: Easily extendable for custom domain resolution logic.

## Installation
### Prerequisites
Ensure you have the following installed on your system:
- **Go 1.18+**
- **Git**

### Clone the Repository
```sh
 git clone https://github.com/sumitbhuia/GoDNS.git
 cd GoDNS
```

### Build and Run
To build the server, run:
```sh
 go build -o GoDNS main.go
```

To execute the server:
```sh
 ./GoDNS
```
Alternatively, you can use the provided script:
```sh
 chmod +x run.sh
 ./run.sh
```

## Usage
The server listens for DNS queries on port 53 by default. You can configure custom settings inside `server.go`. Once running, use a tool like `dig` to test its functionality:
```sh
 dig @localhost example.com
```

## Project Structure
```
├── dns
│   ├── message.go   # DNS message parsing and encoding
│   ├── server.go    # DNS server logic
│   ├── dns_test.go  # Unit tests for DNS handling
├── main.go          # Entry point for the server
├── run.sh           # Script to run the server
├── go.mod           # Go module file
└── README.md        # Documentation
```

## Testing
Unit tests are included to validate DNS query handling. Run the tests with:
```sh
 go test ./dns

# INSIDE DNS DIRECTORY
 go test -v
```
## BUG FIX PENDING
- 9 Byte extra message size at the end . Possibly error in parsing logic.
- 
## Future Enhancements
- **Caching Mechanism**: Implement query caching for faster resolution.
- **Custom Records Support**: Allow user-defined static DNS records.
- **Security Enhancements**: Add DNSSEC support and enhanced logging.

