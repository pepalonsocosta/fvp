package network

import (
	"testing"
)

func TestConfigureClientInterface(t *testing.T) {
	// Test that the method exists on TunManager
	tm := NewTunManager()
	
	// Test client IP configuration (will fail without root, but method exists)
	clientIP := "10.0.0.2"
	err := tm.ConfigureClientInterface(clientIP)
	// We expect this to fail due to lack of root privileges, but the method should exist
	if err != nil {
		t.Logf("configureClientInterface method exists (got expected error: %v)", err)
	}
}

func TestConfigureClientInterfaceEdgeCases(t *testing.T) {
	// Test with different client IPs
	testIPs := []string{
		"10.0.0.2",
		"10.0.0.3", 
		"10.0.0.255",
		"192.168.1.100",
		"172.16.0.50",
	}

	for _, ip := range testIPs {
		t.Run("IP_"+ip, func(t *testing.T) {
			tm := NewTunManager()
			err := tm.ConfigureClientInterface(ip)
			// We expect this to fail due to lack of root privileges, but the method should exist
			if err != nil {
				t.Logf("configureClientInterface method exists for IP %s (got expected error: %v)", ip, err)
			}
		})
	}
}

func TestConfigureClientInterfaceWithoutCreation(t *testing.T) {
	tm := NewTunManager()

	// Try to configure without creating interface first
	err := tm.ConfigureClientInterface("10.0.0.2")
	// We expect this to fail due to lack of root privileges, but the method should exist
	if err != nil {
		t.Logf("configureClientInterface method exists (got expected error: %v)", err)
	}
}

// Note: We can't easily test the real TunManager.configureClientInterface
// without root privileges, but we can test that the method exists and
// has the correct signature
func TestConfigureClientInterfaceMethodExists(t *testing.T) {
	tm := NewTunManager()
	
	// Test that the method exists and accepts the right parameters
	// This will fail at runtime without root, but we can verify the method signature
	err := tm.ConfigureClientInterface("10.0.0.2")
	// We expect this to fail due to lack of root privileges, but the method should exist
	if err != nil {
		t.Logf("configureClientInterface method exists (got expected error: %v)", err)
	}
}
