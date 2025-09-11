package main

import (
	"fmt"
	"log"
	"net"

	"github.com/cwright2022/dhcpwn/internal/shared"
)

func main() {
	// serverIP := net.IP{192, 168, 254, 130} //use for VM-to-VM testing
	serverIP := net.IP{172, 21, 95, 26} //use for WSL testing
	iface, ipAddr, _ := shared.GetActiveInterface()
	nextMAC, err := shared.GetNextHopMac(serverIP.String())
	fmt.Println("Next hop MAC: ", nextMAC)
	if err != nil {
		log.Fatal(err)
	}
	shared.SendMessage(*iface, ipAddr, serverIP, "pwned")
}
