package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DevCommand != nil {
		t.Errorf("Expected DevCommand to be nil, got %v", config.DevCommand)
	}

	if config.PortRange != nil {
		t.Errorf("Expected PortRange to be nil, got %v", config.PortRange)
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	expected := "uzi.yaml"
	actual := GetDefaultConfigPath()

	if actual != expected {
		t.Errorf("Expected default config path %q, got %q", expected, actual)
	}
}

func TestLoadConfig_Success(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Create a valid YAML config file
	configContent := `devCommand: "npm start"
portRange: "3000-4000"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	if config.DevCommand == nil || *config.DevCommand != "npm start" {
		t.Errorf("Expected DevCommand to be 'npm start', got %v", config.DevCommand)
	}

	if config.PortRange == nil || *config.PortRange != "3000-4000" {
		t.Errorf("Expected PortRange to be '3000-4000', got %v", config.PortRange)
	}
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "partial-config.yaml")

	// Create a config file with only one field
	configContent := `devCommand: "go run main.go"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	if config.DevCommand == nil || *config.DevCommand != "go run main.go" {
		t.Errorf("Expected DevCommand to be 'go run main.go', got %v", config.DevCommand)
	}

	// PortRange should be nil since it wasn't specified
	if config.PortRange != nil {
		t.Errorf("Expected PortRange to be nil, got %v", config.PortRange)
	}
}

func TestLoadConfig_EmptyConfig(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty-config.yaml")

	// Create an empty config file
	configContent := ``

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Both fields should be nil for empty config
	if config.DevCommand != nil {
		t.Errorf("Expected DevCommand to be nil, got %v", config.DevCommand)
	}

	if config.PortRange != nil {
		t.Errorf("Expected PortRange to be nil, got %v", config.PortRange)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	// Try to load a non-existent file
	nonExistentPath := "/path/that/does/not/exist/config.yaml"

	config, err := LoadConfig(nonExistentPath)
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}

	if config != nil {
		t.Errorf("Expected config to be nil for missing file, got %v", config)
	}

	// Check that the error is related to file not found
	if !os.IsNotExist(err) {
		t.Errorf("Expected file not found error, got %v", err)
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "malformed-config.yaml")

	// Create a malformed YAML config file
	configContent := `devCommand: "npm start"
portRange: [invalid yaml structure
  missing: close bracket
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the malformed config
	config, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for malformed YAML, got nil")
	}

	if config != nil {
		t.Errorf("Expected config to be nil for malformed YAML, got %v", config)
	}
}

func TestLoadConfig_InvalidYAMLTypes(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid-types-config.yaml")

	// Create a YAML config file with invalid types (arrays instead of strings)
	configContent := `devCommand: ["not", "a", "string"]
portRange: 1234
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config with invalid types
	config, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML types, got nil")
	}

	if config != nil {
		t.Errorf("Expected config to be nil for invalid YAML types, got %v", config)
	}
}

func TestLoadConfig_PermissionDenied(t *testing.T) {
	// Skip this test on Windows as it doesn't handle permissions the same way
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no-read-permission.yaml")

	// Create a config file
	configContent := `devCommand: "npm start"`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Remove read permissions
	err = os.Chmod(configPath, 0000)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}

	// Restore permissions after test for cleanup
	defer func() {
		os.Chmod(configPath, 0644)
	}()

	// Test loading the config without read permissions
	config, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for permission denied, got nil")
	}

	if config != nil {
		t.Errorf("Expected config to be nil for permission denied, got %v", config)
	}
}
