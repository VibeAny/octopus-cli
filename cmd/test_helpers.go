package main

import (
	"path/filepath"
	"testing"

	"octopus-cli/internal/state"
)

// createTestStateManager creates a state manager for testing
func createTestStateManager(t *testing.T) *state.Manager {
	// Create a temporary settings file for testing
	tempDir := t.TempDir()
	settingsFile := filepath.Join(tempDir, "test_settings.toml")

	// Use the new constructor that accepts a settings file path
	return state.NewManagerWithSettingsFile(settingsFile)
}
