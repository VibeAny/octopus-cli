package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleServiceRestart_WithNoRunningService_ShouldSkipRestart(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")
	
	testConfig := `[server]
port = 8080

[[apis]]
id = "test-api"
name = "Test API"
url = "https://api.example.com"
api_key = "test-key"

[settings]
active_api = "test-api"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))
	
	// Create a mock command with output buffer
	var output bytes.Buffer
	cmd := newUpgradeCommand(&configFile, "v0.0.3")
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := handleServiceRestart(cmd, configFile)

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Service was not running - no restart needed")
}

func TestHandleServiceRestart_WithInvalidConfig_ShouldReturnError(t *testing.T) {
	// Arrange
	invalidConfigFile := "/path/that/does/not/exist.toml"
	
	// Create a mock command with output buffer
	var output bytes.Buffer
	cmd := newUpgradeCommand(&invalidConfigFile, "v0.0.3")
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := handleServiceRestart(cmd, invalidConfigFile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve config file")
}

func TestHandleServiceRestart_FunctionSignature_ShouldAcceptCorrectParameters(t *testing.T) {
	// This test verifies the function signature is correct
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")
	
	testConfig := `[server]
port = 8080

[settings]
active_api = ""
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))
	
	var output bytes.Buffer
	cmd := newUpgradeCommand(&configFile, "v0.0.3")
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act & Assert - function should be callable with these parameters
	assert.NotPanics(t, func() {
		_ = handleServiceRestart(cmd, configFile)
	}, "handleServiceRestart should accept Command and config file path")
}

func TestUpgradeCommand_Properties_ShouldHaveCorrectConfiguration(t *testing.T) {
	// Arrange
	configFile := ""
	version := "v0.0.3"

	// Act
	cmd := newUpgradeCommand(&configFile, version)

	// Assert
	assert.Equal(t, "upgrade", cmd.Use)
	assert.Equal(t, "Upgrade to the latest version", cmd.Short)
	assert.Contains(t, cmd.Long, "Check for the latest version")
	
	// Verify flags
	checkFlag := cmd.Flags().Lookup("check")
	assert.NotNil(t, checkFlag)
	assert.Equal(t, "false", checkFlag.DefValue)
	
	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "false", forceFlag.DefValue)
}

func TestUpgradeCommand_Help_ShouldContainUsageInformation(t *testing.T) {
	// Arrange
	configFile := ""
	version := "v0.0.3"
	cmd := newUpgradeCommand(&configFile, version)

	// Act
	helpOutput := cmd.UsageString()

	// Assert
	assert.Contains(t, helpOutput, "upgrade")
	assert.Contains(t, helpOutput, "--check")
	assert.Contains(t, helpOutput, "--force")
	
	// Check that the command has the right short description
	assert.Equal(t, "Upgrade to the latest version", cmd.Short)
}