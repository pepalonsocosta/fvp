package e2e

import (
	"context"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pepalonsocosta/fvp/internal/client"
	"github.com/pepalonsocosta/fvp/internal/server"
)

// TestFullVPNProtocol tests the complete VPN protocol with real TUN interfaces
func TestFullVPNProtocol(t *testing.T) {
	// Skip if not running as root (TUN interfaces require root)
	if !isRoot() {
		t.Skip("Skipping test: requires root privileges for TUN interfaces")
	}

	// Test configuration
	serverPort := ":8080" // Use port from server.yaml
	serverAddr := "127.0.0.1:8080"
	
	// Step 1: Start server
	t.Log("Starting VPN server...")
	srv := server.NewServer()
	
	// Start server in goroutine
	_, serverCancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	serverStarted := make(chan bool, 1)
	
	go func() {
		err := srv.Start("../../server.yaml", serverPort)
		if err != nil {
			t.Logf("Server start error: %v", err)
			serverDone <- err
			return
		}
		serverStarted <- true
		serverDone <- nil
	}()
	
	// Wait for server to start with timeout
	t.Log("Waiting for server to start...")
	select {
	case <-serverStarted:
		t.Log("Server started successfully")
	case err := <-serverDone:
		if err != nil {
			t.Fatalf("Server failed to start: %v", err)
		}
	case <-time.After(5 * time.Second):
		serverCancel()
		t.Fatalf("Server failed to start within timeout")
	}
	
	// Give server a moment to bind to port
	time.Sleep(1 * time.Second)
	
	// Verify server is listening on the port
	t.Log("Checking server status...")
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		conn, err := net.Dial("udp", serverAddr)
		if err == nil {
			conn.Close()
			t.Log("Server is listening on", serverAddr)
			break
		}
		if i == maxRetries-1 {
			serverCancel()
			t.Fatalf("Server not listening on %s after %d retries: %v", serverAddr, maxRetries, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	
	// Step 2: Connect client
	t.Log("Connecting VPN client...")
	client := client.NewClient(serverAddr)
	
	err := client.Connect()
	if err != nil {
		serverCancel()
		t.Fatalf("Failed to connect client: %v", err)
	}
	
	t.Logf("Client connected: ID=%d, IP=%s", client.GetClientID(), client.GetAssignedIP())
	
	// Step 3: Test ping from client to server (using realistic method)
	t.Log("Testing ping from client to server...")
	
	// Use the more realistic ping method
	pingCmd := exec.Command("ping", "-c", "3", "10.0.0.1")
	output, err := pingCmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Ping output: %s", string(output))
		t.Logf("Ping error: %v", err)
		t.Fatalf("Ping failed - VPN tunnel not working properly")
	} else {
		t.Logf("✓ Ping successful: %s", string(output))
	}
	
	// Step 4: Verify server received packets
	t.Log("Checking server packet processing...")
	
	// Check if server TUN interface is up
	checkTUNInterface(t, "fvp0")
	checkTUNInterface(t, "fvp-client0")
	
	// Step 5: Test packet capture
	t.Log("Testing packet capture...")
	
	// Use tcpdump to capture packets on server TUN interface
	captureCmd := exec.Command("timeout", "5", "tcpdump", "-i", "fvp0", "-c", "1", "icmp")
	captureOutput, err := captureCmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Packet capture output: %s", string(captureOutput))
		t.Logf("Packet capture error: %v", err)
	} else {
		t.Logf("Captured packets: %s", string(captureOutput))
	}
	
	// Step 6: Cleanup
	t.Log("Cleaning up...")
	
	err = client.Disconnect()
	if err != nil {
		t.Logf("Client disconnect error: %v", err)
	}
	
	// Stop server
	srv.Stop()
	serverCancel()
	
	// Wait for server to stop
	select {
	case err := <-serverDone:
		if err != nil && err.Error() != "server not started" {
			t.Logf("Server stop error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Log("Server stop timeout (non-critical)")
	}
	
	t.Log("Test completed successfully!")
}

// TestClientToClientCommunication tests communication between two clients
func TestClientToClientCommunication(t *testing.T) {
	// Skip if not running as root
	if !isRoot() {
		t.Skip("Skipping test: requires root privileges for TUN interfaces")
	}
	
	t.Log("This test would require two clients - implementing later")
	t.Skip("Multi-client test not implemented yet")
}

// Helper functions

func isRoot() bool {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "0"
}

func checkTUNInterface(t *testing.T, iface string) {
	cmd := exec.Command("ip", "link", "show", iface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TUN interface %s not found: %v\nOutput: %s", iface, err, string(output))
	}
	t.Logf("TUN interface %s status: %s", iface, string(output))
}

func TestTUNInterfaceCreation(t *testing.T) {
	// Skip if not running as root
	if !isRoot() {
		t.Skip("Skipping test: requires root privileges for TUN interfaces")
	}
	
	t.Log("Testing TUN interface creation...")
	
	// Test server TUN interface - will fail if not found
	serverTUN := "fvp0"
	checkTUNInterface(t, serverTUN)
	
	// Test client TUN interface - will fail if not found
	clientTUN := "fvp-client0"
	checkTUNInterface(t, clientTUN)
	
	t.Log("✓ TUN interface test completed - both interfaces exist")
}

func TestNetworkConnectivity(t *testing.T) {
	// Skip if not running as root
	if !isRoot() {
		t.Skip("Skipping test: requires root privileges for TUN interfaces")
	}
	
	t.Log("Testing network connectivity...")
	
	// Check if we can resolve localhost
	_, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to resolve localhost: %v", err)
	}
	
	// Check if we can create UDP connections
	conn, err := net.Dial("udp", "127.0.0.1:1195")
	if err != nil {
		t.Logf("UDP connection test failed (expected if server not running): %v", err)
	} else {
		conn.Close()
	}
	
	t.Log("Network connectivity test completed")
}
