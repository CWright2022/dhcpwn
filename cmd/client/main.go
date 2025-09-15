package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	serverIP   = "127.0.0.1" // change to server IP if remote
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

func main() {

	messageCounter := 0

	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP(serverIP),
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
		ServerAddress:  serverIP,
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
			log.Print(response)
		case <-stop:
			log.Println("client: received shutdown signal, exiting")
			return
		}
	}
}
