package main

import (
    "fmt"
    "net"
    "os"
)

func main() {
    addr := net.UDPAddr{
        Port: 677,
        IP:   net.ParseIP("0.0.0.0"),
    }
    conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error binding to UDP port: %v\n", err)
        os.Exit(1)
    }
    defer conn.Close()
    fmt.Println("Listening on UDP port 677")

    buf := make([]byte, 2048)
    for {
        n, remoteAddr, err := conn.ReadFromUDP(buf)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error reading UDP packet: %v\n", err)
            continue
        }
        fmt.Printf("Received from %v: %s\n", remoteAddr, string(buf[:n]))
    }
}