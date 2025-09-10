package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPathManager(t *testing.T) {
	pm := NewPathManager()

	// Test that path manager is created successfully
	assert.NotNil(t, pm)
	assert.NotEmpty(t, pm.AppDir())
}

func TestPathManager_CrossPlatform(t *testing.T) {
	pm := NewPathManager()

	// Test configuration paths
	configFile := pm.ConfigFile()
	assert.Contains(t, configFile, "octopus.toml")
	assert.True(t, filepath.IsAbs(configFile))

	// Test log paths
	logFile := pm.LogFile()
	assert.Contains(t, logFile, "octopus.log")
	assert.True(t, filepath.IsAbs(logFile))

	// Test platform-specific app directories
	appDir := pm.AppDir()
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			assert.Contains(t, appDir, "Octopus")
		} else {
			assert.Contains(t, appDir, ".octopus")
		}
	case "darwin":
		assert.Contains(t, appDir, "Library/Application Support/Octopus")
	default:
		assert.Contains(t, appDir, ".octopus")
	}
}

func TestPathManager_EnsureDirs(t *testing.T) {
	// Use a temporary directory for testing
	tempDir := t.TempDir()

	// Create a custom path manager with temp directory
	pm := &PathManager{
		homeDir: tempDir,
		appDir:  filepath.Join(tempDir, ".octopus"),
	}

	// Test directory creation
	err := pm.EnsureDirs()
	assert.NoError(t, err)

	// Verify directories exist
	assert.DirExists(t, pm.AppDir())
	assert.DirExists(t, pm.ConfigDir())
	assert.DirExists(t, pm.LogsDir())
}

func TestPathManager_AllPaths(t *testing.T) {
	pm := NewPathManager()

	// Test all path methods return non-empty strings
	assert.NotEmpty(t, pm.AppDir())
	assert.NotEmpty(t, pm.ConfigFile())
	assert.NotEmpty(t, pm.ConfigDir())
	assert.NotEmpty(t, pm.LogsDir())
	assert.NotEmpty(t, pm.LogFile())
	assert.NotEmpty(t, pm.PIDFile())
	assert.NotEmpty(t, pm.StateFile())

	// Test all paths are absolute
	assert.True(t, filepath.IsAbs(pm.AppDir()))
	assert.True(t, filepath.IsAbs(pm.ConfigFile()))
	assert.True(t, filepath.IsAbs(pm.ConfigDir()))
	assert.True(t, filepath.IsAbs(pm.LogsDir()))
	assert.True(t, filepath.IsAbs(pm.LogFile()))
	assert.True(t, filepath.IsAbs(pm.PIDFile()))
	assert.True(t, filepath.IsAbs(pm.StateFile()))
}

func TestGetDefaultPathManager(t *testing.T) {
	pm1 := GetDefaultPathManager()
	pm2 := GetDefaultPathManager()

	// Both should return valid path managers
	assert.NotNil(t, pm1)
	assert.NotNil(t, pm2)

	// Both should have the same configuration
	assert.Equal(t, pm1.AppDir(), pm2.AppDir())
	assert.Equal(t, pm1.ConfigFile(), pm2.ConfigFile())
}
