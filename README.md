# GoDNS - A Lightweight Custom DNS Resolver in Go

## Overview
GoDNS is a high-performance DNS resolver written in Go, designed for efficiency and scalability. It provides fast domain resolution while maintaining strict compliance with the DNS protocol. By leveraging Go's concurrency model, it ensures low-latency query resolution, making it suitable for personal deployments.

## Features
- **Custom DNS Resolution**: Implements core DNS query handling without relying on third-party resolvers.  
- **Efficient Parsing and Encoding**: Optimized request parsing and response encoding to minimize processing overhead.  
- **Concurrency-Driven Performance**: Uses Goroutines to handle multiple requests concurrently.  
- **Logging and Debugging**: Provides detailed logs for each query, allowing easy debugging and monitoring.  
- **Configurable Query Handling**: Easily extendable for custom domain resolution logic.  

## Installation
### Prerequisites
Ensure you have the following installed:  
- **Go 1.18+**  
- **Git**  

### Clone the Repository
```sh
git clone https://github.com/sumitbhuia/GoDNS.git
cd GoDNS
```

### Build and Run
To build the resolver, run:
```sh
go build -o GoDNS main.go
```

To execute the resolver:
```sh
./GoDNS
```
Alternatively, use the provided script:
```sh
chmod +x run.sh
./run.sh
```

## Usage
The resolver listens for DNS queries on port 53 by default. You can configure custom settings inside `server.go`. Once running, use a tool like `dig` to test its functionality:
```sh
dig @localhost example.com
```

## Project Structure
```
├── dns
│   ├── message.go   # DNS message parsing and encoding
│   ├── server.go    # DNS resolver logic
│   ├── dns_test.go  # Unit tests for DNS handling
├── main.go          # Entry point for the resolver
├── run.sh           # Script to run the resolver
├── go.mod           # Go module file
└── README.md        # Documentation
```

## Testing
Unit tests are included to validate DNS query handling. Run the tests with:
```sh
go test ./dns
```
or inside the `dns` directory:
```sh
go test -v
```

## Pending Bug Fixes
- **Extra 9 Bytes in Response**: An additional 9 bytes appear at the end of the response, possibly due to an error in the parsing logic.  

## Future Enhancements
- **Caching Mechanism**: Implement query caching for faster resolution.  
- **Custom Records Support**: Allow user-defined static DNS records.  
- **Security Enhancements**: Add DNSSEC support and improved logging.  
