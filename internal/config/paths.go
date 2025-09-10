package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// PathManager handles all application paths
type PathManager struct {
	homeDir string
	appDir  string
}

// NewPathManager creates a new path manager
func NewPathManager() *PathManager {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	var appDir string
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\Octopus
		if appData := os.Getenv("APPDATA"); appData != "" {
			appDir = filepath.Join(appData, "Octopus")
		} else {
			appDir = filepath.Join(home, ".octopus")
		}
	case "darwin":
		// macOS: ~/Library/Application Support/Octopus
		appDir = filepath.Join(home, "Library", "Application Support", "Octopus")
	default:
		// Linux/Unix: ~/.octopus
		appDir = filepath.Join(home, ".octopus")
	}

	return &PathManager{
		homeDir: home,
		appDir:  appDir,
	}
}

// AppDir returns the main application directory
func (pm *PathManager) AppDir() string {
	return pm.appDir
}

// ConfigFile returns the main configuration file path
func (pm *PathManager) ConfigFile() string {
	return filepath.Join(pm.appDir, "octopus.toml")
}

// ConfigDir returns the configuration directory path
func (pm *PathManager) ConfigDir() string {
	return filepath.Join(pm.appDir, "configs")
}

// LogsDir returns the logs directory path
func (pm *PathManager) LogsDir() string {
	return filepath.Join(pm.appDir, "logs")
}

// LogFile returns the main log file path
func (pm *PathManager) LogFile() string {
	return filepath.Join(pm.LogsDir(), "octopus.log")
}

// PIDFile returns the PID file path
func (pm *PathManager) PIDFile() string {
	return filepath.Join(pm.appDir, "octopus.pid")
}

// StateFile returns the state file path
func (pm *PathManager) StateFile() string {
	return filepath.Join(pm.appDir, "state.json")
}

// EnsureDirs creates all necessary directories
func (pm *PathManager) EnsureDirs() error {
	dirs := []string{
		pm.appDir,
		pm.ConfigDir(),
		pm.LogsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetDefaultPathManager returns the default path manager instance
func GetDefaultPathManager() *PathManager {
	return NewPathManager()
}
