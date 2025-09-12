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
	clientTID  = "test-client"
)

func main() {

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
	response := doTransaction(serverAddr, clientTID)
	log.Print(response)

	// allow clean shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			log.Printf("doing transaction\n")
			response = doTransaction(serverAddr, clientTID)
			log.Print(response)
		case <-stop:
			log.Println("client: received shutdown signal, exiting")
			return
		}
	}
}
