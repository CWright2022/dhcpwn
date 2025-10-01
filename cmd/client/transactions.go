package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

func getMyIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// Skip down or loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// We want IPv4, not loopback
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip != nil {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IPv4 address found")
}

func buildRequest(payload Payload) *dhcpv4.DHCPv4 {
	tidBytes := []byte(payload.ClientID)
	pkt, _ := dhcpv4.New(dhcpv4.WithTransactionID(dhcpv4.TransactionID(tidBytes)))
	pkt.Options.Update(dhcpv4.OptMessageType(dhcpv4.MessageTypeRequest))
	msg, err := json.Marshal(payload)
	if err != nil {
		msg = []byte("error marshalling json")
		log.Print(string(msg))
	}
	pkt.Options.Update(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, []byte(msg)))
	return pkt
}

// single transaction: create ephemeral UDP socket, send, wait for reply, close
func doTransaction(payload Payload) string {
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP(*payload.BrokerIP),
		Port: serverPort,
	}
	// DialUDP with nil local addr -> OS picks ephemeral local port
	conn, err := net.DialUDP("udp4", nil, serverAddr)
	if err != nil {
		log.Printf("client: DialUDP failed: %v", err)
		return "transaction-failed"
	}
	// ensure close quickly after transaction
	defer conn.Close()

	// Build DHCP-like request
	req := buildRequest(payload)
	out := req.ToBytes()

	// Send request
	_, err = conn.Write(out)
	if err != nil {
		log.Printf("client: write failed: %v", err)
		return "transaction-failed"
	}
	// log.Printf("client: sent request (txn 0x%08x) from %s -> %s", req.TransactionID, conn.LocalAddr(), conn.RemoteAddr())

	// Set read deadline so we don't keep socket open forever
	_ = conn.SetReadDeadline(time.Now().Add(readTO))

	// Buffer and read
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	if err != nil {
		// timeout or other error; log and return (socket will be closed via defer)
		log.Printf("client: read error (likely timeout): %v", err)
		return "transaction-failed"
	}

	reply, err := dhcpv4.FromBytes(buf[:n])
	if err != nil {
		log.Printf("client: failed to parse reply: %v", err)
		return "transaction-failed"
	}
	log.Printf("SENDING: %v", payload)
	// doTransaction(payload)

	vendor := reply.Options.Get(dhcpv4.OptionVendorSpecificInformation)
	log.Printf("RECEIVED: %s", string(vendor))
	return (string(vendor))
}
