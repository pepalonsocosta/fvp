package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pepalonsocosta/fvp/internal/protocol"
	"github.com/pepalonsocosta/fvp/internal/server"
)

// version is injected at build time via -ldflags "-X main.version=VERSION"
// Example: go build -ldflags "-X main.version=1.2.3" -o fvps ./cmd/server
var version string

func main() {
	// Initialize protocol version from app version
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
	case "setup":
		handleSetup()
	case "up":
		handleUp()
	case "status":
		handleStatus()
	case "add-client":
		handleAddClient()
	case "list-clients":
		handleListClients()
	case "remove-client":
		handleRemoveClient()
	case "version":
		showVersion()
	case "help":
		showUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

func handleSetup() {
	flags := flag.NewFlagSet("setup", flag.ExitOnError)
	port := flags.String("port", "", "UDP port to listen on (required)")
	timeout := flags.Int("timeout", 0, "Client timeout in minutes (required)")
	
	flags.Parse(os.Args[2:])

	if *port == "" || *timeout == 0 {
		fmt.Println("Error: --port and --timeout are required")
		fmt.Println("Usage: fvps setup --port <port> --timeout <minutes>")
		os.Exit(1)
	}

	cliSrv := NewCLIServer()
	
	err := cliSrv.Setup(*port, *timeout)
	if err != nil {
		fmt.Printf("Setup failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration created: server.yaml\n")
	fmt.Printf("Server will listen on port %s\n", *port)
	fmt.Printf("Client timeout: %d minutes\n", *timeout)
	fmt.Println("Run 'fvps up' to start the server")
}

func handleUp() {
	cliSrv := NewCLIServer()
	
	setupSignalHandling(cliSrv.server)
	
	err := cliSrv.server.LoadConfig("server.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	
	port := cliSrv.server.GetPort()
	if port == "" {
		port = ":1194" // Default port
	}
	
	err = cliSrv.server.Start("server.yaml", port)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
	
	<-make(chan struct{})
}

func handleStatus() {
	cliSrv := NewCLIServer()
	
	err := cliSrv.Status()
	if err != nil {
		fmt.Printf("Failed to get server status: %v\n", err)
		os.Exit(1)
	}
}

func handleAddClient() {
	cliSrv := NewCLIServer()
	
	clientID, key, err := cliSrv.AddClient()
	if err != nil {
		fmt.Printf("Failed to add client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Client added successfully\n")
	fmt.Printf("Client ID: %d\n", clientID)
	fmt.Printf("Key: %s\n", key)
	fmt.Println("Add this key to your client configuration")
}

func handleListClients() {
	cliSrv := NewCLIServer()
	
	clients, err := cliSrv.ListClientsRealtime()
	if err != nil {
		fmt.Printf("Failed to list clients: %v\n", err)
		os.Exit(1)
	}

	if len(clients) == 0 {
		fmt.Println("No clients configured")
		return
	}

	fmt.Println("Client Status:")
	fmt.Println("ID  IP         Status     Last Connection")
	for _, client := range clients {
		status := "Disconnected"
		if client.Connected {
			status = "Connected"
		}
		
		lastSeen := "Never"
		if !client.LastSeen.IsZero() {
			lastSeen = client.LastSeen.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-3d %-10s %-11s %s\n", client.ID, client.IP, status, lastSeen)
	}
}

func handleRemoveClient() {
	flags := flag.NewFlagSet("remove-client", flag.ExitOnError)
	clientID := flags.Int("id", 0, "Client ID to remove (required)")
	
	flags.Parse(os.Args[2:])

	if *clientID == 0 {
		fmt.Println("Error: --id is required")
		fmt.Println("Usage: fvps remove-client --id <client_id>")
		os.Exit(1)
	}

	cliSrv := NewCLIServer()
	
	err := cliSrv.RemoveClient(uint8(*clientID))
	if err != nil {
		fmt.Printf("Failed to remove client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Client %d removed successfully\n", *clientID)
}

func setupSignalHandling(srv *server.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived %v, shutting down gracefully...\n", sig)
		
		err := srv.Stop()
		if err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		}
		
		fmt.Println("Server stopped")
		os.Exit(0)
	}()
}

func showVersion() {
	if version == "" {
		fmt.Printf("FVP Server version unknown\n")
	} else {
		fmt.Printf("FVP Server version %s\n", version)
	}
}

func showUsage() {
	fmt.Println("FVP Server - Fast VPN Server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fvps <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  setup         Create initial server configuration")
	fmt.Println("  up            Start the VPN server")
	fmt.Println("  status        Show server status")
	fmt.Println("  add-client    Add a new client")
	fmt.Println("  list-clients  List all configured clients")
	fmt.Println("  remove-client Remove a client")
	fmt.Println("  version       Show version information")
	fmt.Println("  help          Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fvps setup --port 1194 --timeout 30")
	fmt.Println("  fvps up")
	fmt.Println("  fvps status")
	fmt.Println("  fvps add-client")
	fmt.Println("  fvps list-clients")
	fmt.Println("  fvps remove-client --id 1")
}
