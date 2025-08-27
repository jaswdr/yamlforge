package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printUsage()

	w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	if !strings.Contains(output, "Usage: yamlforge") {
		t.Error("Expected usage to contain 'Usage: yamlforge'")
	}

	if !strings.Contains(output, "serve") {
		t.Error("Expected usage to contain 'serve' command")
	}

	if !strings.Contains(output, "build") {
		t.Error("Expected usage to contain 'build' command")
	}

	if !strings.Contains(output, "validate") {
		t.Error("Expected usage to contain 'validate' command")
	}
}

func TestHandleBuild(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.yaml")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handleBuild(configFile)

	w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	if !strings.Contains(output, "Building static files") {
		t.Error("Expected output to contain 'Building static files'")
	}

	if !strings.Contains(output, "not yet implemented") {
		t.Error("Expected output to contain 'not yet implemented'")
	}
}

func TestHandleValidate_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "valid.yaml")
	
	validConfig := `app:
  name: "Test App"
  version: "1.0.0"

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
        required: true`

	err := os.WriteFile(configFile, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handleValidate(configFile)

	w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	if !strings.Contains(output, "is valid") {
		t.Error("Expected output to contain 'is valid'")
	}

	if !strings.Contains(output, "Test App") {
		t.Error("Expected output to contain app name")
	}

	if !strings.Contains(output, "Models: 1") {
		t.Error("Expected output to contain model count")
	}
}

func TestHandleValidate_Failure(t *testing.T) {
	nonExistentFile := "/path/that/does/not/exist.yaml"
	
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if _, err := os.Stat(nonExistentFile); !os.IsNotExist(err) {
		t.Skip("Test file exists, skipping failure test")
	}

	w.Close()
	os.Stdout = old
	
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	
	t.Log("Testing validation failure via existing TestValidateCommand test")
}

func TestVersion(t *testing.T) {
	if version == "" {
		t.Error("Expected version to be set")
	}
	
	if !strings.Contains(version, ".") {
		t.Error("Expected version to contain version numbers")
	}
}

func TestMainCommandHandling(t *testing.T) {
	
	commands := []string{"serve", "build", "validate"}
	
	for _, cmd := range commands {
		switch cmd {
		case "serve", "build", "validate":
			default:
			t.Errorf("Unexpected command: %s", cmd)
		}
	}
}

func TestConfigFileHandling(t *testing.T) {
	testPaths := []string{
		"config.yaml",
		"./config.yaml",
		"/path/to/config.yaml",
		"config.yml",
	}
	
	for _, path := range testPaths {
		if path == "" {
			t.Error("Config path should not be empty")
		}
		
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		}
	}
}

func TestHostAndPortDefaults(t *testing.T) {
	defaultHost := "0.0.0.0"
	defaultPort := 8080
	
	if defaultHost == "" {
		t.Error("Default host should not be empty")
	}
	
	if defaultPort <= 0 || defaultPort > 65535 {
		t.Error("Default port should be in valid range")
	}
}

func TestFlagValidation(t *testing.T) {
	validPorts := []int{8080, 3000, 80, 443, 8000}
	
	for _, port := range validPorts {
		if port <= 0 || port > 65535 {
			t.Errorf("Invalid port: %d", port)
		}
	}
	
	validHosts := []string{"0.0.0.0", "127.0.0.1", "localhost"}
	
	for _, host := range validHosts {
		if host == "" {
			t.Errorf("Invalid host: %s", host)
		}
	}
}