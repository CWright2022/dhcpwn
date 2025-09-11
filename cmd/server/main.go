package main

import (
	"fmt"
	"log"
	"net"

	"github.com/cwright2022/dhcpwn/internal/shared"
)

func main() {
	serverIP := net.IP{192, 168, 254, 130}
	iface, ipAddr, err := shared.GetActiveInterface()
	nextMAC, err := shared.GetMAC(serverIP.String())
	fmt.Println("Next hop MAC: ", nextMAC)
	if err != nil {
		log.Fatal(err)
	}
	shared.SendMessage(*iface, ipAddr, serverIP, "pwned")
}
