package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"octopus-cli/internal/config"
)

// ForwardEngineStats represents statistics for the forward engine
type ForwardEngineStats struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalRetries       int64
	StartTime          time.Time
}

// ForwardEngine handles API request forwarding with retry logic
type ForwardEngine struct {
	apiConfig        *config.APIConfig
	client           *http.Client
	timeout          time.Duration
	retryCount       int
	totalRequests    int64
	successfulReqs   int64
	failedReqs       int64
	totalRetries     int64
	startTime        time.Time
}

// NewForwardEngine creates a new forward engine
func NewForwardEngine(apiConfig *config.APIConfig) *ForwardEngine {
	timeout := time.Duration(apiConfig.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second // default timeout
	}

	return &ForwardEngine{
		apiConfig:  apiConfig,
		timeout:    timeout,
		retryCount: apiConfig.RetryCount,
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				Proxy: nil, // Disable proxy to avoid interference
			},
		},
		startTime: time.Now(),
	}
}

// ForwardRequest forwards a request to the target API with retry logic
func (f *ForwardEngine) ForwardRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	atomic.AddInt64(&f.totalRequests, 1)

	// Create target URL
	targetURL := f.apiConfig.URL + req.URL.Path
	if req.URL.RawQuery != "" {
		targetURL += "?" + req.URL.RawQuery
	}

	var lastErr error
	for attempt := 0; attempt < f.retryCount; attempt++ {
		if attempt > 0 {
			atomic.AddInt64(&f.totalRetries, 1)
			// Add exponential backoff delay
			delay := time.Duration(attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				atomic.AddInt64(&f.failedReqs, 1)
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Create new request for this attempt
		targetReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, req.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// Copy headers from original request
		for name, values := range req.Header {
			for _, value := range values {
				targetReq.Header.Add(name, value)
			}
		}

		// Add API key if present
		if f.apiConfig.APIKey != "" {
			targetReq.Header.Set("Authorization", "Bearer "+f.apiConfig.APIKey)
		}

		// Make the request
		resp, err := f.client.Do(targetReq)
		if err != nil {
			lastErr = err
			if f.shouldRetry(0, err) {
				continue
			}
			break
		}

		// Check if we should retry based on status code
		if f.shouldRetry(resp.StatusCode, nil) {
			resp.Body.Close()
			lastErr = fmt.Errorf("received retryable status code: %d", resp.StatusCode)
			continue
		}

		// Success
		atomic.AddInt64(&f.successfulReqs, 1)
		return resp, nil
	}

	// All retries exhausted
	atomic.AddInt64(&f.failedReqs, 1)
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// shouldRetry determines if a request should be retried based on status code or error
func (f *ForwardEngine) shouldRetry(statusCode int, err error) bool {
	// Retry on network errors
	if err != nil {
		errStr := err.Error()
		networkErrors := []string{
			"connection refused",
			"connection reset",
			"i/o timeout",
			"deadline exceeded",
			"no such host",
			"network unreachable",
		}
		
		for _, netErr := range networkErrors {
			if strings.Contains(errStr, netErr) {
				return true
			}
		}
		return false
	}

	// Retry on specific HTTP status codes
	retryableStatusCodes := []int{
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,         // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout,     // 504
	}

	for _, code := range retryableStatusCodes {
		if statusCode == code {
			return true
		}
	}

	return false
}

// GetStats returns current statistics
func (f *ForwardEngine) GetStats() *ForwardEngineStats {
	return &ForwardEngineStats{
		TotalRequests:      atomic.LoadInt64(&f.totalRequests),
		SuccessfulRequests: atomic.LoadInt64(&f.successfulReqs),
		FailedRequests:     atomic.LoadInt64(&f.failedReqs),
		TotalRetries:       atomic.LoadInt64(&f.totalRetries),
		StartTime:          f.startTime,
	}
}