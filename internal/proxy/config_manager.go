package proxy

import (
	"fmt"
	"sync"

	"octopus-cli/internal/config"
)

// ConfigManager handles dynamic configuration management for the proxy
type ConfigManager struct {
	mu     sync.RWMutex
	config *config.Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(cfg *config.Config) *ConfigManager {
	return &ConfigManager{
		config: cfg,
	}
}

// GetActiveAPIID returns the ID of the currently active API
func (cm *ConfigManager) GetActiveAPIID() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Settings.ActiveAPI
}

// GetActiveAPI returns the currently active API configuration
func (cm *ConfigManager) GetActiveAPI() (*config.APIConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	activeID := cm.config.Settings.ActiveAPI
	if activeID == "" {
		return nil, fmt.Errorf("no active API configured")
	}

	for _, api := range cm.config.APIs {
		if api.ID == activeID {
			// Return a copy to prevent external modification
			apiCopy := api
			return &apiCopy, nil
		}
	}

	return nil, fmt.Errorf("active API '%s' not found", activeID)
}

// SwitchAPI switches to a different API configuration
func (cm *ConfigManager) SwitchAPI(apiID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Verify the API exists
	found := false
	for _, api := range cm.config.APIs {
		if api.ID == apiID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("API not found: %s", apiID)
	}

	cm.config.Settings.ActiveAPI = apiID
	return nil
}

// AddAPI adds a new API configuration
func (cm *ConfigManager) AddAPI(apiConfig config.APIConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check for duplicate ID
	for _, existingAPI := range cm.config.APIs {
		if existingAPI.ID == apiConfig.ID {
			return fmt.Errorf("API with ID '%s' already exists", apiConfig.ID)
		}
	}

	cm.config.APIs = append(cm.config.APIs, apiConfig)
	return nil
}

// RemoveAPI removes an API configuration
func (cm *ConfigManager) RemoveAPI(apiID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Find and remove the API
	found := false
	newAPIs := make([]config.APIConfig, 0, len(cm.config.APIs))
	
	for _, api := range cm.config.APIs {
		if api.ID == apiID {
			found = true
			// If removing the active API, clear the active setting
			if cm.config.Settings.ActiveAPI == apiID {
				cm.config.Settings.ActiveAPI = ""
			}
		} else {
			newAPIs = append(newAPIs, api)
		}
	}

	if !found {
		return fmt.Errorf("API with ID '%s' not found", apiID)
	}

	cm.config.APIs = newAPIs
	return nil
}

// GetAllAPIs returns all configured APIs
func (cm *ConfigManager) GetAllAPIs() []config.APIConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy to prevent external modification
	apis := make([]config.APIConfig, len(cm.config.APIs))
	copy(apis, cm.config.APIs)
	return apis
}

// ReloadConfig reloads the configuration with new settings
func (cm *ConfigManager) ReloadConfig(newConfig *config.Config) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate that the active API exists in the new configuration
	if newConfig.Settings.ActiveAPI != "" {
		found := false
		for _, api := range newConfig.APIs {
			if api.ID == newConfig.Settings.ActiveAPI {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("active API '%s' not found in new configuration", newConfig.Settings.ActiveAPI)
		}
	}

	cm.config = newConfig
	return nil
}

// GetConfig returns a copy of the current configuration
func (cm *ConfigManager) GetConfig() *config.Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create a deep copy of the configuration
	configCopy := *cm.config
	
	// Copy the APIs slice
	configCopy.APIs = make([]config.APIConfig, len(cm.config.APIs))
	copy(configCopy.APIs, cm.config.APIs)
	
	return &configCopy
}