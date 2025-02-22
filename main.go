// File: main.go
package main

import (
	"log"

	"github.com/sumitbhuia/GoDNS/dns"
)

func main() {
	server := dns.NewDNSServer(":53", "8.8.8.8:53")
	log.Println("Starting DNS server...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("DNS server error: %v", err)
	}
}
