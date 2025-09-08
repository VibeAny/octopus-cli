package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig_Creation_ShouldHaveCorrectDefaults(t *testing.T) {
	// Act
	config := DefaultConfig()

	// Assert
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "info", config.Server.LogLevel)
	assert.True(t, config.Server.Daemon)
	assert.Equal(t, "octopus.pid", config.Server.PIDFile)
	
	assert.Empty(t, config.APIs)
	assert.Equal(t, "", config.Settings.ActiveAPI)
	assert.Equal(t, "logs/octopus.log", config.Settings.LogFile)
	assert.True(t, config.Settings.ConfigBackup)
}

func TestNewManager_WithDefaultPath_ShouldUseHomeDirectory(t *testing.T) {
	// Act
	manager := NewManager("")

	// Assert
	assert.NotNil(t, manager)
	// The config path should contain the user's home directory
	assert.Contains(t, manager.configPath, ".config/octopus/octopus.toml")
}

func TestNewManager_WithCustomPath_ShouldUseProvidedPath(t *testing.T) {
	// Arrange
	customPath := "/custom/path/config.toml"

	// Act
	manager := NewManager(customPath)

	// Assert
	assert.NotNil(t, manager)
	assert.Equal(t, customPath, manager.configPath)
}

func TestManager_LoadConfig_WithNonExistentFile_ShouldCreateDefaultConfig(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")
	manager := NewManager(configPath)

	// Act
	config, err := manager.LoadConfig()

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, config)
	
	// Verify default values
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "info", config.Server.LogLevel)
	
	// Verify file was created
	assert.FileExists(t, configPath)
}

func TestManager_SaveConfig_WithValidConfig_ShouldWriteToFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "save-test.toml")
	manager := NewManager(configPath)
	
	config := DefaultConfig()
	config.Server.Port = 9090
	config.Settings.ActiveAPI = "test-api"

	// Act
	err := manager.SaveConfig(config)

	// Assert
	require.NoError(t, err)
	assert.FileExists(t, configPath)

	// Verify content by loading it back
	loadedConfig, err := manager.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 9090, loadedConfig.Server.Port)
	assert.Equal(t, "test-api", loadedConfig.Settings.ActiveAPI)
}

func TestManager_AddAPIConfig_WithValidAPI_ShouldAddToList(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "add-api-test.toml")
	manager := NewManager(configPath)

	api := &APIConfig{
		ID:         "test-api",
		Name:       "Test API",
		URL:        "https://api.test.com",
		APIKey:     "test-key",
		IsActive:   true,
		Timeout:    30,
		RetryCount: 3,
	}

	// Act
	err := manager.AddAPIConfig(api)

	// Assert
	require.NoError(t, err)

	// Verify API was added
	config, err := manager.LoadConfig()
	require.NoError(t, err)
	assert.Len(t, config.APIs, 1)
	assert.Equal(t, "test-api", config.APIs[0].ID)
	assert.Equal(t, "Test API", config.APIs[0].Name)
	assert.Equal(t, "https://api.test.com", config.APIs[0].URL)
}

func TestManager_AddAPIConfig_WithDuplicateID_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "duplicate-test.toml")
	manager := NewManager(configPath)

	api1 := &APIConfig{ID: "same-id", Name: "API 1", URL: "https://api1.com", APIKey: "key1"}
	api2 := &APIConfig{ID: "same-id", Name: "API 2", URL: "https://api2.com", APIKey: "key2"}

	// Act
	err1 := manager.AddAPIConfig(api1)
	err2 := manager.AddAPIConfig(api2)

	// Assert
	require.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "already exists")
}

func TestManager_RemoveAPIConfig_WithExistingID_ShouldRemoveFromList(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "remove-test.toml")
	manager := NewManager(configPath)

	// Add two APIs
	api1 := &APIConfig{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"}
	api2 := &APIConfig{ID: "api2", Name: "API 2", URL: "https://api2.com", APIKey: "key2"}
	
	require.NoError(t, manager.AddAPIConfig(api1))
	require.NoError(t, manager.AddAPIConfig(api2))

	// Act - Remove one API
	err := manager.RemoveAPIConfig("api1")

	// Assert
	require.NoError(t, err)

	// Verify API was removed
	config, err := manager.LoadConfig()
	require.NoError(t, err)
	assert.Len(t, config.APIs, 1)
	assert.Equal(t, "api2", config.APIs[0].ID)
}

func TestManager_RemoveAPIConfig_WithNonExistentID_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "remove-nonexistent-test.toml")
	manager := NewManager(configPath)

	// Act
	err := manager.RemoveAPIConfig("nonexistent-api")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_SetActiveAPI_WithExistingAPI_ShouldUpdateActiveSetting(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "set-active-test.toml")
	manager := NewManager(configPath)

	api := &APIConfig{ID: "test-api", Name: "Test API", URL: "https://api.test.com", APIKey: "key"}
	require.NoError(t, manager.AddAPIConfig(api))

	// Act
	err := manager.SetActiveAPI("test-api")

	// Assert
	require.NoError(t, err)

	// Verify active API was set
	config, err := manager.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "test-api", config.Settings.ActiveAPI)
}

func TestManager_SetActiveAPI_WithNonExistentAPI_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "set-active-invalid-test.toml")
	manager := NewManager(configPath)

	// Act
	err := manager.SetActiveAPI("nonexistent-api")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_GetActiveAPI_WithValidActiveAPI_ShouldReturnCorrectAPI(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "get-active-test.toml")
	manager := NewManager(configPath)

	api := &APIConfig{
		ID:     "active-api",
		Name:   "Active API",
		URL:    "https://active.api.com",
		APIKey: "active-key",
	}
	
	require.NoError(t, manager.AddAPIConfig(api))
	require.NoError(t, manager.SetActiveAPI("active-api"))

	// Act
	activeAPI, err := manager.GetActiveAPI()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "active-api", activeAPI.ID)
	assert.Equal(t, "Active API", activeAPI.Name)
	assert.Equal(t, "https://active.api.com", activeAPI.URL)
}

func TestManager_GetActiveAPI_WithNoActiveAPI_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no-active-test.toml")
	manager := NewManager(configPath)

	// Act
	activeAPI, err := manager.GetActiveAPI()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, activeAPI)
	assert.Contains(t, err.Error(), "no active API")
}

func TestManager_RemoveAPIConfig_WhenIsActiveAPI_ShouldClearActiveSetting(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "remove-active-test.toml")
	manager := NewManager(configPath)

	api := &APIConfig{ID: "will-be-removed", Name: "Test API", URL: "https://test.com", APIKey: "key"}
	require.NoError(t, manager.AddAPIConfig(api))
	require.NoError(t, manager.SetActiveAPI("will-be-removed"))

	// Act
	err := manager.RemoveAPIConfig("will-be-removed")

	// Assert
	require.NoError(t, err)

	// Verify active API was cleared
	config, err := manager.LoadConfig()
	require.NoError(t, err)
	assert.Empty(t, config.Settings.ActiveAPI)
}

func TestManager_LoadConfig_WithInvalidTOMLFile_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.toml")
	
	// Create an invalid TOML file
	invalidTOML := "invalid toml content [[[["
	require.NoError(t, os.WriteFile(configPath, []byte(invalidTOML), 0644))
	
	manager := NewManager(configPath)

	// Act
	config, err := manager.LoadConfig()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to decode config file")
}