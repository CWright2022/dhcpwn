package main

import (
	"log"
	"net"
	"fmt"
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
