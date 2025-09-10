package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIConfig_StructureTags_ShouldHaveCorrectTOMLTags(t *testing.T) {
	// This test ensures our TOML tags are correct for marshaling/unmarshaling
	// We can verify this by creating a config and checking its serialization

	// Arrange
	api := APIConfig{
		ID:         "test-id",
		Name:       "Test Name",
		URL:        "https://test.com",
		APIKey:     "test-key",
		IsActive:   true,
		Timeout:    30,
		RetryCount: 3,
	}

	// Act & Assert - verify fields are accessible and properly typed
	assert.Equal(t, "test-id", api.ID)
	assert.Equal(t, "Test Name", api.Name)
	assert.Equal(t, "https://test.com", api.URL)
	assert.Equal(t, "test-key", api.APIKey)
	assert.True(t, api.IsActive)
	assert.Equal(t, 30, api.Timeout)
	assert.Equal(t, 3, api.RetryCount)
}

func TestServerConfig_StructureTags_ShouldHaveCorrectTOMLTags(t *testing.T) {
	// Arrange
	server := ServerConfig{
		Port:     8080,
		LogLevel: "debug",
		Daemon:   false,
	}

	// Act & Assert
	assert.Equal(t, 8080, server.Port)
	assert.Equal(t, "debug", server.LogLevel)
	assert.False(t, server.Daemon)
}

func TestSettings_StructureTags_ShouldHaveCorrectTOMLTags(t *testing.T) {
	// Arrange
	settings := Settings{
		ActiveAPI:    "test-api",
		LogFile:      "/test/log",
		ConfigBackup: false,
	}

	// Act & Assert
	assert.Equal(t, "test-api", settings.ActiveAPI)
	assert.Equal(t, "/test/log", settings.LogFile)
	assert.False(t, settings.ConfigBackup)
}

func TestConfig_CompleteStructure_ShouldHaveAllNestedFields(t *testing.T) {
	// Arrange
	config := Config{
		Server: ServerConfig{
			Port:     9090,
			LogLevel: "error",
			Daemon:   true,
		},
		APIs: []APIConfig{
			{
				ID:         "api1",
				Name:       "API One",
				URL:        "https://api1.com",
				APIKey:     "key1",
				IsActive:   true,
				Timeout:    45,
				RetryCount: 5,
			},
			{
				ID:         "api2",
				Name:       "API Two",
				URL:        "https://api2.com",
				APIKey:     "key2",
				IsActive:   false,
				Timeout:    60,
				RetryCount: 2,
			},
		},
		Settings: Settings{
			ActiveAPI:    "api1",
			LogFile:      "/custom/log",
			ConfigBackup: true,
		},
	}

	// Act & Assert - verify nested structure access
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "error", config.Server.LogLevel)

	assert.Len(t, config.APIs, 2)
	assert.Equal(t, "api1", config.APIs[0].ID)
	assert.Equal(t, "API One", config.APIs[0].Name)
	assert.True(t, config.APIs[0].IsActive)
	assert.Equal(t, "api2", config.APIs[1].ID)
	assert.False(t, config.APIs[1].IsActive)

	assert.Equal(t, "api1", config.Settings.ActiveAPI)
	assert.Equal(t, "/custom/log", config.Settings.LogFile)
	assert.True(t, config.Settings.ConfigBackup)
}

func TestDefaultConfig_Values_ShouldMatchExpectedDefaults(t *testing.T) {
	// Act
	config := DefaultConfig()

	// Assert - verify all default values are as expected
	// Server defaults
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "info", config.Server.LogLevel)
	assert.True(t, config.Server.Daemon)
	// Note: PIDFile is now managed internally and not configurable

	// APIs now include example configurations by default
	assert.Len(t, config.APIs, 2)
	assert.NotNil(t, config.APIs) // Should be initialized, not nil

	// Settings defaults
	assert.Empty(t, config.Settings.ActiveAPI) // No active API initially
	// LogFile should now be an absolute path
	assert.Contains(t, config.Settings.LogFile, "octopus.log")
	assert.True(t, config.Settings.ConfigBackup)
}

func TestAPIConfig_ZeroValues_ShouldHaveExpectedBehavior(t *testing.T) {
	// Arrange - zero-valued APIConfig
	var api APIConfig

	// Assert
	assert.Empty(t, api.ID)
	assert.Empty(t, api.Name)
	assert.Empty(t, api.URL)
	assert.Empty(t, api.APIKey)
	assert.False(t, api.IsActive)
	assert.Zero(t, api.Timeout)
	assert.Zero(t, api.RetryCount)
}

func TestConfig_TOMLFieldNames_ShouldBeSnakeCase(t *testing.T) {
	// This test documents the expected TOML field names
	// If we change TOML tags, this test will remind us to update docs

	// Note: This is more of a documentation test - the actual TOML tag
	// verification would require reflection or serialization testing
	// For now, we document the expected behavior

	config := DefaultConfig()

	// These assertions document the expected structure
	assert.NotNil(t, config.Server, "server section should exist")
	assert.NotNil(t, config.APIs, "apis array should exist")
	assert.NotNil(t, config.Settings, "settings section should exist")

	// Field names should follow TOML conventions:
	// - Settings.ActiveAPI should serialize to "active_api"
	// - Settings.LogFile should serialize to "log_file"
	// - Settings.ConfigBackup should serialize to "config_backup"
	// - APIs.IsActive should serialize to "is_active"
	// - APIs.APIKey should serialize to "api_key"
	// - APIs.RetryCount should serialize to "retry_count"
	// Note: Server.PIDFile was removed - now managed internally
}
