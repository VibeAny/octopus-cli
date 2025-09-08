package utils

import (
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

// TestMain sets up test environment
func TestMain(m *testing.M) {
	// Force color output for testing
	color.NoColor = false
	os.Setenv("FORCE_COLOR", "1")
	os.Exit(m.Run())
}

func TestFormatSuccess_ShouldReturnGreenText(t *testing.T) {
	// Arrange
	message := "Operation completed successfully"
	
	// Act
	result := FormatSuccess(message)
	
	// Assert
	assert.Contains(t, result, message)
	// Should contain ANSI color codes for green
	assert.Contains(t, result, "\x1b[32m") // Green color code
}

func TestFormatError_ShouldReturnRedText(t *testing.T) {
	// Arrange
	message := "Error occurred"
	
	// Act
	result := FormatError(message)
	
	// Assert
	assert.Contains(t, result, message)
	// Should contain ANSI color codes for red
	assert.Contains(t, result, "\x1b[31m") // Red color code
}

func TestFormatWarning_ShouldReturnYellowText(t *testing.T) {
	// Arrange
	message := "Warning message"
	
	// Act
	result := FormatWarning(message)
	
	// Assert
	assert.Contains(t, result, message)
	// Should contain ANSI color codes for yellow
	assert.Contains(t, result, "\x1b[33m") // Yellow color code
}

func TestFormatInfo_ShouldReturnBlueText(t *testing.T) {
	// Arrange
	message := "Information message"
	
	// Act
	result := FormatInfo(message)
	
	// Assert
	assert.Contains(t, result, message)
	// Should contain ANSI color codes for blue
	assert.Contains(t, result, "\x1b[34m") // Blue color code
}

func TestFormatHighlight_ShouldReturnCyanText(t *testing.T) {
	// Arrange
	message := "Highlighted text"
	
	// Act
	result := FormatHighlight(message)
	
	// Assert
	assert.Contains(t, result, message)
	// Should contain ANSI color codes for cyan
	assert.Contains(t, result, "\x1b[36m") // Cyan color code
}

func TestFormatBold_ShouldReturnBoldText(t *testing.T) {
	// Arrange
	message := "Bold text"
	
	// Act
	result := FormatBold(message)
	
	// Assert
	assert.Contains(t, result, message)
	// Should contain ANSI code for bold
	assert.Contains(t, result, "\x1b[1m") // Bold code
}

func TestFormatDim_ShouldReturnDimText(t *testing.T) {
	// Arrange
	message := "Dim text"
	
	// Act
	result := FormatDim(message)
	
	// Assert
	assert.Contains(t, result, message)
	// Should contain ANSI code for dim
	assert.Contains(t, result, "\x1b[2m") // Dim code
}

func TestFormatTable_ShouldReturnFormattedTable(t *testing.T) {
	// Arrange
	headers := []string{"Name", "Status", "URL"}
	rows := [][]string{
		{"API 1", "Active", "https://api1.com"},
		{"API 2", "Inactive", "https://api2.com"},
	}
	
	// Act
	result := FormatTable(headers, rows)
	
	// Assert
	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "Status")
	assert.Contains(t, result, "URL")
	assert.Contains(t, result, "API 1")
	assert.Contains(t, result, "API 2")
	// Should have table formatting
	assert.Contains(t, result, "│") // Table borders
}

func TestFormatStatus_WithRunningStatus_ShouldReturnGreenRunning(t *testing.T) {
	// Act
	result := FormatStatus("running")
	
	// Assert
	assert.Contains(t, result, "running")
	assert.Contains(t, result, "\x1b[32m") // Green for running
}

func TestFormatStatus_WithStoppedStatus_ShouldReturnRedStopped(t *testing.T) {
	// Act
	result := FormatStatus("stopped")
	
	// Assert
	assert.Contains(t, result, "stopped")
	assert.Contains(t, result, "\x1b[31m") // Red for stopped
}

func TestFormatStatus_WithUnknownStatus_ShouldReturnYellowUnknown(t *testing.T) {
	// Act
	result := FormatStatus("unknown")
	
	// Assert
	assert.Contains(t, result, "unknown")
	assert.Contains(t, result, "\x1b[33m") // Yellow for unknown
}

func TestFormatAPIHealth_WithHealthyStatus_ShouldReturnGreenCheck(t *testing.T) {
	// Arrange
	apiName := "Test API"
	isHealthy := true
	responseTime := "123ms"
	
	// Act
	result := FormatAPIHealth(apiName, isHealthy, responseTime)
	
	// Assert
	assert.Contains(t, result, apiName)
	assert.Contains(t, result, responseTime)
	assert.Contains(t, result, "✓") // Check mark
	assert.Contains(t, result, "\x1b[32m") // Green color
}

func TestFormatAPIHealth_WithUnhealthyStatus_ShouldReturnRedCross(t *testing.T) {
	// Arrange
	apiName := "Test API"
	isHealthy := false
	responseTime := "timeout"
	
	// Act
	result := FormatAPIHealth(apiName, isHealthy, responseTime)
	
	// Assert
	assert.Contains(t, result, apiName)
	assert.Contains(t, result, responseTime)
	assert.Contains(t, result, "✗") // Cross mark
	assert.Contains(t, result, "\x1b[31m") // Red color
}

func TestDisableColor_ShouldReturnPlainText(t *testing.T) {
	// Arrange
	message := "Test message"
	DisableColor()
	
	// Act
	result := FormatSuccess(message)
	
	// Assert
	assert.Equal(t, message, result)
	assert.NotContains(t, result, "\x1b[") // No ANSI codes
	
	// Cleanup - re-enable color for other tests
	EnableColor()
}

func TestEnableColor_ShouldReturnColoredText(t *testing.T) {
	// Arrange
	message := "Test message"
	DisableColor()
	EnableColor()
	
	// Act
	result := FormatSuccess(message)
	
	// Assert
	assert.Contains(t, result, message)
	assert.Contains(t, result, "\x1b[32m") // Green color code
}

func TestStripANSI_ShouldRemoveColorCodes(t *testing.T) {
	// Arrange
	coloredText := "\x1b[32mGreen text\x1b[0m"
	
	// Act
	result := stripANSI(coloredText)
	
	// Assert
	assert.Equal(t, "Green text", result)
}

func TestVisualLen_ShouldReturnCorrectLength(t *testing.T) {
	// Arrange
	plainText := "Hello World"
	coloredText := "\x1b[32mHello World\x1b[0m"
	
	// Act
	plainLength := visualLen(plainText)
	coloredLength := visualLen(coloredText)
	
	// Assert
	assert.Equal(t, 11, plainLength)
	assert.Equal(t, 11, coloredLength) // Should be same as plain text
}

func TestFormatTable_WithColoredStatus_ShouldAlignCorrectly(t *testing.T) {
	// Arrange
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"API 1", FormatSuccess("active")},
		{"API 2", FormatError("inactive")},
	}
	
	// Act
	result := FormatTable(headers, rows)
	
	// Assert
	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "Status")
	assert.Contains(t, result, "API 1")
	assert.Contains(t, result, "API 2")
	// Should have proper table formatting with colored cells
	assert.Contains(t, result, "│")
	// The table should be properly aligned despite colored text
	lines := strings.Split(result, "\n")
	assert.True(t, len(lines) > 3) // Should have header, separator, and data rows
}