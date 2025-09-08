package state

import (
	"fmt"
	"os"
	"path/filepath"

	"octopus-cli/internal/config"
)

// EnsureDefaultConfig ensures the default config file exists
func EnsureDefaultConfig() (string, error) {
	defaultConfigPath := filepath.Join("configs", "default.toml")
	
	// Check if default config exists
	if _, err := os.Stat(defaultConfigPath); err == nil {
		return defaultConfigPath, nil
	}

	// Create configs directory if it doesn't exist
	if err := os.MkdirAll("configs", 0755); err != nil {
		return "", fmt.Errorf("failed to create configs directory: %w", err)
	}

	// Create default configuration
	defaultConfig := config.DefaultConfig()
	
	// Create config manager to save the default config
	configManager := config.NewManager(defaultConfigPath)
	if err := configManager.SaveConfig(defaultConfig); err != nil {
		return "", fmt.Errorf("failed to create default config file: %w", err)
	}

	return defaultConfigPath, nil
}

// ValidateConfigFile checks if a config file exists and is readable
func ValidateConfigFile(configFile string) error {
	if configFile == "" {
		return fmt.Errorf("config file path is empty")
	}

	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configFile)
	}

	// Try to load the config to validate it
	configManager := config.NewManager(configFile)
	if _, err := configManager.LoadConfig(); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	return nil
}

// ResolveConfigFile resolves the config file to use based on parameters and state
func ResolveConfigFile(providedConfigFile string, stateManager *Manager) (string, bool, error) {
	var configFile string
	var configChanged bool

	if providedConfigFile != "" {
		// User provided -f parameter
		// Convert to absolute path
		if !filepath.IsAbs(providedConfigFile) {
			wd, err := os.Getwd()
			if err != nil {
				return "", false, fmt.Errorf("failed to get working directory: %w", err)
			}
			providedConfigFile = filepath.Join(wd, providedConfigFile)
		}

		// Validate the provided config file
		if err := ValidateConfigFile(providedConfigFile); err != nil {
			return "", false, fmt.Errorf("provided config file is invalid: %w", err)
		}

		// Check if this is different from current config
		currentConfig, err := stateManager.GetCurrentConfigFile()
		if err != nil {
			return "", false, fmt.Errorf("failed to get current config: %w", err)
		}

		configChanged = (currentConfig != providedConfigFile)
		configFile = providedConfigFile

		// Save the new config file as current
		if err := stateManager.SetCurrentConfigFile(configFile); err != nil {
			return "", false, fmt.Errorf("failed to save current config: %w", err)
		}
	} else {
		// No -f parameter, use saved config or default
		currentConfig, err := stateManager.GetCurrentConfigFile()
		if err != nil {
			return "", false, fmt.Errorf("failed to get current config: %w", err)
		}

		if currentConfig != "" {
			// Use saved config file
			// Check if the saved config file still exists
			if err := ValidateConfigFile(currentConfig); err != nil {
				// Config file was deleted or corrupted, fall back to default
				defaultConfig, err := EnsureDefaultConfig()
				if err != nil {
					return "", false, fmt.Errorf("failed to create default config: %w", err)
				}
				
				// Convert to absolute path
				if !filepath.IsAbs(defaultConfig) {
					wd, err := os.Getwd()
					if err != nil {
						return "", false, fmt.Errorf("failed to get working directory: %w", err)
					}
					defaultConfig = filepath.Join(wd, defaultConfig)
				}
				
				configFile = defaultConfig
				// Update saved config to default
				if err := stateManager.SetCurrentConfigFile(configFile); err != nil {
					return "", false, fmt.Errorf("failed to save default config: %w", err)
				}
			} else {
				configFile = currentConfig
			}
		} else {
			// No saved config, use default
			defaultConfig, err := EnsureDefaultConfig()
			if err != nil {
				return "", false, fmt.Errorf("failed to create default config: %w", err)
			}
			
			// Convert to absolute path
			if !filepath.IsAbs(defaultConfig) {
				wd, err := os.Getwd()
				if err != nil {
					return "", false, fmt.Errorf("failed to get working directory: %w", err)
				}
				defaultConfig = filepath.Join(wd, defaultConfig)
			}
			
			configFile = defaultConfig
			// Save default as current config
			if err := stateManager.SetCurrentConfigFile(configFile); err != nil {
				return "", false, fmt.Errorf("failed to save default config: %w", err)
			}
		}
	}

	return configFile, configChanged, nil
}