package main

import (
	"log"
	"net"
	"fmt"
	"github.com/cwright2022/dhcpwn/pkg/getActiveInterface"
    "github.com/cwright2022/dhcpwn/pkg/getMAC"
    "github.com/cwright2022/dhcpwn/pkg/sendMessage"
)

func main() {
	serverIP := net.IP{192, 168, 254, 130}
	iface, ipAddr, err := getActiveInterface()
	nextMAC, err := getMAC(serverIP.String())
	fmt.Println("Next hop MAC: ", nextMAC)
	if err != nil {
		log.Fatal(err)
	}
	sendMessage(*iface, ipAddr, serverIP, "pwned")
}
