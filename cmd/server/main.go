package main

import (
	// "fmt"
	"log"
	"net"

	"github.com/cwright2022/dhcpwn/internal/shared"
)

func main() {
	clientIP := net.IP{192, 168, 254, 130}
	// clientIP := net.IP{192, 168, 254, 129}
	iface, ipAddr, err := shared.GetActiveInterface()
	if err != nil {
		log.Fatal(err)
	}
	shared.SendMessage(*iface, ipAddr, clientIP, "pwned")
}
