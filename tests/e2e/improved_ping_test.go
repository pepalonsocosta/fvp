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

// TestImprovedPing tests the ping functionality with proper TUN configuration
func TestImprovedPing(t *testing.T) {
	// Skip if not running as root
	if !isRoot() {
		t.Skip("Skipping test: requires root privileges for TUN interfaces")
	}

	// Test configuration
	serverPort := ":8080"
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
			t.Log("✓ Server is listening on", serverAddr)
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
	
	t.Logf("✓ Client connected: ID=%d, IP=%s", client.GetClientID(), client.GetAssignedIP())
	
	// Step 3: Verify TUN interface configuration
	t.Log("Verifying TUN interface configuration...")
	
	// Check server TUN interface
	checkTUNInterface(t, "fvp0")
	serverIP := getTUNInterfaceIP(t, "fvp0")
	if serverIP != "" {
		t.Logf("✓ Server TUN IP: %s", serverIP)
	}
	
	// Check client TUN interface
	checkTUNInterface(t, "fvp-client0")
	clientIP := getTUNInterfaceIP(t, "fvp-client0")
	if clientIP != "" {
		t.Logf("✓ Client TUN IP: %s", clientIP)
	}
	
	// Step 4: Test ping with proper routing
	t.Log("Testing ping with proper routing...")
	
	// Add routing rules for ping to work
	setupRouting(t)
	
	// Test ping from client to server
	// Use the realistic ping method (without -I flag)
	pingCmd := exec.Command("ping", "-c", "3", "-W", "2", "10.0.0.1")
	output, err := pingCmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Ping output: %s", string(output))
		t.Logf("Ping error: %v", err)
		t.Fatalf("Ping failed - VPN tunnel not working properly")
	} else {
		t.Logf("✓ Ping successful: %s", string(output))
	}
	
	// Step 5: Test packet capture
	t.Log("Testing packet capture...")
	
	// Use tcpdump to capture packets on server TUN interface
	captureCmd := exec.Command("timeout", "3", "tcpdump", "-i", "fvp0", "-c", "1", "icmp")
	captureOutput, err := captureCmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Packet capture output: %s", string(captureOutput))
		t.Logf("Packet capture error: %v", err)
	} else {
		t.Logf("✓ Captured packets: %s", string(captureOutput))
	}
	
	// Step 6: Cleanup
	t.Log("Cleaning up...")
	
	// Clean up routing
	cleanupRouting(t)
	
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
	
	t.Log("✓ Test completed successfully!")
}

// Helper functions

func setupRouting(t *testing.T) {
	t.Log("Setting up routing for ping test...")
	
	// Add route from client TUN to server TUN
	routeCmd := exec.Command("ip", "route", "add", "10.0.0.1/32", "dev", "fvp-client0")
	err := routeCmd.Run()
	if err != nil {
		t.Logf("Failed to add route: %v", err)
	} else {
		t.Log("✓ Added route for ping")
	}
	
	// Enable IP forwarding
	forwardCmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	err = forwardCmd.Run()
	if err != nil {
		t.Logf("Failed to enable IP forwarding: %v", err)
	} else {
		t.Log("✓ Enabled IP forwarding")
	}
}

func cleanupRouting(t *testing.T) {
	t.Log("Cleaning up routing...")
	
	// Remove route
	exec.Command("ip", "route", "del", "10.0.0.1/32", "dev", "fvp-client0").Run()
	
	// Disable IP forwarding
	exec.Command("sysctl", "-w", "net.ipv4.ip_forward=0").Run()
}

func getTUNInterfaceIP(t *testing.T, iface string) string {
	cmd := exec.Command("ip", "addr", "show", iface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to get IP for %s: %v", iface, err)
		return ""
	}
	
	// Parse IP from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := strings.Split(parts[1], "/")[0]
				return ip
			}
		}
	}
	
	return ""
}
