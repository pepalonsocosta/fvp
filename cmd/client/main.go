package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pepalonsocosta/fvp/internal/client"
	"github.com/pepalonsocosta/fvp/internal/config"
	"github.com/pepalonsocosta/fvp/internal/protocol"
)

var version string

func main() {
	if err := protocol.InitProtocolVersion(version); err != nil {
		fmt.Printf("Warning: Failed to initialize protocol version: %v\n", err)
		fmt.Println("Using default protocol version 1.0.0")
	}
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "connect":
		handleConnect()
	case "disconnect":
		handleDisconnect()
	case "status":
		handleStatus()
	case "version":
		handleVersion()
	case "help":
		showUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

func handleConnect() {
	// Load client config
	clientMgr, err := config.NewClientManager()
	if err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// Check if config exists and is valid
	cfg, err := clientMgr.Load()
	if err != nil || cfg.ClientID == 0 || cfg.Key == "" || cfg.Server == "" {
		// Config missing or incomplete, create template
		err = clientMgr.CreateTemplate()
		if err != nil {
			fmt.Printf("Failed to create config template: %v\n", err)
			os.Exit(1)
		}

		configPath := clientMgr.GetPath()
		fmt.Printf("Config file created at %s\n", configPath)
		fmt.Println("Fill in client_id, key, and server, then run 'fvpc connect' again")
		os.Exit(0)
	}

	c, err := client.NewClient(cfg.Server, cfg.ClientID, cfg.Key)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	err = c.Connect()
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connected to VPN server at %s\n", cfg.Server)
	fmt.Printf("Client ID: %d\n", c.GetClientID())
	fmt.Printf("Assigned IP: %s\n", c.GetAssignedIP())
	fmt.Println("Press Ctrl+C to disconnect")

	<-sigChan

	err = c.Disconnect()
	if err != nil {
		fmt.Printf("Error during disconnect: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Disconnected from VPN server")
}

func handleDisconnect() {
	fmt.Println("Disconnect command not implemented yet")
	fmt.Println("Use Ctrl+C while connected to disconnect")
}

func handleStatus() {
	fmt.Println("Status command not implemented yet")
	fmt.Println("This will show connection status and statistics")
}

func handleVersion() {
	if version == "" {
		fmt.Printf("FVP Client version unknown\n")
	} else {
		fmt.Printf("FVP Client version %s\n", version)
	}
}

func showUsage() {
	fmt.Println("FVP Client - Fast VPN Client")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  fvpc <command> [flags]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  connect     Connect to VPN server")
	fmt.Println("  disconnect  Disconnect from VPN server")
	fmt.Println("  status      Show connection status")
	fmt.Println("  version     Show version information")
	fmt.Println("  help        Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  fvpc connect")
	fmt.Println("  fvpc status")
	fmt.Println("  fvpc version")
	fmt.Println("")
	fmt.Println("Note: Create ~/.fvp/client.yaml with client_id, key, and server")
}
