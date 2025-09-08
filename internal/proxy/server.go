package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"octopus-cli/internal/config"
	"octopus-cli/internal/utils"
)

// ServerStats represents server statistics
type ServerStats struct {
	RequestCount int64
	ErrorCount   int64
	StartTime    time.Time
	Uptime       time.Duration
}

// Server represents the HTTP proxy server
type Server struct {
	config       *config.Config
	port         int
	actualPort   int
	isRunning    bool
	server       *http.Server
	listener     net.Listener
	stats        *ServerStats
	logger       *utils.Logger
	mu           sync.RWMutex
	requestCount int64
	errorCount   int64
}

// NewServer creates a new proxy server
func NewServer(cfg *config.Config) *Server {
	// Initialize logger
	var logger *utils.Logger
	if cfg.Settings.LogFile != "" {
		if l, err := utils.NewLogger(cfg.Settings.LogFile); err == nil {
			logger = l
		}
	}

	return &Server{
		config: cfg,
		port:   cfg.Server.Port,
		logger: logger,
		stats: &ServerStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the HTTP proxy server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	s.listener = listener
	s.actualPort = listener.Addr().(*net.TCPAddr).Port

	// Log server startup
	if s.logger != nil {
		s.logger.Info("Starting Octopus proxy server on port %d", s.actualPort)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	s.server = &http.Server{
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			if s.logger != nil {
				s.logger.Error("Server error: %v", err)
			}
		}
	}()

	s.isRunning = true
	
	if s.logger != nil {
		s.logger.Info("Octopus proxy server started successfully on port %d", s.actualPort)
	}
	
	return nil
}

// Stop stops the HTTP proxy server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("server is not running")
	}

	if s.logger != nil {
		s.logger.Info("Stopping Octopus proxy server...")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to shutdown server gracefully: %v", err)
		}
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.isRunning = false
	
	if s.logger != nil {
		s.logger.Info("Octopus proxy server stopped successfully")
	}
	
	return nil
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetPort returns the port the server is listening on
func (s *Server) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.isRunning && s.actualPort > 0 {
		return s.actualPort
	}
	return s.port
}

// UpdateConfig updates the server configuration
func (s *Server) UpdateConfig(apiConfig *config.APIConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Implement config update logic
	// For now, just accept any config update
	return nil
}

// GetStats returns current server statistics
func (s *Server) GetStats() *ServerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := *s.stats
	stats.RequestCount = atomic.LoadInt64(&s.requestCount)
	stats.ErrorCount = atomic.LoadInt64(&s.errorCount)
	stats.Uptime = time.Since(s.stats.StartTime)
	return &stats
}

// handleRequest handles incoming HTTP requests and forwards them
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&s.requestCount, 1)
	
	// Log incoming request
	if s.logger != nil {
		s.logger.Info("Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	}

	// Get active API configuration
	activeAPI, err := s.getActiveAPI()
	if err != nil {
		atomic.AddInt64(&s.errorCount, 1)
		if s.logger != nil {
			s.logger.Error("No active API configured: %v", err)
		}
		http.Error(w, fmt.Sprintf("no active API configured: %v", err), http.StatusBadGateway)
		return
	}

	// Log API forwarding
	if s.logger != nil {
		s.logger.Info("Forwarding request to API: %s (%s)", activeAPI.ID, activeAPI.URL)
	}

	// Forward the request
	if err := s.forwardRequest(w, r, activeAPI); err != nil {
		atomic.AddInt64(&s.errorCount, 1)
		if s.logger != nil {
			s.logger.Error("Failed to forward request to %s: %v", activeAPI.URL, err)
		}
		http.Error(w, fmt.Sprintf("failed to forward request: %v", err), http.StatusBadGateway)
		return
	}
	
	// Log successful forwarding
	if s.logger != nil {
		s.logger.Info("Request forwarded successfully to %s", activeAPI.ID)
	}
}

// getActiveAPI returns the currently active API configuration
func (s *Server) getActiveAPI() (*config.APIConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config.Settings.ActiveAPI == "" {
		return nil, fmt.Errorf("no active API")
	}

	for _, api := range s.config.APIs {
		if api.ID == s.config.Settings.ActiveAPI {
			return &api, nil
		}
	}

	return nil, fmt.Errorf("active API '%s' not found", s.config.Settings.ActiveAPI)
}

// forwardRequest forwards the request to the target API
func (s *Server) forwardRequest(w http.ResponseWriter, r *http.Request, api *config.APIConfig) error {
	// Parse target URL
	targetURL, err := url.Parse(api.URL)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}

	// Create target request
	targetURL.Path = r.URL.Path
	targetURL.RawQuery = r.URL.RawQuery

	// Configure timeout
	timeout := time.Duration(api.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second // default timeout
	}

	// Create request with timeout context
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	targetReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL.String(), r.Body)
	if err != nil {
		return fmt.Errorf("failed to create target request: %w", err)
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			targetReq.Header.Add(name, value)
		}
	}

	// Add API key if present
	if api.APIKey != "" {
		targetReq.Header.Set("Authorization", "Bearer "+api.APIKey)
	}

	// Make request to target (using the timeout from context)
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: nil, // Disable proxy to get direct connection errors
		},
	}

	resp, err := client.Do(targetReq)
	if err != nil {
		// Don't write anything to response here - let handleRequest do it
		return fmt.Errorf("request to target failed: %w", err)
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		// Response already started writing, can't change status code now
		return fmt.Errorf("failed to copy response body: %w", err)
	}

	return nil
}