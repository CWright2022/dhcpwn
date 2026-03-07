package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

const (
	serverAddr = "http://192.168.10.168:8000/checkin" // change to server IP if remote
)

func sendToServer(data []byte, brokerID string) []byte {
	req, err := http.NewRequest("POST", serverAddr, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("error creating request: %v", err)
		return []byte("error creating request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Broker-ID", brokerID)
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
	addr := net.UDPAddr{
		Port: 69, // custom server port
		IP:   net.IPv4zero,
	}

	fmt.Print("sending test")
	sendToServer([]byte("test"), "notanid")

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "error fetching hostname"
	}
	brokerID := hostname

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

		//get data from vendor option
		vendorOpt := pkt.Options.Get(dhcpv4.OptionVendorSpecificInformation)
		fmt.Printf("Received DHCP packet from %v with vendor option: %v\n", clientAddr, string(vendorOpt))

		//forward to server
		response := sendToServer(vendorOpt, brokerID)
		fmt.Printf("Got server reply: %s", response)

		// Send a reply to client
		reply, _ := dhcpv4.NewReplyFromRequest(pkt)
		reply.Options.Update(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
		reply.Options.Update(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, response))
		conn.WriteToUDP(reply.ToBytes(), clientAddr)
	}
}
