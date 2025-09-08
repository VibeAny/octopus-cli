package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager_ShouldCreateManagerWithCorrectPath(t *testing.T) {
	manager, err := NewManager()
	
	require.NoError(t, err)
	assert.NotNil(t, manager)
	
	// The implementation uses executable directory + "settings.toml"
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	expectedPath := filepath.Join(execDir, "settings.toml")
	assert.Equal(t, expectedPath, manager.GetSettingsFile())
}

func TestManager_LoadSettings_WithNoFile_ShouldReturnEmptySettings(t *testing.T) {
	// Create temporary settings file path
	tempDir := t.TempDir()
	settingsFile := filepath.Join(tempDir, "settings.toml")
	
	manager := &Manager{settingsFile: settingsFile}
	
	settings, err := manager.LoadSettings()
	
	require.NoError(t, err)
	assert.Equal(t, "", settings.CurrentConfigFile)
}

func TestManager_SaveAndLoadSettings_ShouldPersistData(t *testing.T) {
	// Create temporary settings file path
	tempDir := t.TempDir()
	settingsFile := filepath.Join(tempDir, "settings.toml")
	
	manager := &Manager{settingsFile: settingsFile}
	
	// Save settings
	originalSettings := &Settings{
		CurrentConfigFile: "/path/to/config.toml",
	}
	
	err := manager.SaveSettings(originalSettings)
	require.NoError(t, err)
	
	// Load settings
	loadedSettings, err := manager.LoadSettings()
	require.NoError(t, err)
	
	assert.Equal(t, originalSettings.CurrentConfigFile, loadedSettings.CurrentConfigFile)
}

func TestManager_SetCurrentConfigFile_ShouldSaveAndRetrieve(t *testing.T) {
	// Create temporary settings file path
	tempDir := t.TempDir()
	settingsFile := filepath.Join(tempDir, "settings.toml")
	
	manager := &Manager{settingsFile: settingsFile}
	
	configFile := "/path/to/test-config.toml"
	
	// Set config file
	err := manager.SetCurrentConfigFile(configFile)
	require.NoError(t, err)
	
	// Get config file
	retrievedConfigFile, err := manager.GetCurrentConfigFile()
	require.NoError(t, err)
	
	assert.Equal(t, configFile, retrievedConfigFile)
}

func TestManager_ClearCurrentConfigFile_ShouldClearSetting(t *testing.T) {
	// Create temporary settings file path
	tempDir := t.TempDir()
	settingsFile := filepath.Join(tempDir, "settings.toml")
	
	manager := &Manager{settingsFile: settingsFile}
	
	// Set config file
	err := manager.SetCurrentConfigFile("/path/to/config.toml")
	require.NoError(t, err)
	
	// Clear config file
	err = manager.ClearCurrentConfigFile()
	require.NoError(t, err)
	
	// Verify it's cleared
	configFile, err := manager.GetCurrentConfigFile()
	require.NoError(t, err)
	assert.Equal(t, "", configFile)
}

func TestEnsureDefaultConfig_WithExistingConfig_ShouldReturnPath(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	// Change to temp directory
	err := os.Chdir(tempDir)
	require.NoError(t, err)
	
	// Create configs directory and default.toml
	configsDir := filepath.Join(tempDir, "configs")
	err = os.MkdirAll(configsDir, 0755)
	require.NoError(t, err)
	
	defaultConfigPath := filepath.Join(configsDir, "default.toml")
	err = os.WriteFile(defaultConfigPath, []byte(`
[server]
port = 8080

[settings]
active_api = ""
`), 0644)
	require.NoError(t, err)
	
	// Test function
	resultPath, err := EnsureDefaultConfig()
	require.NoError(t, err)
	
	expectedPath := filepath.Join("configs", "default.toml")
	assert.Equal(t, expectedPath, resultPath)
}

func TestValidateConfigFile_WithNonExistentFile_ShouldReturnError(t *testing.T) {
	err := ValidateConfigFile("/non/existent/config.toml")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file does not exist")
}

func TestValidateConfigFile_WithEmptyPath_ShouldReturnError(t *testing.T) {
	err := ValidateConfigFile("")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file path is empty")
}

func TestResolveConfigFile_WithProvidedConfig_ShouldUseProvided(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	// Change to temp directory
	err := os.Chdir(tempDir)
	require.NoError(t, err)
	
	// Create temporary state manager
	settingsFile := filepath.Join(tempDir, "settings.toml")
	stateManager := &Manager{settingsFile: settingsFile}
	
	// Create a test config file
	configsDir := filepath.Join(tempDir, "configs")
	err = os.MkdirAll(configsDir, 0755)
	require.NoError(t, err)
	
	testConfigPath := filepath.Join(configsDir, "test.toml")
	err = os.WriteFile(testConfigPath, []byte(`
[server]
port = 8080

[settings]
active_api = ""
`), 0644)
	require.NoError(t, err)
	
	// Test function
	providedConfig := filepath.Join("configs", "test.toml")
	resolvedConfig, configChanged, err := ResolveConfigFile(providedConfig, stateManager)
	
	require.NoError(t, err)
	assert.True(t, configChanged)
	assert.True(t, filepath.IsAbs(resolvedConfig))
	assert.Contains(t, resolvedConfig, "test.toml")
}