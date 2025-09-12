package main

import (
    "fmt"
    "log"
    "net"

    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
)

func main() {
    // DHCP client listens on UDP 68
    addr := &net.UDPAddr{
        IP:   net.IPv4zero,
        Port: 667,
    }

    conn, err := net.ListenUDP("udp4", addr)
    if err != nil {
        log.Fatalf("failed to bind UDP socket: %v", err)
    }
    defer conn.Close()

    log.Println("DHCP client listening on port 667...")

    buf := make([]byte, 1500) // enough for a full Ethernet frame
    for {
        n, src, err := conn.ReadFromUDP(buf)
        if err != nil {
            log.Printf("read error: %v", err)
            continue
        }

        log.Printf("Received %d bytes from %s", n, src)

        // Decode packet
        packet := gopacket.NewPacket(buf[:n], layers.LayerTypeDHCPv4, gopacket.Default)
        if dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4); dhcpLayer != nil {
            dhcp, _ := dhcpLayer.(*layers.DHCPv4)

            fmt.Printf("DHCP Message from %s:\n", src)
            fmt.Printf("  Xid: 0x%x\n", dhcp.Xid)
            fmt.Printf("  ClientHWAddr: %s\n", dhcp.ClientHWAddr)
            fmt.Printf("  YourIP: %s\n", dhcp.YourClientIP)
            fmt.Printf("  ServerIP: %s\n", dhcp.ServerName)

            // Dump DHCP options
            for _, opt := range dhcp.Options {
                fmt.Printf("  Option %d: %v\n", opt.Type, opt.Data)
            }
        } else {
            log.Println("Not a DHCP packet")
        }
    }
}
