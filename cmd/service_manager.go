package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	
	"octopus-cli/internal/config"
	"octopus-cli/internal/process"
	"octopus-cli/internal/proxy"
)

// ServiceManager manages the lifecycle of the Octopus proxy service
type ServiceManager struct {
	configManager  *config.Manager
	processManager *process.Manager
	proxyServer    *proxy.Server
	configFile     string
}

// NewServiceManager creates a new service manager
func NewServiceManager(configFile string) (*ServiceManager, error) {
	// Load configuration
	configManager := config.NewManager(configFile)
	cfg, err := configManager.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create process manager
	processManager := process.NewManager(cfg.Server.PIDFile, "octopus")

	// Create proxy server
	proxyServer := proxy.NewServer(cfg)

	return &ServiceManager{
		configManager:  configManager,
		processManager: processManager,
		proxyServer:    proxyServer,
		configFile:     configFile,
	}, nil
}

// Start starts the proxy service as a daemon
func (sm *ServiceManager) Start() error {
	// Check if already running
	status, err := sm.processManager.GetDaemonStatus()
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if status.IsRunning {
		return fmt.Errorf("service is already running with PID %d", status.PID)
	}

	// Fork a daemon process
	if err := sm.forkDaemon(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	return nil
}

// forkDaemon creates a daemon process
func (sm *ServiceManager) forkDaemon() error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Convert config file to absolute path
	configFile := sm.configFile
	if !filepath.IsAbs(configFile) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		configFile = filepath.Join(wd, configFile)
	}

	// Prepare daemon command
	args := []string{
		"--daemon-mode",
		"--config", configFile,
	}

	// Create the daemon process
	cmd := exec.Command(execPath, args...)
	cmd.Env = os.Environ()
	cmd.Dir = "/"
	
	// Redirect outputs to devnull for true daemon behavior
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open devnull: %w", err)
	}
	defer devNull.Close()
	
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	// Start the daemon process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// Write the PID file
	if err := sm.processManager.WritePIDFile(cmd.Process.Pid); err != nil {
		// Kill the process if we can't write PID file
		cmd.Process.Kill()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Detach from the daemon process
	cmd.Process.Release()

	return nil
}

// Stop stops the proxy service
func (sm *ServiceManager) Stop() error {
	// Check if running
	status, err := sm.processManager.GetDaemonStatus()
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if !status.IsRunning {
		return fmt.Errorf("service is not running")
	}

	// Stop the proxy server if it's the current process
	if sm.proxyServer.IsRunning() {
		if err := sm.proxyServer.Stop(); err != nil {
			return fmt.Errorf("failed to stop proxy server: %w", err)
		}
	}

	// Stop the daemon process
	if err := sm.processManager.StopDaemon(); err != nil {
		return fmt.Errorf("failed to stop daemon process: %w", err)
	}

	return nil
}

// Status returns the current service status
func (sm *ServiceManager) Status() (*ServiceStatus, error) {
	cfg, err := sm.configManager.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	processStatus, err := sm.processManager.GetDaemonStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get process status: %w", err)
	}

	var proxyStats *proxy.ServerStats
	if sm.proxyServer.IsRunning() {
		proxyStats = sm.proxyServer.GetStats()
	}

	return &ServiceStatus{
		IsRunning:    processStatus.IsRunning,
		PID:          processStatus.PID,
		Port:         cfg.Server.Port,
		ActiveAPI:    cfg.Settings.ActiveAPI,
		StartTime:    processStatus.StartTime,
		Uptime:       processStatus.Uptime,
		ProxyStats:   proxyStats,
	}, nil
}

// ServiceStatus represents the current status of the service
type ServiceStatus struct {
	IsRunning  bool
	PID        int
	Port       int
	ActiveAPI  string
	StartTime  interface{}
	Uptime     interface{}
	ProxyStats *proxy.ServerStats
}