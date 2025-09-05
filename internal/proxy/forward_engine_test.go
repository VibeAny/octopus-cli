package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"octopus-cli/internal/config"
)

func TestNewForwardEngine_WithValidConfig_ShouldCreateEngine(t *testing.T) {
	// Arrange
	apiConfig := &config.APIConfig{
		ID:         "test-api",
		Name:       "Test API",
		URL:        "https://api.test.com",
		APIKey:     "test-key",
		Timeout:    30,
		RetryCount: 3,
	}

	// Act
	engine := NewForwardEngine(apiConfig)

	// Assert
	assert.NotNil(t, engine)
	assert.Equal(t, apiConfig, engine.apiConfig)
	assert.NotNil(t, engine.client)
	assert.Equal(t, time.Duration(30)*time.Second, engine.timeout)
	assert.Equal(t, 3, engine.retryCount)
}

func TestForwardEngine_ForwardRequest_WithValidTarget_ShouldSucceed(t *testing.T) {
	// Arrange - Create a mock target server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer targetServer.Close()

	apiConfig := &config.APIConfig{
		ID:         "test-api",
		URL:        targetServer.URL,
		APIKey:     "test-key",
		Timeout:    5,
		RetryCount: 1,
	}

	engine := NewForwardEngine(apiConfig)

	// Create test request
	req := httptest.NewRequest("POST", "/api/test", strings.NewReader(`{"test": "data"}`))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := engine.ForwardRequest(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestForwardEngine_ForwardRequest_WithRetry_ShouldRetryOnFailure(t *testing.T) {
	// Arrange - Create a server that fails first two times then succeeds
	callCount := 0
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success after retry"}`))
	}))
	defer targetServer.Close()

	apiConfig := &config.APIConfig{
		ID:         "test-api",
		URL:        targetServer.URL,
		Timeout:    5,
		RetryCount: 3,
	}

	engine := NewForwardEngine(apiConfig)

	// Create test request
	req := httptest.NewRequest("GET", "/api/test", nil)

	// Act
	resp, err := engine.ForwardRequest(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, callCount, "Should have made exactly 3 attempts")
}

func TestForwardEngine_ForwardRequest_WithAllRetriesFailed_ShouldReturnError(t *testing.T) {
	// Arrange - Create a server that always fails
	callCount := 0
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer targetServer.Close()

	apiConfig := &config.APIConfig{
		ID:         "test-api",
		URL:        targetServer.URL,
		Timeout:    5,
		RetryCount: 2,
	}

	engine := NewForwardEngine(apiConfig)

	// Create test request
	req := httptest.NewRequest("GET", "/api/test", nil)

	// Act
	resp, err := engine.ForwardRequest(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "max retries exceeded")
	assert.Equal(t, 2, callCount, "Should have made exactly 2 retry attempts")
}

func TestForwardEngine_ForwardRequest_WithTimeout_ShouldTimeout(t *testing.T) {
	// Arrange - Create a slow server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Sleep longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer targetServer.Close()

	apiConfig := &config.APIConfig{
		ID:         "test-api",
		URL:        targetServer.URL,
		Timeout:    1, // 1 second timeout
		RetryCount: 1,
	}

	engine := NewForwardEngine(apiConfig)

	// Create test request with short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/api/test", nil)

	// Act
	resp, err := engine.ForwardRequest(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestForwardEngine_ForwardRequest_WithNetworkError_ShouldRetryAndFail(t *testing.T) {
	// Arrange - Use invalid URL to simulate network error
	apiConfig := &config.APIConfig{
		ID:         "test-api",
		URL:        "http://invalid-host-that-does-not-exist:9999",
		Timeout:    5,
		RetryCount: 2,
	}

	engine := NewForwardEngine(apiConfig)

	// Create test request
	req := httptest.NewRequest("GET", "/api/test", nil)

	// Act
	resp, err := engine.ForwardRequest(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "max retries exceeded")
}

func TestForwardEngine_ShouldRetry_WithRetryableStatusCodes_ShouldReturnTrue(t *testing.T) {
	// Arrange
	apiConfig := &config.APIConfig{RetryCount: 3}
	engine := NewForwardEngine(apiConfig)

	retryableCodes := []int{
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,         // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout,     // 504
	}

	for _, code := range retryableCodes {
		t.Run(string(rune(code)), func(t *testing.T) {
			// Act & Assert
			assert.True(t, engine.shouldRetry(code, nil), 
				"Status code %d should be retryable", code)
		})
	}
}

func TestForwardEngine_ShouldRetry_WithNonRetryableStatusCodes_ShouldReturnFalse(t *testing.T) {
	// Arrange
	apiConfig := &config.APIConfig{RetryCount: 3}
	engine := NewForwardEngine(apiConfig)

	nonRetryableCodes := []int{
		http.StatusOK,                  // 200
		http.StatusBadRequest,          // 400
		http.StatusUnauthorized,        // 401
		http.StatusForbidden,           // 403
		http.StatusNotFound,            // 404
		http.StatusMethodNotAllowed,    // 405
	}

	for _, code := range nonRetryableCodes {
		t.Run(string(rune(code)), func(t *testing.T) {
			// Act & Assert
			assert.False(t, engine.shouldRetry(code, nil),
				"Status code %d should not be retryable", code)
		})
	}
}

func TestForwardEngine_ShouldRetry_WithNetworkErrors_ShouldReturnTrue(t *testing.T) {
	// Arrange
	apiConfig := &config.APIConfig{RetryCount: 3}
	engine := NewForwardEngine(apiConfig)

	networkErrors := []error{
		errors.New("dial tcp: connection refused"),
		errors.New("dial tcp: i/o timeout"),
		errors.New("read tcp: connection reset by peer"),
		context.DeadlineExceeded,
	}

	for _, err := range networkErrors {
		t.Run(err.Error(), func(t *testing.T) {
			// Act & Assert
			assert.True(t, engine.shouldRetry(0, err),
				"Network error should be retryable: %v", err)
		})
	}
}

func TestForwardEngine_ForwardRequest_WithCustomHeaders_ShouldPreserveHeaders(t *testing.T) {
	// Arrange
	receivedHeaders := make(http.Header)
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Store received headers for verification
		for name, values := range r.Header {
			receivedHeaders[name] = values
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer targetServer.Close()

	apiConfig := &config.APIConfig{
		ID:         "test-api",
		URL:        targetServer.URL,
		APIKey:     "test-key",
		Timeout:    5,
		RetryCount: 1,
	}

	engine := NewForwardEngine(apiConfig)

	// Create request with custom headers
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("User-Agent", "octopus-cli/1.0")

	// Act
	resp, err := engine.ForwardRequest(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Verify headers were preserved
	assert.Equal(t, "Bearer test-key", receivedHeaders.Get("Authorization"))
	assert.Equal(t, "custom-value", receivedHeaders.Get("X-Custom-Header"))
	assert.Equal(t, "octopus-cli/1.0", receivedHeaders.Get("User-Agent"))
}

func TestForwardEngine_GetStats_ShouldReturnStatistics(t *testing.T) {
	// Arrange
	apiConfig := &config.APIConfig{
		ID:         "test-api",
		URL:        "https://api.test.com",
		Timeout:    5,
		RetryCount: 1,
	}

	engine := NewForwardEngine(apiConfig)

	// Act
	stats := engine.GetStats()

	// Assert
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.SuccessfulRequests)
	assert.Equal(t, int64(0), stats.FailedRequests)
	assert.Equal(t, int64(0), stats.TotalRetries)
	assert.NotZero(t, stats.StartTime)
}