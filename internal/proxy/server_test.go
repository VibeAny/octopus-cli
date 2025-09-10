package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"octopus-cli/internal/config"
)

func TestNewServer_WithValidConfig_ShouldCreateServer(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:     8080,
			LogLevel: "info",
		},
	}

	// Act
	server := NewServer(cfg)

	// Assert
	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.Equal(t, 8080, server.port)
	assert.False(t, server.isRunning)
}

func TestServer_Start_WithValidPort_ShouldStartServer(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:     0, // Use random port for testing
			LogLevel: "info",
		},
	}
	server := NewServer(cfg)

	// Act
	err := server.Start()

	// Assert
	require.NoError(t, err)
	assert.True(t, server.IsRunning())

	// Cleanup
	server.Stop()
}

func TestServer_Start_WhenAlreadyRunning_ShouldReturnError(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:     0, // Use random port
			LogLevel: "info",
		},
	}
	server := NewServer(cfg)
	require.NoError(t, server.Start())

	// Act
	err := server.Start()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Cleanup
	server.Stop()
}

func TestServer_Stop_WhenRunning_ShouldStopServer(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:     0,
			LogLevel: "info",
		},
	}
	server := NewServer(cfg)
	require.NoError(t, server.Start())
	assert.True(t, server.IsRunning())

	// Act
	err := server.Stop()

	// Assert
	require.NoError(t, err)
	assert.False(t, server.IsRunning())
}

func TestServer_Stop_WhenNotRunning_ShouldReturnError(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
	}
	server := NewServer(cfg)

	// Act
	err := server.Stop()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestServer_IsRunning_InitialState_ShouldReturnFalse(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
	}
	server := NewServer(cfg)

	// Act & Assert
	assert.False(t, server.IsRunning())
}

func TestServer_UpdateConfig_WithValidConfig_ShouldUpdateConfiguration(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		APIs: []config.APIConfig{
			{ID: "api1", Name: "API 1", URL: "https://api1.com"},
		},
	}
	server := NewServer(cfg)

	newAPI := &config.APIConfig{
		ID:   "api2",
		Name: "API 2",
		URL:  "https://api2.com",
	}

	// Act
	err := server.UpdateConfig(newAPI)

	// Assert
	require.NoError(t, err)
	// Verify config was updated (implementation will define how)
}

func TestServer_GetStats_ShouldReturnServerStats(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
	}
	server := NewServer(cfg)

	// Act
	stats := server.GetStats()

	// Assert
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.RequestCount)
	assert.Equal(t, int64(0), stats.ErrorCount)
	assert.NotZero(t, stats.StartTime)
}

func TestServer_HandleRequest_WithValidTarget_ShouldForwardRequest(t *testing.T) {
	// Arrange - Create a mock target server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello from target"}`))
	}))
	defer targetServer.Close()

	// Configure proxy with the target
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 0},
		APIs: []config.APIConfig{
			{
				ID:       "target",
				Name:     "Target API",
				URL:      targetServer.URL,
				IsActive: true,
			},
		},
		Settings: config.Settings{ActiveAPI: "target"},
	}

	server := NewServer(cfg)
	require.NoError(t, server.Start())
	defer server.Stop()

	// Act - Make request to proxy
	proxyURL := fmt.Sprintf("http://localhost:%d/test", server.GetPort())
	resp, err := http.Get(proxyURL)

	// Assert
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Hello from target")

	// Verify stats were updated
	stats := server.GetStats()
	assert.Greater(t, stats.RequestCount, int64(0))
}

func TestServer_HandleRequest_WithNoActiveAPI_ShouldReturnError(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server:   config.ServerConfig{Port: 0},
		APIs:     []config.APIConfig{},
		Settings: config.Settings{ActiveAPI: ""},
	}

	server := NewServer(cfg)
	require.NoError(t, server.Start())
	defer server.Stop()

	// Act
	proxyURL := fmt.Sprintf("http://localhost:%d/test", server.GetPort())
	resp, err := http.Get(proxyURL)

	// Assert
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "no active API")
}

