package main

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	iface := "ens33"
	handle, err := pcap.OpenLive(iface, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	srcMAC, _ := net.ParseMAC("00:01:02:ab:cd:ef")
	dstMAC, _ := net.ParseMAC("00:15:5d:38:e2:21")

	ethernet := &layers.Ethernet{
		SrcMAC:       srcMAC,
		DstMAC:       dstMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    net.IP{192, 168, 254, 129},
		DstIP:    net.IP{192, 168, 254, 130},
	}

	udp := &layers.UDP{
		SrcPort: 67,
		DstPort: 68,
	}
	udp.SetNetworkLayerForChecksum(ip)

	dhcp := &layers.DHCPv4{
		Operation:    layers.DHCPOpRequest,
		HardwareType: layers.LinkTypeEthernet,
		Xid:          0xabcdeabc,
		Flags:        0x8000,
		ClientHWAddr: srcMAC,
		Options: layers.DHCPOptions{
			layers.NewDHCPOption(
				layers.DHCPOptVendorOption,
				[]byte("Hello World!"),
			),
		},
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err = gopacket.SerializeLayers(buffer, opts,
		ethernet, ip, udp, dhcp,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Send packet
	if err := handle.WritePacketData(buffer.Bytes()); err != nil {
		log.Fatal(err)
	}

	log.Println("DHCP Discover sent")
}
