package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewServiceManager_ValidConfig_ShouldCreateManager tests service manager creation with valid config
func TestNewServiceManager_ValidConfig_ShouldCreateManager(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")
	
	testConfig := `[server]
port = 8080
log_level = "info"
pid_file = "octopus.pid"

[[apis]]
id = "test-api"
name = "Test API"
url = "https://api.test.com"
api_key = "test-key"
is_active = true

[settings]
active_api = "test-api"
log_file = "logs/octopus.log"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	// Act
	serviceManager, err := NewServiceManager(configFile)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, serviceManager)
	assert.NotNil(t, serviceManager.configManager)
	assert.NotNil(t, serviceManager.processManager)
	assert.NotNil(t, serviceManager.proxyServer)
	assert.Equal(t, configFile, serviceManager.configFile)
}

// TestNewServiceManager_InvalidConfig_ShouldReturnError tests service manager creation with invalid config
func TestNewServiceManager_InvalidConfig_ShouldReturnError(t *testing.T) {
	// Arrange
	invalidConfigFile := "/nonexistent/config.toml"

	// Act
	serviceManager, err := NewServiceManager(invalidConfigFile)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, serviceManager)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

// TestServiceManager_Status_ShouldReturnServiceStatus tests status retrieval
func TestServiceManager_Status_ShouldReturnServiceStatus(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")
	
	testConfig := `[server]
port = 8080
log_level = "info"
pid_file = "octopus.pid"

[[apis]]
id = "test-api"
name = "Test API"
url = "https://api.test.com"
api_key = "test-key"
is_active = true

[settings]
active_api = "test-api"
log_file = "logs/octopus.log"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	serviceManager, err := NewServiceManager(configFile)
	require.NoError(t, err)

	// Act
	status, err := serviceManager.Status()

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, 8080, status.Port)
	assert.Equal(t, "test-api", status.ActiveAPI)
	assert.False(t, status.IsRunning) // Should not be running in test
}

// TestServiceManager_Start_WhenAlreadyRunning_ShouldReturnError tests starting when already running
func TestServiceManager_Start_WhenAlreadyRunning_ShouldReturnError(t *testing.T) {
	// This test is difficult to implement without mocking because it requires
	// actual process management. In a real test environment, we would mock
	// the process manager to simulate a running state.
	t.Skip("Requires process manager mocking for complete implementation")
}

// TestServiceManager_Stop_WhenNotRunning_ShouldReturnError tests stopping when not running
func TestServiceManager_Stop_WhenNotRunning_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.toml")
	
	testConfig := `[server]
port = 8080
log_level = "info"
pid_file = "octopus.pid"

[[apis]]
id = "test-api"
name = "Test API"
url = "https://api.test.com"
api_key = "test-key"
is_active = true

[settings]
active_api = "test-api"
log_file = "logs/octopus.log"
`
	require.NoError(t, os.WriteFile(configFile, []byte(testConfig), 0644))

	serviceManager, err := NewServiceManager(configFile)
	require.NoError(t, err)

	// Act
	err = serviceManager.Stop()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service is not running")
}

// TestServiceStatus_Properties_ShouldHaveExpectedFields tests service status structure
func TestServiceStatus_Properties_ShouldHaveExpectedFields(t *testing.T) {
	// Arrange
	status := &ServiceStatus{
		IsRunning: true,
		PID:       12345,
		Port:      8080,
		ActiveAPI: "test-api",
	}

	// Assert
	assert.True(t, status.IsRunning)
	assert.Equal(t, 12345, status.PID)
	assert.Equal(t, 8080, status.Port)
	assert.Equal(t, "test-api", status.ActiveAPI)
}