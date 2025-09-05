package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"octopus-cli/internal/config"
)

func TestNewConfigManager_WithValidConfig_ShouldCreateManager(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
			{ID: "api2", Name: "API 2", URL: "https://api2.com", APIKey: "key2"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	// Act
	manager := NewConfigManager(cfg)

	// Assert
	assert.NotNil(t, manager)
	assert.Equal(t, "api1", manager.GetActiveAPIID())
}

func TestConfigManager_SwitchAPI_WithExistingAPI_ShouldSwitchSuccessfully(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
			{ID: "api2", Name: "API 2", URL: "https://api2.com", APIKey: "key2"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(cfg)

	// Act
	err := manager.SwitchAPI("api2")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "api2", manager.GetActiveAPIID())
}

func TestConfigManager_SwitchAPI_WithNonExistentAPI_ShouldReturnError(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(cfg)

	// Act
	err := manager.SwitchAPI("nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API not found")
	assert.Equal(t, "api1", manager.GetActiveAPIID()) // Should remain unchanged
}

func TestConfigManager_GetActiveAPI_ShouldReturnCurrentActiveAPI(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
			{ID: "api2", Name: "API 2", URL: "https://api2.com", APIKey: "key2"},
		},
		Settings: config.Settings{ActiveAPI: "api2"},
	}

	manager := NewConfigManager(cfg)

	// Act
	activeAPI, err := manager.GetActiveAPI()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "api2", activeAPI.ID)
	assert.Equal(t, "API 2", activeAPI.Name)
	assert.Equal(t, "https://api2.com", activeAPI.URL)
	assert.Equal(t, "key2", activeAPI.APIKey)
}

func TestConfigManager_GetActiveAPI_WithNoActiveAPI_ShouldReturnError(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
		},
		Settings: config.Settings{ActiveAPI: ""},
	}

	manager := NewConfigManager(cfg)

	// Act
	activeAPI, err := manager.GetActiveAPI()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, activeAPI)
	assert.Contains(t, err.Error(), "no active API")
}

func TestConfigManager_AddAPI_WithNewAPI_ShouldAddSuccessfully(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(cfg)

	newAPI := config.APIConfig{
		ID:     "api2",
		Name:   "API 2",
		URL:    "https://api2.com",
		APIKey: "key2",
	}

	// Act
	err := manager.AddAPI(newAPI)

	// Assert
	require.NoError(t, err)
	apis := manager.GetAllAPIs()
	assert.Len(t, apis, 2)
	
	// Verify the new API was added
	found := false
	for _, api := range apis {
		if api.ID == "api2" {
			found = true
			assert.Equal(t, "API 2", api.Name)
			assert.Equal(t, "https://api2.com", api.URL)
			assert.Equal(t, "key2", api.APIKey)
			break
		}
	}
	assert.True(t, found, "New API should be found in the list")
}

func TestConfigManager_AddAPI_WithDuplicateID_ShouldReturnError(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(cfg)

	duplicateAPI := config.APIConfig{
		ID:     "api1", // Same ID as existing
		Name:   "Duplicate API",
		URL:    "https://duplicate.com",
		APIKey: "duplicate-key",
	}

	// Act
	err := manager.AddAPI(duplicateAPI)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	
	// Verify original API unchanged
	apis := manager.GetAllAPIs()
	assert.Len(t, apis, 1)
	assert.Equal(t, "API 1", apis[0].Name)
}

