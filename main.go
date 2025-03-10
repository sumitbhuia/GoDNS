// File: main.go
package main

import (
	"log"

	"github.com/sumitbhuia/GoDNS/dns"
)

func main() {
	// Original configuration (requires admin privileges and service conflicts):
	server := dns.NewDNSServer(":53", "8.8.8.8:53")

	// Windows-friendly configuration:
	// server := dns.NewDNSServer(":5353", "8.8.8.8:53")

	log.Println("Starting DNS server...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("DNS server error: %v", err)
	}
}
