package state

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Settings represents the application state settings
type Settings struct {
	CurrentConfigFile string `toml:"current_config_file"`
}

// Manager manages application state persistence
type Manager struct {
	settingsFile string
}

// NewManager creates a new state manager
func NewManager() (*Manager, error) {
	// Get the directory where the binary is located
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	
	execDir := filepath.Dir(execPath)
	settingsFile := filepath.Join(execDir, "settings.toml")

	return &Manager{
		settingsFile: settingsFile,
	}, nil
}

// NewManagerWithSettingsFile creates a new state manager with a specific settings file
// This is primarily for testing purposes
func NewManagerWithSettingsFile(settingsFile string) *Manager {
	return &Manager{
		settingsFile: settingsFile,
	}
}

// LoadSettings loads the current settings
func (m *Manager) LoadSettings() (*Settings, error) {
	settings := &Settings{}

	// If settings file doesn't exist, return default settings
	if _, err := os.Stat(m.settingsFile); os.IsNotExist(err) {
		return settings, nil
	}

	// Load settings from file
	if _, err := toml.DecodeFile(m.settingsFile, settings); err != nil {
		return nil, fmt.Errorf("failed to decode settings file: %w", err)
	}

	return settings, nil
}

// SaveSettings saves the current settings
func (m *Manager) SaveSettings(settings *Settings) error {
	// Create or overwrite settings file
	file, err := os.Create(m.settingsFile)
	if err != nil {
		return fmt.Errorf("failed to create settings file: %w", err)
	}
	defer file.Close()

	// Encode settings to TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(settings); err != nil {
		return fmt.Errorf("failed to encode settings: %w", err)
	}

	return nil
}

// SetCurrentConfigFile sets and saves the current config file path
func (m *Manager) SetCurrentConfigFile(configFile string) error {
	settings, err := m.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	settings.CurrentConfigFile = configFile

	if err := m.SaveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// GetCurrentConfigFile gets the current config file path
func (m *Manager) GetCurrentConfigFile() (string, error) {
	settings, err := m.LoadSettings()
	if err != nil {
		return "", fmt.Errorf("failed to load settings: %w", err)
	}

	return settings.CurrentConfigFile, nil
}

// ClearCurrentConfigFile clears the current config file setting
func (m *Manager) ClearCurrentConfigFile() error {
	settings, err := m.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	settings.CurrentConfigFile = ""

	if err := m.SaveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// GetSettingsFile returns the path to the settings file
func (m *Manager) GetSettingsFile() string {
	return m.settingsFile
}