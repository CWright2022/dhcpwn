package shared

import (
    "fmt"
    "net"
    "os/exec"
    "strings"
    "time"
)

// sendARPing sends a single ping to the IP to populate the ARP table.
func sendARPing(ip string) {
    // Use ping with 1 packet and short timeout
    _ = exec.Command("ping", "-c", "1", "-W", "1", ip).Run()
}

func GetNextMAC(ip string) (string, error) {
    sendARPing(ip) // Try to populate ARP table

    // Wait a moment for ARP table to update
    time.Sleep(200 * time.Millisecond)

    out, err := exec.Command("arp", "-n", ip).Output()
    if err != nil {
        return "", err
    }

    output := string(out)
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.Contains(line, ip) {
            fields := strings.Fields(line)
            // Try to find a valid MAC address in the fields
            for _, field := range fields {
                if hw, err := net.ParseMAC(field); err == nil && len(hw) == 6 {
                    return field, nil
                }
            }
        }
    }
    return "", fmt.Errorf("MAC not found for IP %s", ip)
}