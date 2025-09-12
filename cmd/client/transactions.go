package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

func buildRequestTxn(tid string) *dhcpv4.DHCPv4 {
	tidBytes := []byte(tid)
	pkt, _ := dhcpv4.New(dhcpv4.WithTransactionID(dhcpv4.TransactionID(tidBytes)))
	pkt.Options.Update(dhcpv4.OptMessageType(dhcpv4.MessageTypeRequest))
	msg := fmt.Sprintf("hello from client %s", tid)
	pkt.Options.Update(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, []byte(msg)))
	return pkt
}

// single transaction: create ephemeral UDP socket, send, wait for reply, close
func doTransaction(serverAddr *net.UDPAddr, tid string) string {
	// DialUDP with nil local addr -> OS picks ephemeral local port
	conn, err := net.DialUDP("udp4", nil, serverAddr)
	if err != nil {
		log.Printf("client: DialUDP failed: %v", err)
		return "transaction-failed"
	}
	// ensure close quickly after transaction
	defer conn.Close()

	// Build DHCP-like request
	req := buildRequestTxn(tid)
	out := req.ToBytes()

	// Send request
	_, err = conn.Write(out)
	if err != nil {
		log.Printf("client: write failed: %v", err)
		return "transaction-failed"
	}
	log.Printf("client: sent request (txn 0x%08x) from %s -> %s", req.TransactionID, conn.LocalAddr(), conn.RemoteAddr())

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

	vendor := reply.Options.Get(dhcpv4.OptionVendorSpecificInformation)
	return (string(vendor))
}
