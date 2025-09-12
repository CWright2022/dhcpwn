package main

import (
	"fmt"
	"log"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

func main() {
	serverAddr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 68, // match custom server port
	}

	localAddr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 67, // custom client port
	}

	conn, err := net.ListenUDP("udp4", &localAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer conn.Close()

	// Construct a simple DHCP Discover packet
	pkt, _ := dhcpv4.New(dhcpv4.WithTransactionID(dhcpv4.TransactionID{0x12, 0x34, 0x56, 0x78}))
	pkt.Options.Update(dhcpv4.OptMessageType(dhcpv4.MessageTypeDiscover))
	pkt.Options.Update(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, []byte("Hello from client")))

	// Send to server
	_, err = conn.WriteToUDP(pkt.ToBytes(), &serverAddr)
	if err != nil {
		log.Fatalf("failed to send: %v", err)
	}

	fmt.Println("Sent DHCP Discover to server")

	// Wait for reply
	buf := make([]byte, 1500)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Fatalf("error reading: %v", err)
	}

	reply, _ := dhcpv4.FromBytes(buf[:n])
	vendorOpt := reply.Options.Get(dhcpv4.OptionVendorSpecificInformation)
	fmt.Printf("Received reply from server with vendor option: %v\n", vendorOpt)
}
