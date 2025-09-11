package shared

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetNextHopMac(ip string) (string, error) {
	out, err := exec.Command("arp", "-n", ip).Output()
	if err != nil {
		return "", err
	}

	output := string(out)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ip) {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[2], nil // MAC address usually is the 3rd field
			}
		}
	}
	return "", fmt.Errorf("MAC not found")
}
