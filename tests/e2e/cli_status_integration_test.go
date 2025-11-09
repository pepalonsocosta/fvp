package e2e

import (
	"testing"
)

// TestCLIStatusIntegration tests the new status and real-time client listing functionality
func TestCLIStatusIntegration(t *testing.T) {
	te := SetupTestEnvironment(t)
	defer te.CleanupTestEnvironment()

	// Test 1: Setup server configuration
	t.Run("SetupServer", func(t *testing.T) {
		output, err := te.RunCommand(t, "setup", "--port", "1195", "--timeout", "15")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
		AssertOutputContains(t, output, "Configuration created: server.yaml")
		AssertOutputContains(t, output, "Server will listen on port 1195")
	})

	// Test 2: Add a client
	t.Run("AddClient", func(t *testing.T) {
		output, err := te.RunCommand(t, "add-client")
		if err != nil {
			t.Fatalf("Add client failed: %v", err)
		}
		AssertOutputContains(t, output, "Client added successfully")
		AssertOutputContains(t, output, "Client ID: 1")
		AssertOutputContains(t, output, "Key:")
	})

	// Test 3: Check status when server is stopped
	t.Run("StatusWhenStopped", func(t *testing.T) {
		output, err := te.RunCommand(t, "status")
		if err != nil {
			t.Fatalf("Status command failed: %v", err)
		}
		AssertOutputContains(t, output, "Server Status:")
		AssertOutputContains(t, output, "Status: stopped")
	})

	// Test 4: List clients (should show config data when server stopped)
	t.Run("ListClientsWhenStopped", func(t *testing.T) {
		output, err := te.RunCommand(t, "list-clients")
		if err != nil {
			t.Fatalf("List clients failed: %v", err)
		}
		AssertOutputContains(t, output, "Client Status:")
		AssertOutputContains(t, output, "ID  IP         Status     Last Connection")
		AssertOutputContains(t, output, "1   10.0.0.2   Disconnected Never")
	})

	// Test 5: Add another client
	t.Run("AddSecondClient", func(t *testing.T) {
		output, err := te.RunCommand(t, "add-client")
		if err != nil {
			t.Fatalf("Add second client failed: %v", err)
		}
		AssertOutputContains(t, output, "Client ID: 2")
	})

	// Test 6: List clients with multiple clients
	t.Run("ListMultipleClients", func(t *testing.T) {
		output, err := te.RunCommand(t, "list-clients")
		if err != nil {
			t.Fatalf("List multiple clients failed: %v", err)
		}
		AssertOutputContains(t, output, "1   10.0.0.2   Disconnected Never")
		AssertOutputContains(t, output, "2   10.0.0.3   Disconnected Never")
	})

	// Test 7: Remove a client
	t.Run("RemoveClient", func(t *testing.T) {
		output, err := te.RunCommand(t, "remove-client", "--id", "2")
		if err != nil {
			t.Fatalf("Remove client failed: %v", err)
		}
		AssertOutputContains(t, output, "Client 2 removed successfully")
	})

	// Test 8: Verify client was removed
	t.Run("VerifyClientRemoved", func(t *testing.T) {
		output, err := te.RunCommand(t, "list-clients")
		if err != nil {
			t.Fatalf("List clients after removal failed: %v", err)
		}
		AssertOutputContains(t, output, "1   10.0.0.2   Disconnected Never")
		AssertOutputNotContains(t, output, "2   10.0.0.3")
	})
}

// TestCLIStatusErrorHandling tests error scenarios for the new commands
func TestCLIStatusErrorHandling(t *testing.T) {
	te := SetupTestEnvironment(t)
	defer te.CleanupTestEnvironment()

	// Test 1: Status without config
	t.Run("StatusWithoutConfig", func(t *testing.T) {
		output, err := te.RunCommand(t, "status")
		if err != nil {
			t.Fatalf("Status command failed: %v", err)
		}
		AssertOutputContains(t, output, "Status: stopped")
	})

	// Test 2: List clients without config
	t.Run("ListClientsWithoutConfig", func(t *testing.T) {
		output, err := te.RunCommand(t, "list-clients")
		if err == nil {
			t.Fatalf("Expected error for list-clients without config, but got success")
		}
		AssertOutputContains(t, output, "no configuration found, run 'fvps setup' first")
	})

	// Test 3: Remove non-existent client
	t.Run("RemoveNonExistentClient", func(t *testing.T) {
		// First setup
		_, err := te.RunCommand(t, "setup", "--port", "1196", "--timeout", "15")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		output, err := te.RunCommand(t, "remove-client", "--id", "999")
		if err == nil {
			t.Fatalf("Expected error for non-existent client, but got success")
		}
		AssertOutputContains(t, output, "Failed to remove client: client 231 not found")
	})

	// Test 4: Remove client without ID
	t.Run("RemoveClientWithoutID", func(t *testing.T) {
		output, err := te.RunCommand(t, "remove-client")
		if err == nil {
			t.Fatalf("Expected error for missing ID, but got success")
		}
		AssertOutputContains(t, output, "Error: --id is required")
	})
}

// TestCLIStatusHelp tests the help and usage information
func TestCLIStatusHelp(t *testing.T) {
	te := SetupTestEnvironment(t)
	defer te.CleanupTestEnvironment()

	// Test 1: Help command
	t.Run("HelpCommand", func(t *testing.T) {
		output, err := te.RunCommand(t, "help")
		if err != nil {
			t.Fatalf("Help command failed: %v", err)
		}
		AssertOutputContains(t, output, "FVP Server - Fast VPN Server")
		AssertOutputContains(t, output, "status        Show server status")
		AssertOutputContains(t, output, "fvps status")
	})

	// Test 2: Version command
	t.Run("VersionCommand", func(t *testing.T) {
		output, err := te.RunCommand(t, "version")
		if err != nil {
			t.Fatalf("Version command failed: %v", err)
		}
		AssertOutputContains(t, output, "FVP Server version 1.0.0")
	})

	// Test 3: Unknown command
	t.Run("UnknownCommand", func(t *testing.T) {
		output, err := te.RunCommand(t, "unknown-command")
		if err == nil {
			t.Fatalf("Expected error for unknown command, but got success")
		}
		AssertOutputContains(t, output, "Unknown command: unknown-command")
		AssertOutputContains(t, output, "fvps <command> [flags]")
	})
}