func TestServer_HandleRequest_WithInvalidTarget_ShouldReturnError(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 0},
		APIs: []config.APIConfig{
			{
				ID:       "invalid",
				Name:     "Invalid API",
				URL:      "http://invalid-host-that-does-not-exist:9999",
				IsActive: true,
			},
		},
		Settings: config.Settings{ActiveAPI: "invalid"},
	}

	server := NewServer(cfg)
	require.NoError(t, server.Start())
	defer server.Stop()

	// Act
	proxyURL := fmt.Sprintf("http://localhost:%d/test", server.GetPort())
	resp, err := http.Get(proxyURL)

	// Assert
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)

	// Verify error stats were updated
	stats := server.GetStats()
	assert.Greater(t, stats.ErrorCount, int64(0))
}

func TestServer_HandleRequest_WithPOSTMethod_ShouldForwardBody(t *testing.T) {
	// Arrange - Target server that echoes request body
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"received": "%s"}`, string(body))))
	}))
	defer targetServer.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{Port: 0},
		APIs: []config.APIConfig{
			{ID: "target", URL: targetServer.URL, IsActive: true},
		},
		Settings: config.Settings{ActiveAPI: "target"},
	}

	server := NewServer(cfg)
	require.NoError(t, server.Start())
	defer server.Stop()

	// Act - Make POST request
	requestBody := `{"test": "data"}`
	proxyURL := fmt.Sprintf("http://localhost:%d/api/test", server.GetPort())
	resp, err := http.Post(proxyURL, "application/json", strings.NewReader(requestBody))

	// Assert
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"test": "data"`)
}

func TestServer_HandleRequest_ShouldPreserveHeaders(t *testing.T) {
	// Arrange - Target server that returns received headers
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received-Auth", r.Header.Get("Authorization"))
		w.Header().Set("X-Received-Custom", r.Header.Get("X-Custom-Header"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer targetServer.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{Port: 0},
		APIs: []config.APIConfig{
			{ID: "target", URL: targetServer.URL, IsActive: true},
		},
		Settings: config.Settings{ActiveAPI: "target"},
	}

	server := NewServer(cfg)
	require.NoError(t, server.Start())
	defer server.Stop()

	// Act - Make request with custom headers
	client := &http.Client{}
	proxyURL := fmt.Sprintf("http://localhost:%d/test", server.GetPort())
	req, _ := http.NewRequest("GET", proxyURL, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Custom-Header", "custom-value")

	resp, err := client.Do(req)

	// Assert
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer test-token", resp.Header.Get("X-Received-Auth"))
	assert.Equal(t, "custom-value", resp.Header.Get("X-Received-Custom"))
}

func TestServer_Graceful_Shutdown_ShouldCompleteActiveRequests(t *testing.T) {
	// Arrange - Slow target server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "completed"}`))
	}))
	defer targetServer.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{Port: 0},
		APIs: []config.APIConfig{
			{ID: "target", URL: targetServer.URL, IsActive: true},
		},
		Settings: config.Settings{ActiveAPI: "target"},
	}

	server := NewServer(cfg)
	require.NoError(t, server.Start())

	// Act - Start request and shutdown concurrently
	responseCh := make(chan *http.Response, 1)
	errorCh := make(chan error, 1)

	go func() {
		proxyURL := fmt.Sprintf("http://localhost:%d/test", server.GetPort())
		resp, err := http.Get(proxyURL)
		if err != nil {
			errorCh <- err
			return
		}
		responseCh <- resp
	}()

	// Give request time to start, then shutdown
	time.Sleep(50 * time.Millisecond)
	shutdownErr := server.Stop()

	// Assert
	require.NoError(t, shutdownErr)

	// Verify request completed successfully
	select {
	case resp := <-responseCh:
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	case err := <-errorCh:
		t.Fatalf("Request failed: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Request did not complete within timeout")
	}
}
