
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gosnmp/gosnmp"
)

const (
	bgpPeerStateOID   = "1.3.6.1.4.1.9.2.3.5.1.1.7" // Example OID for BGP peer state
	bgpPrefixesOID    = "1.3.6.1.4.1.9.2.3.5.1.1.8"  // Example OID for received prefixes
)

func main() {
	
var community = flag.String("community", "public", "SNMP community string")
flag.Parse()
if len(os.Args) < 2 {

		log.Fatal("Usage: check_bgp_peers <snmp_target>")
	}

	target := os.Args[1]
	
	// Initialize SNMP
	gosnmp.Default.Target = target
	gosnmp.Default.Port = 161
	gosnmp.Default.Community = community
	gosnmp.Default.Version = gosnmp.Version2

	// Get BGP peer state
	state, err := getSNMPValue(bgpPeerStateOID)
	if err != nil {
		log.Fatalf("Error getting BGP peer state: %v", err)
	}

	// Get received prefixes
	prefixes, err := getSNMPValue(bgpPrefixesOID)
	if err != nil {
		log.Fatalf("Error getting BGP prefixes: %v", err)
	}

	// Check if peer is up (state = 1) and has prefixes
	if state != "1" || prefixes == "0" {
		fmt.Println("BGP peer down or no prefixes received")
		os.Exit(1)
	}

	fmt.Println("BGP peer is up with prefixes")
}

func getSNMPValue(oid string) (string, error) {
	result, err := gosnmp.Get([]string{oid})
	if err != nil {
		return "", err
	}

	if len(result.Variables) == 0 {
		return "0", nil
	}

	return result.Variables[0].Value.String(), nil
}
