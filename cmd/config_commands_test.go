package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigListCommand_Execute_ShouldListAllAPIs tests the config list functionality
func TestConfigListCommand_Execute_ShouldListAllAPIs(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	testConfig := `[server]
port = 8080

[[apis]]
id = "api1"
name = "API One"
url = "https://api1.example.com"
api_key = "key1"

[[apis]]
id = "api2"
name = "API Two"
url = "https://api2.example.com"
api_key = "key2"

[settings]
active_api = "api1"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigListCommand(&configFile, stateManager)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "API Configurations:")
	assert.Contains(t, outputStr, "api1")
	assert.Contains(t, outputStr, "API One")
	assert.Contains(t, outputStr, "https://api1.example.com")
	assert.Contains(t, outputStr, "api2")
	assert.Contains(t, outputStr, "API Two")
	assert.Contains(t, outputStr, "active") // Should mark active API status
}

func TestConfigListCommand_Execute_WithNoAPIs_ShouldShowEmptyMessage(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	testConfig := `[server]
port = 8080

[settings]
active_api = ""
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigListCommand(&configFile, stateManager)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	// Since default config now includes example APIs, we check for their presence instead
	assert.Contains(t, outputStr, "official-example")
	assert.Contains(t, outputStr, "proxy-example")
}

// TestConfigAddCommand_Execute_ShouldAddNewAPI tests adding a new API configuration
func TestConfigAddCommand_Execute_ShouldAddNewAPI(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	// Create initial config with one API
	testConfig := `[server]
port = 8080

[[apis]]
id = "existing"
name = "Existing API"
url = "https://existing.com"

[settings]
active_api = "existing"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigAddCommand(&configFile, stateManager)
	cmd.SetArgs([]string{"new-api", "https://new.example.com", "sk-new-key"})

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Added API configuration: new-api")

	// Verify the config was actually saved
	stateManager = createTestStateManager(t)
	listCmd := newConfigListCommand(&configFile, stateManager)
	var listOutput bytes.Buffer
	listCmd.SetOut(&listOutput)
	listCmd.SetErr(&listOutput)

	require.NoError(t, listCmd.Execute())
	listOutputStr := listOutput.String()
	assert.Contains(t, listOutputStr, "new-api")
	assert.Contains(t, listOutputStr, "https://new.example.com")
}

func TestConfigAddCommand_Execute_WithDuplicateID_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	testConfig := `[server]
port = 8080

[[apis]]
id = "existing"
name = "Existing API"
url = "https://existing.com"

[settings]
active_api = "existing"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigAddCommand(&configFile, stateManager)
	cmd.SetArgs([]string{"existing", "https://duplicate.com", "sk-duplicate"})

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	assert.Error(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "already exists")
}

// TestConfigSwitchCommand_Execute_ShouldSwitchActiveAPI tests switching active API
func TestConfigSwitchCommand_Execute_ShouldSwitchActiveAPI(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	testConfig := `[server]
port = 8080

[[apis]]
id = "api1"
name = "API One"
url = "https://api1.com"

[[apis]]
id = "api2"  
name = "API Two"
url = "https://api2.com"

[settings]
active_api = "api1"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigSwitchCommand(&configFile, stateManager)
	cmd.SetArgs([]string{"api2"})

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Switched to API: api2")

	// Verify the switch was saved
	stateManager = createTestStateManager(t)
	listCmd := newConfigListCommand(&configFile, stateManager)
	var listOutput bytes.Buffer
	listCmd.SetOut(&listOutput)
	listCmd.SetErr(&listOutput)

	require.NoError(t, listCmd.Execute())
	listOutputStr := listOutput.String()

	// Should show api2 as active now
	lines := strings.Split(listOutputStr, "\n")
	api2Line := ""
	for _, line := range lines {
		if strings.Contains(line, "api2") {
			api2Line = line
			break
		}
	}
	assert.Contains(t, api2Line, "active")
}

func TestConfigSwitchCommand_Execute_WithNonExistentAPI_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	testConfig := `[server]
port = 8080

[[apis]]
id = "api1"
name = "API One"
url = "https://api1.com"

[settings]
active_api = "api1"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigSwitchCommand(&configFile, stateManager)
	cmd.SetArgs([]string{"nonexistent"})

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	assert.Error(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "not found")
}

// TestConfigRemoveCommand_Execute_ShouldRemoveAPI tests removing an API
func TestConfigRemoveCommand_Execute_ShouldRemoveAPI(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	testConfig := `[server]
port = 8080

[[apis]]
id = "api1"
name = "API One"
url = "https://api1.com"

[[apis]]
id = "api2"
name = "API Two"
url = "https://api2.com"

[settings]
active_api = "api1"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigRemoveCommand(&configFile, stateManager)
	cmd.SetArgs([]string{"api2"})

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Removed API configuration: api2")

	// Verify it was actually removed
	stateManager = createTestStateManager(t)
	listCmd := newConfigListCommand(&configFile, stateManager)
	var listOutput bytes.Buffer
	listCmd.SetOut(&listOutput)
	listCmd.SetErr(&listOutput)

	require.NoError(t, listCmd.Execute())
	listOutputStr := listOutput.String()
	assert.NotContains(t, listOutputStr, "api2")
	assert.Contains(t, listOutputStr, "api1") // Should still have api1
}

func TestConfigRemoveCommand_Execute_RemoveActiveAPI_ShouldClearActive(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	testConfig := `[server]
port = 8080

[[apis]]
id = "api1"
name = "API One"
url = "https://api1.com"

[settings]
active_api = "api1"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigRemoveCommand(&configFile, stateManager)
	cmd.SetArgs([]string{"api1"})

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Removed API configuration: api1")
	assert.Contains(t, outputStr, "Cleared active API")
}

// TestConfigShowCommand_Execute_ShouldShowAPIDetails tests showing API details
func TestConfigShowCommand_Execute_ShouldShowAPIDetails(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")

	testConfig := `[server]
port = 8080

[[apis]]
id = "api1"
name = "API One"
url = "https://api1.example.com"
api_key = "sk-secret-key"
timeout = 30
retry_count = 3

[settings]
active_api = "api1"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	stateManager := createTestStateManager(t)
	cmd := newConfigShowCommand(&configFile, stateManager)
	cmd.SetArgs([]string{"api1"})

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "API Configuration: api1")
	assert.Contains(t, outputStr, "Name: API One")
	assert.Contains(t, outputStr, "URL: https://api1.example.com")
	assert.Contains(t, outputStr, "API Key: sk-***") // Should mask the key
	assert.Contains(t, outputStr, "Timeout: 30")
	assert.Contains(t, outputStr, "Retry Count: 3")
	assert.Contains(t, outputStr, "Status: Active")
}
