package process

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager_WithValidParameters_ShouldCreateManager(t *testing.T) {
	// Arrange
	pidFile := "/tmp/test.pid"
	name := "test-daemon"

	// Act
	manager := NewManager(pidFile, name)

	// Assert
	assert.NotNil(t, manager)
	assert.Equal(t, pidFile, manager.pidFile)
	assert.Equal(t, name, manager.name)
}

func TestManager_GetDaemonStatus_WithNoPIDFile_ShouldReturnNotRunning(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "nonexistent.pid")
	manager := NewManager(pidFile, "test")

	// Act
	status, err := manager.GetDaemonStatus()

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.IsRunning)
	assert.Zero(t, status.PID)
}

func TestManager_StartDaemon_WithNoPreviousProcess_ShouldCreatePIDFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "start-test.pid")
	manager := NewManager(pidFile, "test")

	// Act
	err := manager.StartDaemon()

	// Assert
	require.NoError(t, err)
	assert.FileExists(t, pidFile)

	// Verify PID file contains current process PID
	pidData, err := os.ReadFile(pidFile)
	require.NoError(t, err)
	
	writtenPID, err := strconv.Atoi(string(pidData))
	require.NoError(t, err)
	assert.Equal(t, os.Getpid(), writtenPID)

	// Cleanup
	manager.CleanupPIDFile()
}

func TestManager_StartDaemon_WithExistingRunningProcess_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "duplicate-start-test.pid")
	manager := NewManager(pidFile, "test")

	// Start daemon first time
	require.NoError(t, manager.StartDaemon())

	// Act - try to start again
	err := manager.StartDaemon()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Cleanup
	manager.CleanupPIDFile()
}

func TestManager_GetDaemonStatus_WithValidRunningProcess_ShouldReturnRunningStatus(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "status-test.pid")
	manager := NewManager(pidFile, "test")

	// Start daemon first
	require.NoError(t, manager.StartDaemon())

	// Act
	status, err := manager.GetDaemonStatus()

	// Assert
	require.NoError(t, err)
	assert.True(t, status.IsRunning)
	assert.Equal(t, os.Getpid(), status.PID)
	assert.NotZero(t, status.StartTime)

	// Cleanup
	manager.CleanupPIDFile()
}

func TestManager_GetDaemonStatus_WithStalePIDFile_ShouldCleanupAndReturnNotRunning(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "stale-pid-test.pid")
	manager := NewManager(pidFile, "test")

	// Create a stale PID file with non-existent PID
	stalePID := 999999 // Assuming this PID doesn't exist
	require.NoError(t, os.WriteFile(pidFile, []byte(strconv.Itoa(stalePID)), 0644))

	// Act
	status, err := manager.GetDaemonStatus()

	// Assert
	require.NoError(t, err)
	assert.False(t, status.IsRunning)
	assert.Zero(t, status.PID)

	// Verify stale PID file was cleaned up
	assert.NoFileExists(t, pidFile)
}

func TestManager_CleanupPIDFile_WithExistingFile_ShouldRemoveFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "cleanup-test.pid")
	manager := NewManager(pidFile, "test")

	// Create PID file
	require.NoError(t, os.WriteFile(pidFile, []byte("12345"), 0644))
	assert.FileExists(t, pidFile)

	// Act
	err := manager.CleanupPIDFile()

	// Assert
	require.NoError(t, err)
	assert.NoFileExists(t, pidFile)
}

func TestManager_CleanupPIDFile_WithNonExistentFile_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "nonexistent-cleanup.pid")
	manager := NewManager(pidFile, "test")

	// Act
	err := manager.CleanupPIDFile()

	// Assert
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestManager_SendSignal_WithNoRunningProcess_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "signal-no-process-test.pid")
	manager := NewManager(pidFile, "test")

	// Act
	err := manager.SendSignal(syscall.SIGTERM)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestManager_SendSignal_WithRunningProcess_ShouldSendSignalSuccessfully(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "signal-test.pid")
	manager := NewManager(pidFile, "test")

	// Start daemon
	require.NoError(t, manager.StartDaemon())

	// Act - send signal 0 (test signal that doesn't affect the process)
	err := manager.SendSignal(syscall.Signal(0))

	// Assert
	require.NoError(t, err)

	// Cleanup
	manager.CleanupPIDFile()
}

func TestManager_readPIDFile_WithValidPIDFile_ShouldReturnCorrectPID(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "read-pid-test.pid")
	manager := NewManager(pidFile, "test")

	expectedPID := 12345
	require.NoError(t, os.WriteFile(pidFile, []byte(strconv.Itoa(expectedPID)), 0644))

	// Act
	pid, err := manager.readPIDFile()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedPID, pid)
}

func TestManager_readPIDFile_WithInvalidPIDContent_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "invalid-pid-test.pid")
	manager := NewManager(pidFile, "test")

	require.NoError(t, os.WriteFile(pidFile, []byte("not-a-number"), 0644))

	// Act
	pid, err := manager.readPIDFile()

	// Assert
	assert.Error(t, err)
	assert.Zero(t, pid)
	assert.Contains(t, err.Error(), "invalid PID")
}

func TestManager_readPIDFile_WithNonExistentFile_ShouldReturnError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "nonexistent-read.pid")
	manager := NewManager(pidFile, "test")

	// Act
	pid, err := manager.readPIDFile()

	// Assert
	assert.Error(t, err)
	assert.Zero(t, pid)
	assert.True(t, os.IsNotExist(err))
}

func TestManager_writePIDFile_WithValidPID_ShouldCreateFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "write-pid-test.pid")
	manager := NewManager(pidFile, "test")

	testPID := 54321

	// Act
	err := manager.writePIDFile(testPID)

	// Assert
	require.NoError(t, err)
	assert.FileExists(t, pidFile)

	// Verify content
	content, err := os.ReadFile(pidFile)
	require.NoError(t, err)
	assert.Equal(t, strconv.Itoa(testPID), string(content))
}

func TestProcessStatus_ZeroValue_ShouldHaveExpectedDefaults(t *testing.T) {
	// Arrange
	var status ProcessStatus

	// Assert
	assert.False(t, status.IsRunning)
	assert.Zero(t, status.PID)
	assert.Zero(t, status.Uptime)
	assert.True(t, status.StartTime.IsZero())
}

func TestManager_SetupSignalHandling_ShouldNotPanic(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "signal-setup-test.pid")
	manager := NewManager(pidFile, "test")

	cleanup := func() {
		// Cleanup function for signal handling
	}

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		manager.SetupSignalHandling(cleanup)
	})

	// Note: We can't easily test the actual signal handling without
	// sending real signals to the test process, which could be flaky
}