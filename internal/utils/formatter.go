package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// Color formatters for different message types
var (
	successColor   = color.New(color.FgGreen)
	errorColor     = color.New(color.FgRed)
	warningColor   = color.New(color.FgYellow)
	infoColor      = color.New(color.FgBlue)
	highlightColor = color.New(color.FgCyan)
	boldFormat     = color.New(color.Bold)
	dimFormat      = color.New(color.Faint)
)

// ANSI color code regex for stripping color codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// visualLen returns the visual length of a string (excluding ANSI codes)
func visualLen(s string) int {
	return len(stripANSI(s))
}

// FormatSuccess returns text formatted in green color for success messages
func FormatSuccess(message string) string {
	return successColor.Sprint(message)
}

// FormatError returns text formatted in red color for error messages
func FormatError(message string) string {
	return errorColor.Sprint(message)
}

// FormatWarning returns text formatted in yellow color for warning messages
func FormatWarning(message string) string {
	return warningColor.Sprint(message)
}

// FormatInfo returns text formatted in blue color for info messages
func FormatInfo(message string) string {
	return infoColor.Sprint(message)
}

// FormatHighlight returns text formatted in cyan color for highlighted text
func FormatHighlight(message string) string {
	return highlightColor.Sprint(message)
}

// FormatBold returns text formatted in bold
func FormatBold(message string) string {
	return boldFormat.Sprint(message)
}

// FormatDim returns text formatted as dim/faint
func FormatDim(message string) string {
	return dimFormat.Sprint(message)
}

// FormatStatus returns colored status text based on status type
func FormatStatus(status string) string {
	switch strings.ToLower(status) {
	case "running", "active", "healthy", "online":
		return FormatSuccess(status)
	case "stopped", "inactive", "offline", "failed", "error":
		return FormatError(status)
	case "unknown", "pending", "warning":
		return FormatWarning(status)
	default:
		return FormatDim(status)
	}
}

// FormatAPIHealth formats API health status with icons and colors
func FormatAPIHealth(apiName string, isHealthy bool, responseTime string) string {
	var icon, status string
	
	if isHealthy {
		icon = FormatSuccess("✓")
		status = FormatSuccess("healthy")
	} else {
		icon = FormatError("✗")
		status = FormatError("unhealthy")
	}
	
	return fmt.Sprintf("%s %s (%s) - %s", 
		icon, 
		FormatBold(apiName), 
		status, 
		FormatDim(responseTime))
}

// FormatTable creates a formatted table with headers and rows
func FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}
	
	// Calculate column widths using visual length (excluding ANSI codes)
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = visualLen(header)
	}
	
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && visualLen(cell) > colWidths[i] {
				colWidths[i] = visualLen(cell)
			}
		}
	}
	
	var result strings.Builder
	
	// Top border
	result.WriteString("┌")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			result.WriteString("┬")
		}
	}
	result.WriteString("┐\n")
	
	// Headers
	result.WriteString("│")
	for i, header := range headers {
		result.WriteString(" ")
		// Use visual length for padding calculation
		padding := colWidths[i] - visualLen(header)
		result.WriteString(FormatBold(header))
		result.WriteString(strings.Repeat(" ", padding))
		result.WriteString(" │")
	}
	result.WriteString("\n")
	
	// Header separator
	result.WriteString("├")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			result.WriteString("┼")
		}
	}
	result.WriteString("┤\n")
	
	// Rows
	for _, row := range rows {
		result.WriteString("│")
		for i, cell := range row {
			if i < len(colWidths) {
				result.WriteString(" ")
				// Use visual length for padding calculation
				padding := colWidths[i] - visualLen(cell)
				result.WriteString(cell)
				result.WriteString(strings.Repeat(" ", padding))
				result.WriteString(" │")
			}
		}
		result.WriteString("\n")
	}
	
	// Bottom border
	result.WriteString("└")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			result.WriteString("┴")
		}
	}
	result.WriteString("┘")
	
	return result.String()
}

// DisableColor disables all color output
func DisableColor() {
	color.NoColor = true
}

// EnableColor enables color output
func EnableColor() {
	color.NoColor = false
}

// IsColorEnabled returns whether color output is enabled
func IsColorEnabled() bool {
	return !color.NoColor
}