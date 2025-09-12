<<<<<<< HEAD
package main

import (
	"fmt"
	"log"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

func main() {
	addr := net.UDPAddr{
		Port: 1067, // custom server port
		IP:   net.IPv4zero,
	}

	conn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer conn.Close()
	fmt.Println("Server listening on port 1067")

	buf := make([]byte, 1500)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("error reading: %v", err)
			continue
		}

		pkt, err := dhcpv4.FromBytes(buf[:n])
		if err != nil {
			log.Printf("failed to parse DHCP packet: %v", err)
			continue
		}

		vendorOpt := pkt.Options.Get(dhcpv4.OptionVendorSpecificInformation)
		fmt.Printf("Received DHCP packet from %v with vendor option: %v\n", clientAddr, vendorOpt)

		// Send a reply
		reply, _ := dhcpv4.NewReplyFromRequest(pkt)
		reply.Options.Update(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
		reply.Options.Update(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, []byte("Hello from client")))
		conn.WriteToUDP(reply.ToBytes(), clientAddr)
	}
}
=======
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
>>>>>>> fc1202bd83ea870aa9d6930bd343fa8df2cb00ec
