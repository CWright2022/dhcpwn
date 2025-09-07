package main

import (
	"log"
	"net"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func sendMessage(iface net.Interface, myIPAddr net.IP, dstIPAddr net.IP, message string) {
	handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()
	srcMAC := iface.HardwareAddr
	dstMAC, _ := net.ParseMAC("00:00:00:38:e2:21")

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
		SrcIP:    myIPAddr,
		DstIP:    dstIPAddr,
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
		YourClientIP: myIPAddr,
		Options: layers.DHCPOptions{
			layers.NewDHCPOption(
				layers.DHCPOptMessageType,
				[]byte{byte(layers.DHCPMsgTypeRequest)},
			),
			layers.NewDHCPOption(
                layers.DHCPOptParamsRequest,
                []byte{
                    byte(layers.DHCPOptSubnetMask),
                    byte(layers.DHCPOptRouter),
                    byte(layers.DHCPOptDNS),
                },
            ),
			layers.NewDHCPOption(
				layers.DHCPOptVendorOption,
				[]byte(message),
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

	log.Printf("Message sent: %s\n", message)
}