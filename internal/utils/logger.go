package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Logger represents a simple logger
type Logger struct {
	*log.Logger
	filePath string
}

// NewLogger creates a new logger that writes to the specified file
func NewLogger(filePath string) (*Logger, error) {
	// Convert relative paths to absolute paths based on executable directory
	if !filepath.IsAbs(filePath) {
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			filePath = filepath.Join(execDir, filePath)
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create log file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := log.New(file, "", log.LstdFlags)
	return &Logger{
		Logger:   logger,
		filePath: filePath,
	}, nil
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.Printf("[INFO] "+format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.Printf("[ERROR] "+format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.Printf("[WARN] "+format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.Printf("[DEBUG] "+format, v...)
}

// Close closes the logger (if needed for cleanup)
func (l *Logger) Close() error {
	// Note: log.Logger doesn't expose the underlying writer
	// In a more sophisticated implementation, we'd keep track of the file
	return nil
}
