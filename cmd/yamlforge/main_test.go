package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectedOutput string
		expectError    bool
	}{
		{
			name: "valid config",
			configContent: `app:
  name: "Test App"
  version: "1.0.0"
  description: "A test application"

database:
  type: sqlite
  path: "./test.db"

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true`,
			expectedOutput: "Configuration",
			expectError:    false,
		},
		{
			name: "invalid config - empty app name",
			configContent: `app:
  name: ""
  version: "1.0.0"

models:
  User:
    fields:
      id:
        type: id
        primary: true`,
			expectedOutput: "app.name is required",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Build the yamlforge binary
			binaryPath := filepath.Join(tmpDir, "yamlforge")
			buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
			buildCmd.Dir = "."
			if err := buildCmd.Run(); err != nil {
				t.Fatalf("Failed to build yamlforge binary: %v", err)
			}

			// Run validate command
			cmd := exec.Command(binaryPath, "validate", configFile)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Check expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation to fail, but it succeeded. Output: %s", outputStr)
				}
				if !strings.Contains(outputStr, tt.expectedOutput) {
					t.Errorf("Expected output to contain '%s', but got: %s", tt.expectedOutput, outputStr)
				}
			} else {
				if err != nil {
					t.Errorf("Expected validation to succeed, but it failed with error: %v. Output: %s", err, outputStr)
				}
				if !strings.Contains(outputStr, tt.expectedOutput) {
					t.Errorf("Expected output to contain '%s', but got: %s", tt.expectedOutput, outputStr)
				}
			}
		})
	}
}

func TestValidateCommandMissingFile(t *testing.T) {
	// Build the yamlforge binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "yamlforge")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build yamlforge binary: %v", err)
	}

	// Run validate command with missing file
	cmd := exec.Command(binaryPath, "validate", "nonexistent.yaml")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err == nil {
		t.Errorf("Expected validation to fail with missing file, but it succeeded")
	}

	if !strings.Contains(outputStr, "Validation failed") {
		t.Errorf("Expected output to contain 'Validation failed', but got: %s", outputStr)
	}
}

func TestValidateCommandMissingArgument(t *testing.T) {
	// Build the yamlforge binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "yamlforge")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build yamlforge binary: %v", err)
	}

	// Run validate command without config file argument
	cmd := exec.Command(binaryPath, "validate")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err == nil {
		t.Errorf("Expected validation to fail with missing argument, but it succeeded")
	}

	if !strings.Contains(outputStr, "Error: missing YAML configuration file") {
		t.Errorf("Expected output to contain 'Error: missing YAML configuration file', but got: %s", outputStr)
	}
}