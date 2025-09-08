package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnsureDefaultConfig_NoExistingConfig_ShouldCreateDefault tests creating default config when none exists
func TestEnsureDefaultConfig_NoExistingConfig_ShouldCreateDefault(t *testing.T) {
	// Arrange
	originalWd, _ := os.Getwd()
	tempDir := t.TempDir()
	defer os.Chdir(originalWd)
	
	// Change to temp directory so configs/default.toml gets created there
	os.Chdir(tempDir)

	// Act
	configPath, err := EnsureDefaultConfig()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("configs", "default.toml"), configPath)
	
	// Verify file was created
	fullPath := filepath.Join(tempDir, "configs", "default.toml")
	_, err = os.Stat(fullPath)
	assert.NoError(t, err, "Default config file should be created")
	
	// Verify directory was created
	configsDir := filepath.Join(tempDir, "configs")
	info, err := os.Stat(configsDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestEnsureDefaultConfig_ExistingConfig_ShouldReturnExisting tests behavior when default config already exists
func TestEnsureDefaultConfig_ExistingConfig_ShouldReturnExisting(t *testing.T) {
	// Arrange
	originalWd, _ := os.Getwd()
	tempDir := t.TempDir()
	defer os.Chdir(originalWd)
	
	os.Chdir(tempDir)
	
	// Create configs directory and default.toml
	configsDir := filepath.Join(tempDir, "configs")
	os.MkdirAll(configsDir, 0755)
	
	existingConfig := `[server]
port = 9999
`
	defaultConfigPath := filepath.Join(configsDir, "default.toml")
	require.NoError(t, os.WriteFile(defaultConfigPath, []byte(existingConfig), 0644))

	// Act
	configPath, err := EnsureDefaultConfig()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("configs", "default.toml"), configPath)
	
	// Verify existing config wasn't overwritten
	content, err := os.ReadFile(defaultConfigPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "port = 9999")
}

// TestValidateConfigFile_ValidFile_ShouldPass tests validation of valid config file
func TestValidateConfigFile_ValidFile_ShouldPass(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "valid.toml")
	
	validConfig := `[server]
port = 8080
log_level = "info"

[[apis]]
id = "test"
name = "Test API"
url = "https://api.test.com"
api_key = "key123"

[settings]
active_api = "test"
`
	require.NoError(t, os.WriteFile(configFile, []byte(validConfig), 0644))

	// Act
	err := ValidateConfigFile(configFile)

	// Assert
	assert.NoError(t, err)
}

// TestValidateConfigFile_EmptyPath_ShouldReturnError tests validation with empty path
func TestValidateConfigFile_EmptyPath_ShouldReturnError(t *testing.T) {
	// Act
	err := ValidateConfigFile("")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file path is empty")
}

// TestValidateConfigFile_NonExistentFile_ShouldReturnError tests validation with non-existent file
func TestValidateConfigFile_NonExistentFile_ShouldReturnError(t *testing.T) {
	// Arrange
	nonExistentFile := "/nonexistent/path/config.toml"

	// Act
	err := ValidateConfigFile(nonExistentFile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file does not exist")
}

// TestValidateConfigFile_InvalidFile_ShouldReturnError tests validation with invalid TOML
func TestValidateConfigFile_InvalidFile_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid.toml")
	
	invalidConfig := `[server
port = 8080` // Invalid TOML - missing closing bracket
	require.NoError(t, os.WriteFile(configFile, []byte(invalidConfig), 0644))

	// Act
	err := ValidateConfigFile(configFile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config file")
}

// TestResolveConfigFile_WithProvidedFile_ShouldUseProvided tests resolving with provided config file
func TestResolveConfigFile_WithProvidedFile_ShouldUseProvided(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "custom.toml")
	
	validConfig := `[server]
port = 8080

[[apis]]
id = "test"
name = "Test API"
url = "https://api.test.com"
api_key = "key123"

[settings]
active_api = "test"
`
	require.NoError(t, os.WriteFile(configFile, []byte(validConfig), 0644))

	// Create a temporary state manager
	stateManager, err := createTestStateManager(t, tempDir)
	require.NoError(t, err)

	// Act
	resolvedPath, configChanged, err := ResolveConfigFile(configFile, stateManager)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, configFile, resolvedPath)
	assert.True(t, configChanged) // Should be true since no previous config
}

// TestResolveConfigFile_NoProvidedFile_ShouldUseDefault tests resolving without provided file
func TestResolveConfigFile_NoProvidedFile_ShouldUseDefault(t *testing.T) {
	// Arrange
	originalWd, _ := os.Getwd()
	tempDir := t.TempDir()
	defer os.Chdir(originalWd)
	
	os.Chdir(tempDir)

	// Create a temporary state manager
	stateManager, err := createTestStateManager(t, tempDir)
	require.NoError(t, err)

	// Act
	resolvedPath, configChanged, err := ResolveConfigFile("", stateManager)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, resolvedPath, "default.toml")
	assert.False(t, configChanged) // No config change when using default
}

// TestResolveConfigFile_InvalidProvidedFile_ShouldReturnError tests resolving with invalid provided file
func TestResolveConfigFile_InvalidProvidedFile_ShouldReturnError(t *testing.T) {
	// Arrange
	invalidConfigFile := "/nonexistent/config.toml"
	
	tempDir := t.TempDir()
	stateManager, err := createTestStateManager(t, tempDir)
	require.NoError(t, err)

	// Act
	resolvedPath, configChanged, err := ResolveConfigFile(invalidConfigFile, stateManager)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, resolvedPath)
	assert.False(t, configChanged)
	assert.Contains(t, err.Error(), "provided config file is invalid")
}

// createTestStateManager creates a state manager for testing
func createTestStateManager(t *testing.T, baseDir string) (*Manager, error) {
	// Create a temporary settings file path
	settingsFile := filepath.Join(baseDir, "test_settings.toml")
	
	// Create a state manager with a custom settings file for testing
	manager := &Manager{
		settingsFile: settingsFile,
	}
	
	return manager, nil
}

// TestResolveConfigFile_RelativePath_ShouldConvertToAbsolute tests relative path conversion
func TestResolveConfigFile_RelativePath_ShouldConvertToAbsolute(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	// Change to temp directory
	os.Chdir(tempDir)
	
	// Create relative config file
	relativeConfigFile := "custom.toml"
	validConfig := `[server]
port = 8080

[[apis]]
id = "test"
name = "Test API"
url = "https://api.test.com"
api_key = "key123"

[settings]
active_api = "test"
`
	require.NoError(t, os.WriteFile(relativeConfigFile, []byte(validConfig), 0644))

	stateManager, err := createTestStateManager(t, tempDir)
	require.NoError(t, err)

	// Act
	resolvedPath, _, err := ResolveConfigFile(relativeConfigFile, stateManager)

	// Assert
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(resolvedPath), "Resolved path should be absolute")
	assert.Contains(t, resolvedPath, "custom.toml")
}