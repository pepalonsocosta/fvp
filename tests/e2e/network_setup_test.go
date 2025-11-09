package e2e

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNetworkSetup tests the network configuration for VPN testing
func TestNetworkSetup(t *testing.T) {
	// Skip if not running as root
	if !isRoot() {
		t.Skip("Skipping test: requires root privileges for TUN interfaces")
	}

	t.Log("Testing network setup for VPN...")

	// Test 1: Check if TUN interfaces can be created
	t.Log("1. Testing TUN interface creation...")
	
	// Create test TUN interface
	createCmd := exec.Command("ip", "tuntap", "add", "dev", "test-tun", "mode", "tun")
	err := createCmd.Run()
	if err != nil {
		t.Logf("Failed to create test TUN: %v", err)
	} else {
		t.Log("✓ TUN interface creation works")
		
		// Clean up
		exec.Command("ip", "tuntap", "del", "dev", "test-tun", "mode", "tun").Run()
	}

	// Test 2: Check IP assignment
	t.Log("2. Testing IP assignment...")
	
	// Create temporary TUN for IP testing
	createCmd = exec.Command("ip", "tuntap", "add", "dev", "test-ip", "mode", "tun")
	err = createCmd.Run()
	if err != nil {
		t.Fatalf("Failed to create test TUN for IP testing: %v", err)
	}
	defer exec.Command("ip", "tuntap", "del", "dev", "test-ip", "mode", "tun").Run()
	
	// Bring up interface
	upCmd := exec.Command("ip", "link", "set", "test-ip", "up")
	err = upCmd.Run()
	if err != nil {
		t.Fatalf("Failed to bring up test TUN: %v", err)
	}
	
	// Assign IP
	ipCmd := exec.Command("ip", "addr", "add", "10.0.0.100/24", "dev", "test-ip")
	err = ipCmd.Run()
	if err != nil {
		t.Fatalf("Failed to assign IP to test TUN: %v", err)
	}
	
	// Verify IP assignment
	checkCmd := exec.Command("ip", "addr", "show", "test-ip")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to check IP assignment: %v", err)
	}
	
	if strings.Contains(string(output), "10.0.0.100") {
		t.Log("✓ IP assignment works")
	} else {
		t.Logf("IP assignment output: %s", string(output))
	}

	// Test 3: Check routing
	t.Log("3. Testing routing configuration...")
	
	// Add route
	routeCmd := exec.Command("ip", "route", "add", "10.0.0.0/24", "dev", "test-ip")
	err = routeCmd.Run()
	if err != nil {
		t.Logf("Failed to add route: %v", err)
	} else {
		t.Log("✓ Route addition works")
	}
	
	// Clean up route
	exec.Command("ip", "route", "del", "10.0.0.0/24", "dev", "test-ip").Run()

	// Test 4: Check ping capability
	t.Log("4. Testing ping capability...")
	
	// Try to ping the assigned IP
	pingCmd := exec.Command("ping", "-c", "1", "-W", "1", "10.0.0.100")
	err = pingCmd.Run()
	if err != nil {
		t.Logf("Ping test failed (expected): %v", err)
	} else {
		t.Log("✓ Ping works")
	}

	t.Log("Network setup test completed")
}

// TestTUNInterfaceConfiguration tests the specific TUN interface configuration
func TestTUNInterfaceConfiguration(t *testing.T) {
	// Skip if not running as root
	if !isRoot() {
		t.Skip("Skipping test: requires root privileges for TUN interfaces")
	}

	t.Log("Testing TUN interface configuration...")

	// Test server TUN configuration
	t.Log("Testing server TUN configuration...")
	
	// Create server TUN
	createCmd := exec.Command("ip", "tuntap", "add", "dev", "test-server", "mode", "tun")
	err := createCmd.Run()
	if err != nil {
		t.Fatalf("Failed to create server TUN: %v", err)
	}
	defer exec.Command("ip", "tuntap", "del", "dev", "test-server", "mode", "tun").Run()
	
	// Configure server TUN (10.0.0.1/24)
	upCmd := exec.Command("ip", "link", "set", "test-server", "up")
	err = upCmd.Run()
	if err != nil {
		t.Fatalf("Failed to bring up server TUN: %v", err)
	}
	
	ipCmd := exec.Command("ip", "addr", "add", "10.0.0.1/24", "dev", "test-server")
	err = ipCmd.Run()
	if err != nil {
		t.Fatalf("Failed to assign IP to server TUN: %v", err)
	}
	
	// Test client TUN configuration
	t.Log("Testing client TUN configuration...")
	
	// Create client TUN
	createCmd = exec.Command("ip", "tuntap", "add", "dev", "test-client", "mode", "tun")
	err = createCmd.Run()
	if err != nil {
		t.Fatalf("Failed to create client TUN: %v", err)
	}
	defer exec.Command("ip", "tuntap", "del", "dev", "test-client", "mode", "tun").Run()
	
	// Configure client TUN (10.0.0.2/24)
	upCmd = exec.Command("ip", "link", "set", "test-client", "up")
	err = upCmd.Run()
	if err != nil {
		t.Fatalf("Failed to bring up client TUN: %v", err)
	}
	
	ipCmd = exec.Command("ip", "addr", "add", "10.0.0.2/24", "dev", "test-client")
	err = ipCmd.Run()
	if err != nil {
		t.Fatalf("Failed to assign IP to client TUN: %v", err)
	}
	
	// Test ping between TUN interfaces
	t.Log("Testing ping between TUN interfaces...")
	
	// Add routing rules
	exec.Command("ip", "route", "add", "10.0.0.1/32", "dev", "test-client").Run()
	exec.Command("ip", "route", "add", "10.0.0.2/32", "dev", "test-server").Run()
	
	// Try ping from client to server
	pingCmd := exec.Command("ping", "-c", "1", "-W", "1", "-I", "test-client", "10.0.0.1")
	output, err := pingCmd.CombinedOutput()
	if err != nil {
		t.Logf("Ping failed (expected without proper routing): %v", err)
		t.Logf("Ping output: %s", string(output))
	} else {
		t.Log("✓ Ping between TUN interfaces works")
	}
	
	// Clean up routes
	exec.Command("ip", "route", "del", "10.0.0.1/32", "dev", "test-client").Run()
	exec.Command("ip", "route", "del", "10.0.0.2/32", "dev", "test-server").Run()

	t.Log("TUN interface configuration test completed")
}
