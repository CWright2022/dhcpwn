<<<<<<< HEAD
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
		Port: 1067, // match custom server port
	}

	localAddr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 1068, // custom client port
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
=======
package main

import (
	"fmt"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func ListenMessages(ifaceName string, localIP net.IP) {
	// If local IP is 127.0.0.1, use UDP socket
	if localIP.IsLoopback() {
		addr := &net.UDPAddr{
			IP:   localIP,
			Port: 677, // server port used in SendMessage
		}

		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			log.Fatalf("Failed to bind UDP port: %v", err)
		}
		defer conn.Close()

		log.Printf("[UDP] Listening on %s:677\n", localIP)

		buf := make([]byte, 4096)
		for {
			n, src, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Read error: %v", err)
				continue
			}
			log.Printf("[UDP] Message from %s: %s\n", src, string(buf[:n]))
		}
		return
	}

	// Otherwise, use raw packet capture
	handle, err := pcap.OpenLive(ifaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Filter only DHCP-style traffic (UDP 677 in this example)
	if err := handle.SetBPFFilter("udp port 677"); err != nil {
		log.Fatal(err)
	}

	log.Printf("[RAW] Listening on interface %s\n", ifaceName)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4); dhcpLayer != nil {
			dhcp := dhcpLayer.(*layers.DHCPv4)
			fmt.Printf("[RAW] DHCP packet: Xid=0x%x, YourIP=%s\n", dhcp.Xid, dhcp.YourClientIP)
			for _, opt := range dhcp.Options {
				fmt.Printf("  Option %d: %v\n", opt.Type, opt.Data)
			}
		} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp := udpLayer.(*layers.UDP)
			appPayload := udp.Payload
			fmt.Printf("[RAW UDP] Payload from %s: %s\n", packet.NetworkLayer().NetworkFlow().Src().String(), string(appPayload))
		}
	}
}

func main() {
	// Example usage
	ifaceName := "ens33"             // real NIC for raw capture
	localIP := net.ParseIP("192.168.254.129") // change to test IP

	ListenMessages(ifaceName, localIP)
}
>>>>>>> fc1202bd83ea870aa9d6930bd343fa8df2cb00ec
