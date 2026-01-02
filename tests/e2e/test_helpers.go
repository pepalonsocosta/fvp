package e2e

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestConfig represents the server configuration for testing
type TestConfig struct {
	Server struct {
		Port           string `yaml:"port"`
		TimeoutMinutes int    `yaml:"timeout_minutes"`
	} `yaml:"server"`
	Clients []struct {
		ID  uint8  `yaml:"id"`
		Key string `yaml:"key"`
	} `yaml:"clients"`
}

type TestEnvironment struct {
	TestDir    string
	ConfigPath string
	OriginalDir string
}

func getProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	for {
		goModPath := filepath.Join(cwd, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return cwd, nil
		}
		
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break // Reached root directory
		}
		cwd = parent
	}
	
	return "", fmt.Errorf("project root not found")
}

// SetupTestEnvironment creates a clean test environment
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	// Create a temporary directory for test files
	testDir := t.TempDir()
	configPath := filepath.Join(testDir, "server.yaml")
	
	// Get the project root directory
	projectRoot, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to get project root directory: %v", err)
	}
	
	// Build fvps binary with test version for consistent testing
	fvpsPath := filepath.Join(projectRoot, "fvps")
	if err := buildBinaryWithVersion(t, projectRoot, "fvps", "./cmd/server", "1.0.0"); err != nil {
		t.Fatalf("Failed to build fvps binary: %v", err)
	}
	// Verify binary exists
	if _, err := os.Stat(fvpsPath); os.IsNotExist(err) {
		t.Fatalf("fvps binary not found after build at %s", fvpsPath)
	}
	
	// Change to test directory
	err = os.Chdir(testDir)
	if err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	return &TestEnvironment{
		TestDir:     testDir,
		ConfigPath:  configPath,
		OriginalDir: projectRoot,
	}
}

// buildBinaryWithVersion builds a binary with a specific version injected
func buildBinaryWithVersion(t *testing.T, projectRoot, binaryName, mainPath, version string) error {
	binaryPath := filepath.Join(projectRoot, binaryName)
	
	// Build command
	cmd := exec.Command("go", "build", 
		"-ldflags", fmt.Sprintf("-X main.version=%s", version),
		"-o", binaryPath,
		mainPath,
	)
	cmd.Dir = projectRoot
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build %s: %v\nOutput: %s", binaryName, err, string(output))
	}
	
	return nil
}

// CleanupTestEnvironment cleans up the test environment
func (te *TestEnvironment) CleanupTestEnvironment() {
	os.Chdir(te.OriginalDir)
}

// RunCommand runs a CLI command and returns output
func (te *TestEnvironment) RunCommand(t *testing.T, args ...string) (string, error) {
	// Get the absolute path to the fvps binary from the original directory
	fvpsPath := filepath.Join(te.OriginalDir, "fvps")
	
	// Verify the binary exists
	if _, err := os.Stat(fvpsPath); os.IsNotExist(err) {
		t.Fatalf("fvps binary not found at %s", fvpsPath)
	}
	
	// Build command with fvps binary path
	cmdArgs := append([]string{fvpsPath}, args...)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCommandExpectSuccess runs a CLI command and expects success
func (te *TestEnvironment) RunCommandExpectSuccess(t *testing.T, args ...string) string {
	output, err := te.RunCommand(t, args...)
	if err != nil {
		t.Fatalf("Command %v failed: %v\nOutput: %s", args, err, output)
	}
	return output
}

// RunCommandExpectFailure runs a CLI command and expects failure
func (te *TestEnvironment) RunCommandExpectFailure(t *testing.T, args ...string) string {
	output, err := te.RunCommand(t, args...)
	if err == nil {
		t.Fatalf("Command %v should have failed but succeeded\nOutput: %s", args, output)
	}
	return output
}

// AssertOutputContains checks if output contains expected text
func AssertOutputContains(t *testing.T, output, expected string) {
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain '%s', got: %s", expected, output)
	}
}

// AssertOutputNotContains checks if output does not contain unexpected text
func AssertOutputNotContains(t *testing.T, output, unexpected string) {
	if strings.Contains(output, unexpected) {
		t.Errorf("Expected output to not contain '%s', got: %s", unexpected, output)
	}
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, filepath string) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", filepath)
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, filepath string) {
	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		t.Errorf("Expected file %s to not exist", filepath)
	}
}

// AssertConfigFileValid checks if the config file is valid YAML
func (te *TestEnvironment) AssertConfigFileValid(t *testing.T) {
	// Try to load the config file
	config, err := te.LoadConfig(te.ConfigPath)
	if err != nil {
		t.Fatalf("Config file is not valid YAML: %v", err)
	}
	
	// Basic validation
	if config.Server.Port == "" {
		t.Error("Config file missing server port")
	}
	
	if config.Server.TimeoutMinutes == 0 {
		t.Error("Config file missing server timeout")
	}
}

// LoadConfig loads a config file
func LoadConfig(path string) (*TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config TestConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfig loads a config file
func (te *TestEnvironment) LoadConfig(path string) (*TestConfig, error) {
	return LoadConfig(path)
}

// AssertClientCount checks if the config has the expected number of clients
func (te *TestEnvironment) AssertClientCount(t *testing.T, expectedCount int) {
	config, err := te.LoadConfig(te.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if len(config.Clients) != expectedCount {
		t.Errorf("Expected %d clients, got %d", expectedCount, len(config.Clients))
	}
}

// AssertClientExists checks if a client with the given ID exists
func (te *TestEnvironment) AssertClientExists(t *testing.T, clientID uint8) {
	config, err := te.LoadConfig(te.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	for _, client := range config.Clients {
		if client.ID == clientID {
			return
		}
	}
	
	t.Errorf("Client with ID %d not found", clientID)
}

// AssertClientNotExists checks if a client with the given ID does not exist
func (te *TestEnvironment) AssertClientNotExists(t *testing.T, clientID uint8) {
	config, err := te.LoadConfig(te.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	for _, client := range config.Clients {
		if client.ID == clientID {
			t.Errorf("Client with ID %d should not exist", clientID)
			return
		}
	}
}

// AssertClientKeyValid checks if a client's key is valid
func (te *TestEnvironment) AssertClientKeyValid(t *testing.T, clientID uint8) {
	config, err := te.LoadConfig(te.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	for _, client := range config.Clients {
		if client.ID == clientID {
			// Check if key is valid hex
			key, err := hex.DecodeString(client.Key)
			if err != nil {
				t.Errorf("Client %d has invalid hex key: %v", clientID, err)
				return
			}
			
			// Check if key is 32 bytes
			if len(key) != 32 {
				t.Errorf("Client %d key should be 32 bytes, got %d", clientID, len(key))
				return
			}
			
			return
		}
	}
	
	t.Errorf("Client with ID %d not found", clientID)
}
