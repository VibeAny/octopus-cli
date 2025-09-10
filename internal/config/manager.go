package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Manager handles configuration operations
type Manager struct {
	configPath string
	config     *Config
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	if configPath == "" {
		configPath = defaultConfigPath()
	}

	return &Manager{
		configPath: configPath,
		config:     DefaultConfig(),
	}
}

// LoadConfig loads configuration from TOML file
func (m *Manager) LoadConfig() (*Config, error) {
	// Create config directory if it doesn't exist
	if err := m.ensureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// If config file doesn't exist, create it with default values
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		if err := m.SaveConfig(m.config); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return m.config, nil
	}

	// Load existing config
	if _, err := toml.DecodeFile(m.configPath, m.config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return m.config, nil
}

// SaveConfig saves configuration to TOML file
func (m *Manager) SaveConfig(config *Config) error {
	if err := m.ensureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	m.config = config
	return nil
}

// AddAPIConfig adds a new API configuration
func (m *Manager) AddAPIConfig(api *APIConfig) error {
	if m.config == nil {
		if _, err := m.LoadConfig(); err != nil {
			return err
		}
	}

	// Check for duplicate IDs
	for _, existing := range m.config.APIs {
		if existing.ID == api.ID {
			return fmt.Errorf("API configuration with ID '%s' already exists", api.ID)
		}
	}

	m.config.APIs = append(m.config.APIs, *api)
	return m.SaveConfig(m.config)
}

// RemoveAPIConfig removes an API configuration by ID
func (m *Manager) RemoveAPIConfig(id string) error {
	if m.config == nil {
		if _, err := m.LoadConfig(); err != nil {
			return err
		}
	}

	for i, api := range m.config.APIs {
		if api.ID == id {
			// Remove the API from slice
			m.config.APIs = append(m.config.APIs[:i], m.config.APIs[i+1:]...)

			// If this was the active API, clear the active setting
			if m.config.Settings.ActiveAPI == id {
				m.config.Settings.ActiveAPI = ""
			}

			return m.SaveConfig(m.config)
		}
	}

	return fmt.Errorf("API configuration with ID '%s' not found", id)
}

// SetActiveAPI sets the active API configuration
func (m *Manager) SetActiveAPI(id string) error {
	if m.config == nil {
		if _, err := m.LoadConfig(); err != nil {
			return err
		}
	}

	// Verify the API exists
	found := false
	for _, api := range m.config.APIs {
		if api.ID == id {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("API configuration with ID '%s' not found", id)
	}

	m.config.Settings.ActiveAPI = id
	return m.SaveConfig(m.config)
}

// GetActiveAPI returns the currently active API configuration
func (m *Manager) GetActiveAPI() (*APIConfig, error) {
	if m.config == nil {
		if _, err := m.LoadConfig(); err != nil {
			return nil, err
		}
	}

	activeID := m.config.Settings.ActiveAPI
	if activeID == "" {
		return nil, fmt.Errorf("no active API configuration set")
	}

	for _, api := range m.config.APIs {
		if api.ID == activeID {
			return &api, nil
		}
	}

	return nil, fmt.Errorf("active API configuration '%s' not found", activeID)
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// defaultConfigPath returns the default configuration file path
func defaultConfigPath() string {
	pm := GetDefaultPathManager()
	return pm.ConfigFile()
}

// ensureConfigDir creates the configuration directory if it doesn't exist
func (m *Manager) ensureConfigDir() error {
	pm := GetDefaultPathManager()
	return pm.EnsureDirs()
}
