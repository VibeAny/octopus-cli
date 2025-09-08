package process

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// ProcessStatus represents the status of the daemon process
type ProcessStatus struct {
	IsRunning bool
	PID       int
	Uptime    time.Duration
	StartTime time.Time
}

// Manager handles process lifecycle management
type Manager struct {
	pidFile string
	name    string
}

// NewManager creates a new process manager
func NewManager(pidFile, name string) *Manager {
	// Convert relative paths to absolute paths based on executable directory
	if !filepath.IsAbs(pidFile) {
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			pidFile = filepath.Join(execDir, pidFile)
		}
	}
	
	return &Manager{
		pidFile: pidFile,
		name:    name,
	}
}

// StartDaemon starts the service as a daemon process
func (m *Manager) StartDaemon() error {
	// Check if already running
	if status, _ := m.GetDaemonStatus(); status != nil && status.IsRunning {
		return fmt.Errorf("daemon is already running with PID %d", status.PID)
	}

	// Create PID file with current process ID
	pid := os.Getpid()
	if err := m.writePIDFile(pid); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// WritePIDFile writes the PID to file (public method)
func (m *Manager) WritePIDFile(pid int) error {
	return m.writePIDFile(pid)
}

// StopDaemon stops the running daemon process
func (m *Manager) StopDaemon() error {
	status, err := m.GetDaemonStatus()
	if err != nil {
		return fmt.Errorf("failed to get daemon status: %w", err)
	}

	if !status.IsRunning {
		return fmt.Errorf("daemon is not running")
	}

	// Send SIGTERM to the process
	process, err := os.FindProcess(status.PID)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", status.PID, err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", status.PID, err)
	}

	// Wait for graceful shutdown, then cleanup
	time.Sleep(100 * time.Millisecond)
	return m.CleanupPIDFile()
}

// GetDaemonStatus returns the current status of the daemon
func (m *Manager) GetDaemonStatus() (*ProcessStatus, error) {
	// Read PID from file
	pid, err := m.readPIDFile()
	if err != nil {
		// If PID file doesn't exist, daemon is not running
		return &ProcessStatus{IsRunning: false}, nil
	}

	// Check if process is actually running
	process, err := os.FindProcess(pid)
	if err != nil {
		// Process not found, cleanup stale PID file
		m.CleanupPIDFile()
		return &ProcessStatus{IsRunning: false}, nil
	}

	// Try to send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// Process doesn't exist, cleanup stale PID file
		m.CleanupPIDFile()
		return &ProcessStatus{IsRunning: false}, nil
	}

	// TODO: Get actual start time and calculate uptime
	return &ProcessStatus{
		IsRunning: true,
		PID:       pid,
		StartTime: time.Now(), // Placeholder
		Uptime:    time.Hour,  // Placeholder
	}, nil
}

// SendSignal sends a signal to the daemon process
func (m *Manager) SendSignal(signal os.Signal) error {
	status, err := m.GetDaemonStatus()
	if err != nil {
		return err
	}

	if !status.IsRunning {
		return fmt.Errorf("daemon is not running")
	}

	process, err := os.FindProcess(status.PID)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", status.PID, err)
	}

	return process.Signal(signal)
}

// CleanupPIDFile removes the PID file
func (m *Manager) CleanupPIDFile() error {
	return os.Remove(m.pidFile)
}

// SetupSignalHandling sets up graceful shutdown on signals
func (m *Manager) SetupSignalHandling(cleanup func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		if cleanup != nil {
			cleanup()
		}
		m.CleanupPIDFile()
		os.Exit(0)
	}()
}

// readPIDFile reads the PID from the PID file
func (m *Manager) readPIDFile() (int, error) {
	data, err := os.ReadFile(m.pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %w", err)
	}

	return pid, nil
}

// writePIDFile writes the PID to the PID file
func (m *Manager) writePIDFile(pid int) error {
	return os.WriteFile(m.pidFile, []byte(strconv.Itoa(pid)), 0644)
}