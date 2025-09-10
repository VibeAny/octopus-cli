package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLogger_ValidPath_ShouldCreateLogger tests logger creation with valid path
func TestNewLogger_ValidPath_ShouldCreateLogger(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	// Act
	logger, err := NewLogger(logFile)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.Logger)
	assert.Equal(t, logFile, logger.filePath)

	// Verify log file was created
	_, err = os.Stat(logFile)
	assert.NoError(t, err, "Log file should be created")
}

// TestNewLogger_NestedDirectory_ShouldCreateDirectories tests logger creation with nested directories
func TestNewLogger_NestedDirectory_ShouldCreateDirectories(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "logs", "subdir", "test.log")

	// Act
	logger, err := NewLogger(logFile)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, logger)

	// Verify directories were created
	logsDir := filepath.Join(tempDir, "logs", "subdir")
	info, err := os.Stat(logsDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir(), "Nested directories should be created")

	// Verify log file was created
	_, err = os.Stat(logFile)
	assert.NoError(t, err, "Log file should be created")
}

// TestNewLogger_RelativePath_ShouldConvertToAbsolute tests relative path handling
func TestNewLogger_RelativePath_ShouldConvertToAbsolute(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory to test relative path resolution
	os.Chdir(tempDir)

	relativeLogFile := "logs/app.log"

	// Act
	logger, err := NewLogger(relativeLogFile)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, logger)

	// The actual path should be based on executable directory, not working directory
	// In test environment, this will be resolved relative to the test binary location
	assert.True(t, filepath.IsAbs(logger.filePath), "Logger should store absolute path")
	assert.Contains(t, logger.filePath, "logs/app.log")
}

// TestLogger_Info_ShouldWriteInfoMessage tests Info logging
func TestLogger_Info_ShouldWriteInfoMessage(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "info.log")

	logger, err := NewLogger(logFile)
	require.NoError(t, err)

	// Act
	logger.Info("Test info message with %s", "parameter")

	// Assert
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "[INFO]")
	assert.Contains(t, logContent, "Test info message with parameter")
}

// TestLogger_Error_ShouldWriteErrorMessage tests Error logging
func TestLogger_Error_ShouldWriteErrorMessage(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "error.log")

	logger, err := NewLogger(logFile)
	require.NoError(t, err)

	// Act
	logger.Error("Test error message: %v", "error details")

	// Assert
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "[ERROR]")
	assert.Contains(t, logContent, "Test error message: error details")
}

// TestLogger_Warn_ShouldWriteWarnMessage tests Warn logging
func TestLogger_Warn_ShouldWriteWarnMessage(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "warn.log")

	logger, err := NewLogger(logFile)
	require.NoError(t, err)

	// Act
	logger.Warn("Test warning message")

	// Assert
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "[WARN]")
	assert.Contains(t, logContent, "Test warning message")
}

// TestLogger_Debug_ShouldWriteDebugMessage tests Debug logging
func TestLogger_Debug_ShouldWriteDebugMessage(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "debug.log")

	logger, err := NewLogger(logFile)
	require.NoError(t, err)

	// Act
	logger.Debug("Debug info: %d", 42)

	// Assert
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "[DEBUG]")
	assert.Contains(t, logContent, "Debug info: 42")
}

// TestLogger_MultipleMessages_ShouldAppend tests that multiple messages are appended
func TestLogger_MultipleMessages_ShouldAppend(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "multi.log")

	logger, err := NewLogger(logFile)
	require.NoError(t, err)

	// Act
	logger.Info("First message")
	logger.Error("Second message")
	logger.Warn("Third message")

	// Assert
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)
	lines := strings.Split(strings.TrimSpace(logContent), "\n")

	assert.Len(t, lines, 3, "Should have 3 log lines")
	assert.Contains(t, lines[0], "[INFO] First message")
	assert.Contains(t, lines[1], "[ERROR] Second message")
	assert.Contains(t, lines[2], "[WARN] Third message")
}

// TestLogger_Close_ShouldNotError tests Close method
func TestLogger_Close_ShouldNotError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "close.log")

	logger, err := NewLogger(logFile)
	require.NoError(t, err)

	// Act
	err = logger.Close()

	// Assert
	assert.NoError(t, err, "Close should not return error")
}

// TestLogger_TimestampFormat_ShouldIncludeTimestamp tests that log messages include timestamp
func TestLogger_TimestampFormat_ShouldIncludeTimestamp(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "timestamp.log")

	logger, err := NewLogger(logFile)
	require.NoError(t, err)

	// Act
	logger.Info("Timestamp test")

	// Assert
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)

	// Should contain timestamp pattern (YYYY/MM/DD HH:MM:SS)
	assert.Regexp(t, `\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`, logContent, "Log should contain timestamp")
	assert.Contains(t, logContent, "[INFO] Timestamp test")
}

// TestNewLogger_InvalidPath_ShouldReturnError tests logger creation with invalid path
func TestNewLogger_InvalidPath_ShouldReturnError(t *testing.T) {
	// Arrange - try to create a log file in a directory that can't be created
	invalidPath := "/invalid_root_path/cannot_create/test.log"

	// Act
	logger, err := NewLogger(invalidPath)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, logger)
	assert.Contains(t, err.Error(), "failed to create log directory")
}
