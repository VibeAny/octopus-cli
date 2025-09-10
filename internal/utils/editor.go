package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// EditorInfo contains information about a detected editor
type EditorInfo struct {
	Name string
	Path string
	Args []string
}

// DetectSystemEditor detects and returns the best available system editor
func DetectSystemEditor() (*EditorInfo, error) {
	switch runtime.GOOS {
	case "linux", "darwin":
		return detectUnixEditor()
	case "windows":
		return detectWindowsEditor()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// detectUnixEditor detects editors on Unix-like systems (Linux, macOS)
func detectUnixEditor() (*EditorInfo, error) {
	// Priority 1: Check EDITOR environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		if path, err := exec.LookPath(editor); err == nil {
			return &EditorInfo{
				Name: editor,
				Path: path,
				Args: []string{},
			}, nil
		}
	}

	// Priority 2: Check common editors in order of preference
	editors := []string{
		"vim", "nvim", "nano", "vi", "emacs",
		"code", "subl", "atom", "gedit", "kate",
	}

	for _, editor := range editors {
		if path, err := exec.LookPath(editor); err == nil {
			// Special handling for VS Code
			if editor == "code" {
				return &EditorInfo{
					Name: "Visual Studio Code",
					Path: path,
					Args: []string{"--wait"},
				}, nil
			}
			// Special handling for Sublime Text
			if editor == "subl" {
				return &EditorInfo{
					Name: "Sublime Text",
					Path: path,
					Args: []string{"--wait"},
				}, nil
			}

			return &EditorInfo{
				Name: editor,
				Path: path,
				Args: []string{},
			}, nil
		}
	}

	return nil, fmt.Errorf("no suitable editor found")
}

// detectWindowsEditor detects editors on Windows
func detectWindowsEditor() (*EditorInfo, error) {
	// Priority 1: Check common GUI editors
	guiEditors := map[string][]string{
		"code.exe":      {"Visual Studio Code", "--wait"},
		"notepad++.exe": {"Notepad++", "-multiInst", "-nosession"},
		"subl.exe":      {"Sublime Text", "--wait"},
		"atom.exe":      {"Atom", "--wait"},
	}

	for exe, info := range guiEditors {
		if path, err := exec.LookPath(exe); err == nil {
			return &EditorInfo{
				Name: info[0],
				Path: path,
				Args: info[1:],
			}, nil
		}
	}

	// Priority 2: Check notepad.exe (always available on Windows)
	if path, err := exec.LookPath("notepad.exe"); err == nil {
		return &EditorInfo{
			Name: "Notepad",
			Path: path,
			Args: []string{},
		}, nil
	}

	return nil, fmt.Errorf("no suitable editor found")
}

// OpenFileInEditor opens a file using the specified or detected system editor
func OpenFileInEditor(filePath string, customEditor string) error {
	var editor *EditorInfo
	var err error

	// Use custom editor if specified
	if customEditor != "" {
		if path, lookErr := exec.LookPath(customEditor); lookErr == nil {
			editor = &EditorInfo{
				Name: customEditor,
				Path: path,
				Args: []string{},
			}
		} else {
			return fmt.Errorf("custom editor '%s' not found: %w", customEditor, lookErr)
		}
	} else {
		// Auto-detect system editor
		editor, err = DetectSystemEditor()
		if err != nil {
			return fmt.Errorf("failed to detect system editor: %w", err)
		}
	}

	// Validate file path
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Prepare command arguments
	args := append(editor.Args, filePath)

	// Create and execute command
	cmd := exec.Command(editor.Path, args...)

	// For GUI editors, we want to wait for them to close
	// For terminal editors, we need to pass through stdin/stdout
	if isTerminalEditor(editor.Name) {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// isTerminalEditor returns true if the editor runs in terminal
func isTerminalEditor(editorName string) bool {
	terminalEditors := []string{
		"vim", "nvim", "vi", "nano", "emacs", "ed", "ex",
	}

	lowerName := strings.ToLower(editorName)

	// Check for exact matches or word boundaries
	for _, te := range terminalEditors {
		if lowerName == te {
			return true
		}
		// Check if it's a word boundary match (not part of a larger word)
		if strings.HasPrefix(lowerName, te+" ") || strings.HasSuffix(lowerName, " "+te) ||
			strings.Contains(lowerName, " "+te+" ") {
			return true
		}
	}

	return false
}

// GetEditorInfo returns information about the current system editor
func GetEditorInfo() (string, error) {
	editor, err := DetectSystemEditor()
	if err != nil {
		return "", err
	}

	info := fmt.Sprintf("Editor: %s\nPath: %s", editor.Name, editor.Path)
	if len(editor.Args) > 0 {
		info += fmt.Sprintf("\nDefault Args: %s", strings.Join(editor.Args, " "))
	}

	return info, nil
}