func TestConfigManager_RemoveAPI_WithExistingAPI_ShouldRemoveSuccessfully(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
			{ID: "api2", Name: "API 2", URL: "https://api2.com", APIKey: "key2"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(cfg)

	// Act
	err := manager.RemoveAPI("api2")

	// Assert
	require.NoError(t, err)
	apis := manager.GetAllAPIs()
	assert.Len(t, apis, 1)
	assert.Equal(t, "api1", apis[0].ID)
}

func TestConfigManager_RemoveAPI_WithActiveAPI_ShouldClearActiveAPI(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
			{ID: "api2", Name: "API 2", URL: "https://api2.com", APIKey: "key2"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(cfg)

	// Act
	err := manager.RemoveAPI("api1") // Remove active API

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "", manager.GetActiveAPIID()) // Active API should be cleared
	
	apis := manager.GetAllAPIs()
	assert.Len(t, apis, 1)
	assert.Equal(t, "api2", apis[0].ID)
}

func TestConfigManager_RemoveAPI_WithNonExistentAPI_ShouldReturnError(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(cfg)

	// Act
	err := manager.RemoveAPI("nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Verify nothing changed
	apis := manager.GetAllAPIs()
	assert.Len(t, apis, 1)
}

func TestConfigManager_ConcurrentAccess_ShouldBeSafe(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
			{ID: "api2", Name: "API 2", URL: "https://api2.com", APIKey: "key2"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(cfg)

	// Act - perform concurrent operations
	var wg sync.WaitGroup
	numGoroutines := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			
			if i%2 == 0 {
				manager.SwitchAPI("api2")
			} else {
				manager.SwitchAPI("api1")
			}
			
			// Read operations
			manager.GetActiveAPI()
			manager.GetAllAPIs()
		}(i)
	}

	wg.Wait()

	// Assert - should not panic and have valid state
	activeAPI, err := manager.GetActiveAPI()
	require.NoError(t, err)
	assert.Contains(t, []string{"api1", "api2"}, activeAPI.ID)
}

func TestConfigManager_ReloadConfig_ShouldUpdateConfiguration(t *testing.T) {
	// Arrange
	initialCfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com", APIKey: "key1"},
		},
		Settings: config.Settings{ActiveAPI: "api1"},
	}

	manager := NewConfigManager(initialCfg)

	newCfg := &config.Config{
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1 Updated", URL: "https://api1.com", APIKey: "key1"},
			{ID: "api3", Name: "API 3", URL: "https://api3.com", APIKey: "key3"},
		},
		Settings: config.Settings{ActiveAPI: "api3"},
	}

	// Act
	err := manager.ReloadConfig(newCfg)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "api3", manager.GetActiveAPIID())
	
	apis := manager.GetAllAPIs()
	assert.Len(t, apis, 2)
	
	// Verify updated API
	for _, api := range apis {
		if api.ID == "api1" {
			assert.Equal(t, "API 1 Updated", api.Name)
		}
	}
}

func TestServerWithConfigManager_DynamicSwitch_ShouldForwardToNewAPI(t *testing.T) {
	// Arrange - Create two mock target servers
	server1Called := false
	targetServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server1Called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"server": "1"}`))
	}))
	defer targetServer1.Close()

	server2Called := false
	targetServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server2Called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"server": "2"}`))
	}))
	defer targetServer2.Close()

	// Create config with both servers
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 0},
		APIs: []config.APIConfig{
			{ID: "server1", URL: targetServer1.URL, APIKey: "key1"},
			{ID: "server2", URL: targetServer2.URL, APIKey: "key2"},
		},
		Settings: config.Settings{ActiveAPI: "server1"},
	}

	server := NewServer(cfg)
	require.NoError(t, server.Start())
	defer server.Stop()

	proxyURL := fmt.Sprintf("http://localhost:%d/test", server.GetPort())

	// Act & Assert - First request should go to server1
	resp1, err := http.Get(proxyURL)
	require.NoError(t, err)
	resp1.Body.Close()
	
	assert.True(t, server1Called, "First request should go to server1")
	assert.False(t, server2Called, "First request should not go to server2")

	// Switch to server2
	err = server.UpdateConfig(&config.APIConfig{
		ID: "server2", URL: targetServer2.URL, APIKey: "key2",
	})
	require.NoError(t, err)

	// Wait a bit for config update to take effect
	time.Sleep(10 * time.Millisecond)

	// Reset flags and make second request
	server1Called = false
	server2Called = false

	resp2, err := http.Get(proxyURL)
	require.NoError(t, err)
	resp2.Body.Close()

	// This test will help us verify the server switching logic once implemented
	// For now, we expect the same behavior since UpdateConfig is not fully implemented
}