// File: main.go
package main

import (
	"log"

	"github.com/sumitbhuia/GoDNS/dns"
)

func main() {
	// Listen on standard DNS port, forward to Google's DNS
	server := dns.NewDNSServer(":53", "8.8.8.8:53")

	log.Println("Starting DNS server...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("DNS server error: %v", err)
	}
}
