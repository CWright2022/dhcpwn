package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

func sendToServer(data []byte) []byte {
	req, err := http.NewRequest("POST", "http://localhost:8000/checkin", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("error creating request: %v", err)
		return []byte("error creating request")
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error sending request: %v", err)
		return []byte("error sending request")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body
}
func main() {
	i := 0
	addr := net.UDPAddr{
		Port: 68, // custom server port
		IP:   net.IPv4zero,
	}

	conn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer conn.Close()
	fmt.Println("Server listening on port " + strconv.Itoa(addr.Port))

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
		fmt.Printf("Received DHCP packet from %v with vendor option: %v\n", clientAddr, string(vendorOpt))

		//TODO:
		//
		response := sendToServer(vendorOpt)
		// Send a reply
		reply, _ := dhcpv4.NewReplyFromRequest(pkt)
		reply.Options.Update(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
		reply.Options.Update(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, response))
		i++
		conn.WriteToUDP(reply.ToBytes(), clientAddr)
	}
}
