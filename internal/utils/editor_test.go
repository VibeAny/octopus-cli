package utils

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectSystemEditor(t *testing.T) {
	editor, err := DetectSystemEditor()

	// Should not error on supported systems
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.NoError(t, err)
		assert.NotNil(t, editor)
		assert.NotEmpty(t, editor.Name)
		assert.NotEmpty(t, editor.Path)
	}
}

func TestDetectUnixEditor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix editor test on Windows")
	}

	editor, err := detectUnixEditor()
	assert.NoError(t, err)
	assert.NotNil(t, editor)
	assert.NotEmpty(t, editor.Name)
	assert.NotEmpty(t, editor.Path)
}

func TestDetectWindowsEditor(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows editor test on non-Windows")
	}

	editor, err := detectWindowsEditor()
	assert.NoError(t, err)
	assert.NotNil(t, editor)
	assert.NotEmpty(t, editor.Name)
	assert.NotEmpty(t, editor.Path)
}

func TestDetectSystemEditorWithCustomEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping environment variable test on Windows")
	}

	// Backup original EDITOR env
	originalEditor := os.Getenv("EDITOR")
	defer func() {
		if originalEditor != "" {
			os.Setenv("EDITOR", originalEditor)
		} else {
			os.Unsetenv("EDITOR")
		}
	}()

	// Set a custom editor (vi should be available on most Unix systems)
	os.Setenv("EDITOR", "vi")

	editor, err := detectUnixEditor()
	assert.NoError(t, err)
	assert.NotNil(t, editor)
	assert.Equal(t, "vi", editor.Name)
}

func TestOpenFileInEditor_InvalidFile(t *testing.T) {
	err := OpenFileInEditor("", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file path cannot be empty")

	err = OpenFileInEditor("/nonexistent/file.txt", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file does not exist")
}

func TestOpenFileInEditor_CustomEditorNotFound(t *testing.T) {
	// Create a temp file for testing
	tempFile := t.TempDir() + "/test.txt"
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	err = OpenFileInEditor(tempFile, "nonexistent-editor")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "custom editor 'nonexistent-editor' not found")
}

func TestIsTerminalEditor(t *testing.T) {
	testCases := []struct {
		editor   string
		expected bool
	}{
		{"vim", true},
		{"nvim", true},
		{"vi", true},
		{"nano", true},
		{"emacs", true},
		{"code", false},
		{"notepad", false},
		{"Visual Studio Code", false},
		{"Notepad++", false},
	}

	for _, tc := range testCases {
		result := isTerminalEditor(tc.editor)
		assert.Equal(t, tc.expected, result, "Editor: %s", tc.editor)
	}
}

func TestGetEditorInfo(t *testing.T) {
	info, err := GetEditorInfo()

	// Should work on supported systems
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.NoError(t, err)
		assert.NotEmpty(t, info)
		assert.Contains(t, info, "Editor:")
		assert.Contains(t, info, "Path:")
	}
}
