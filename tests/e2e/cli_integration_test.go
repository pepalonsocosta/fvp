package e2e

import (
	"testing"
)

// TestCLIIntegration tests basic CLI functionality with helper functions
func TestCLIIntegration(t *testing.T) {
	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.CleanupTestEnvironment()

	// Test 1: Setup command
	t.Run("SetupIntegration", func(t *testing.T) {
		output := env.RunCommandExpectSuccess(t, "setup", "--port", "1194", "--timeout", "30")
		
		// Verify output
		AssertOutputContains(t, output, "Configuration created: server.yaml")
		AssertOutputContains(t, output, "Server will listen on port 1194")
		AssertOutputContains(t, output, "Client timeout: 30 minutes")
		
		// Verify config file
		env.AssertConfigFileValid(t)
		env.AssertClientCount(t, 0)
	})

	// Test 2: Add client
	t.Run("AddClientIntegration", func(t *testing.T) {
		output := env.RunCommandExpectSuccess(t, "add-client")
		
		// Verify output
		AssertOutputContains(t, output, "Client added successfully")
		AssertOutputContains(t, output, "Client ID: 1")
		AssertOutputContains(t, output, "Key: ")
		
		// Verify config file
		env.AssertClientCount(t, 1)
		env.AssertClientExists(t, 1)
		env.AssertClientKeyValid(t, 1)
	})

	// Test 3: Add second client
	t.Run("AddSecondClientIntegration", func(t *testing.T) {
		output := env.RunCommandExpectSuccess(t, "add-client")
		
		// Verify output
		AssertOutputContains(t, output, "Client ID: 2")
		
		// Verify config file
		env.AssertClientCount(t, 2)
		env.AssertClientExists(t, 2)
		env.AssertClientKeyValid(t, 2)
	})

	// Test 4: List clients
	t.Run("ListClientsIntegration", func(t *testing.T) {
		output := env.RunCommandExpectSuccess(t, "list-clients")
		
		// Verify output
		AssertOutputContains(t, output, "Client Status:")
		AssertOutputContains(t, output, "ID  IP         Status     Last Connection")
		AssertOutputContains(t, output, "1   10.0.0.2   Disconnected Never")
		AssertOutputContains(t, output, "2   10.0.0.3   Disconnected Never")
	})

	// Test 5: Remove client
	t.Run("RemoveClientIntegration", func(t *testing.T) {
		output := env.RunCommandExpectSuccess(t, "remove-client", "--id", "2")
		
		// Verify output
		AssertOutputContains(t, output, "Client 2 removed successfully")
		
		// Verify config file
		env.AssertClientCount(t, 1)
		env.AssertClientExists(t, 1)
		env.AssertClientNotExists(t, 2)
	})

	// Test 6: List clients after removal
	t.Run("ListClientsAfterRemovalIntegration", func(t *testing.T) {
		output := env.RunCommandExpectSuccess(t, "list-clients")
		
		// Verify output
		AssertOutputContains(t, output, "1   10.0.0.2   Disconnected Never")
		AssertOutputNotContains(t, output, "2   10.0.0.3")
	})
}

// TestCLIErrorHandling tests error conditions
func TestCLIErrorHandling(t *testing.T) {
	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.CleanupTestEnvironment()

	// Test 1: Setup without required flags
	t.Run("SetupWithoutFlagsError", func(t *testing.T) {
		output := env.RunCommandExpectFailure(t, "setup")
		AssertOutputContains(t, output, "--port and --timeout are required")
	})

	// Test 2: Add client without config
	t.Run("AddClientWithoutConfigError", func(t *testing.T) {
		output := env.RunCommandExpectFailure(t, "add-client")
		AssertOutputContains(t, output, "no configuration found, run 'fvps setup' first")
	})

	// Test 3: List clients without config
	t.Run("ListClientsWithoutConfigError", func(t *testing.T) {
		output := env.RunCommandExpectFailure(t, "list-clients")
		AssertOutputContains(t, output, "no configuration found, run 'fvps setup' first")
	})

	// Test 4: Remove client without config
	t.Run("RemoveClientWithoutConfigError", func(t *testing.T) {
		output := env.RunCommandExpectFailure(t, "remove-client", "--id", "1")
		AssertOutputContains(t, output, "no configuration found, run 'fvps setup' first")
	})

	// Test 5: Remove non-existent client
	t.Run("RemoveNonExistentClientError", func(t *testing.T) {
		// First create a config
		env.RunCommandExpectSuccess(t, "setup", "--port", "1194", "--timeout", "30")
		
		// Try to remove non-existent client
		output := env.RunCommandExpectFailure(t, "remove-client", "--id", "999")
		AssertOutputContains(t, output, "not found")
	})
}

// TestCLIHelpAndVersionIntegration tests help and version commands
func TestCLIHelpAndVersionIntegration(t *testing.T) {
	// Setup test environment
	env := SetupTestEnvironment(t)
	defer env.CleanupTestEnvironment()

	// Test 1: Help command
	t.Run("HelpIntegration", func(t *testing.T) {
		output := env.RunCommandExpectSuccess(t, "help")
		
		AssertOutputContains(t, output, "FVP Server - Fast VPN Server")
		AssertOutputContains(t, output, "Commands:")
		AssertOutputContains(t, output, "setup")
		AssertOutputContains(t, output, "up")
		AssertOutputContains(t, output, "add-client")
		AssertOutputContains(t, output, "list-clients")
		AssertOutputContains(t, output, "remove-client")
	})

	// Test 2: Version command
	t.Run("VersionIntegration", func(t *testing.T) {
		output := env.RunCommandExpectSuccess(t, "version")
		
		AssertOutputContains(t, output, "FVP Server version 1.0.0")
	})

	// Test 3: No arguments (should show help)
	t.Run("NoArgumentsIntegration", func(t *testing.T) {
		output := env.RunCommandExpectFailure(t)
		
		AssertOutputContains(t, output, "FVP Server - Fast VPN Server")
		AssertOutputContains(t, output, "Commands:")
	})
}
