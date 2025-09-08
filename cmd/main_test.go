package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand_Creation_ShouldHaveCorrectProperties(t *testing.T) {
	// Arrange
	version := "test-version"

	// Act
	cmd := newRootCommand(version, nil)

	// Assert
	assert.Equal(t, "octopus", cmd.Use)
	assert.Contains(t, cmd.Short, "Octopus CLI")
	assert.Contains(t, cmd.Long, "Claude Code API")
	assert.Equal(t, version, cmd.Version)
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)
}

func TestRootCommand_Subcommands_ShouldHaveAllExpectedCommands(t *testing.T) {
	// Arrange
	cmd := newRootCommand("test", nil)
	expectedCommands := []string{
		"version", "start", "stop", "status", 
		"config", "health", "logs",
	}

	// Act
	actualCommands := make([]string, 0, len(cmd.Commands()))
	for _, subCmd := range cmd.Commands() {
		actualCommands = append(actualCommands, subCmd.Name())
	}

	// Assert - verify all our commands are present
	for _, expected := range expectedCommands {
		assert.Contains(t, actualCommands, expected, 
			"Missing expected command: %s", expected)
	}
	
	// Verify minimum command count (our commands + any auto-generated)
	assert.GreaterOrEqual(t, len(actualCommands), len(expectedCommands),
		"Should have at least the expected number of commands")
}

func TestRootCommand_GlobalFlags_ShouldHaveConfigAndVerbose(t *testing.T) {
	// Arrange
	cmd := newRootCommand("test", nil)

	// Act & Assert
	configFlag := cmd.PersistentFlags().Lookup("config")
	require.NotNil(t, configFlag, "config flag should exist")
	assert.Equal(t, "", configFlag.DefValue)

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	require.NotNil(t, verboseFlag, "verbose flag should exist")
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestVersionCommand_Execute_ShouldOutputCorrectVersion(t *testing.T) {
	// Arrange
	version := "1.2.3"
	cmd := newVersionCommand(version)
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{}) // No arguments needed for version command

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, version)
	assert.Contains(t, outputStr, "Octopus CLI version")
}

func TestStartCommand_Properties_ShouldHaveCorrectConfiguration(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newStartCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "start", cmd.Use)
	assert.Contains(t, cmd.Short, "Start")
	assert.Contains(t, cmd.Long, "proxy service")
	assert.NotNil(t, cmd.RunE)
}

func TestStopCommand_Properties_ShouldHaveCorrectConfiguration(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newStopCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "stop", cmd.Use)
	assert.Contains(t, cmd.Short, "Stop")
	assert.Contains(t, cmd.Long, "proxy service")
	assert.NotNil(t, cmd.RunE)
}

func TestStatusCommand_Properties_ShouldHaveCorrectConfiguration(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newStatusCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "status", cmd.Use)
	assert.Contains(t, cmd.Short, "status")
	assert.Contains(t, cmd.Long, "status")
	assert.NotNil(t, cmd.RunE)
}

func TestHealthCommand_Properties_ShouldHaveCorrectConfiguration(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newHealthCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "health", cmd.Use)
	assert.Contains(t, cmd.Short, "health")
	assert.Contains(t, cmd.Long, "health")
	assert.NotNil(t, cmd.RunE)
}

func TestLogsCommand_Properties_ShouldHaveCorrectConfiguration(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newLogsCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "logs", cmd.Use)
	assert.Contains(t, cmd.Short, "logs")
	assert.Contains(t, cmd.Long, "logs")
	assert.NotNil(t, cmd.RunE)
}

func TestConfigCommand_Subcommands_ShouldHaveAllExpectedSubcommands(t *testing.T) {
	// Arrange
	var configFile string
	cmd := newConfigCommand(&configFile, nil)
	expectedSubcommands := []string{"list", "add", "remove", "switch", "show"}

	// Act
	actualSubcommands := make([]string, 0, len(cmd.Commands()))
	for _, subCmd := range cmd.Commands() {
		actualSubcommands = append(actualSubcommands, subCmd.Name())
	}

	// Assert
	for _, expected := range expectedSubcommands {
		assert.Contains(t, actualSubcommands, expected, 
			"Missing expected subcommand: %s", expected)
	}
}

func TestConfigListCommand_Properties_ShouldHaveCorrectConfiguration(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newConfigListCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "list", cmd.Use)
	assert.Contains(t, cmd.Short, "List")
	assert.Contains(t, cmd.Aliases, "ls")
	assert.NotNil(t, cmd.RunE)
}

func TestConfigAddCommand_Properties_ShouldRequireThreeArguments(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newConfigAddCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "add <name> <url> <api-key>", cmd.Use)
	assert.Contains(t, cmd.Short, "Add")
	assert.NotNil(t, cmd.Args)
	assert.NotNil(t, cmd.RunE)
	assert.NotEmpty(t, cmd.Example)
}

func TestConfigRemoveCommand_Properties_ShouldRequireOneArgument(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newConfigRemoveCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "remove <name>", cmd.Use)
	assert.Contains(t, cmd.Short, "Remove")
	assert.Contains(t, cmd.Aliases, "rm")
	assert.Contains(t, cmd.Aliases, "delete")
	assert.NotNil(t, cmd.Args)
	assert.NotNil(t, cmd.RunE)
}

func TestConfigSwitchCommand_Properties_ShouldRequireOneArgument(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newConfigSwitchCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "switch <name>", cmd.Use)
	assert.Contains(t, cmd.Short, "Switch")
	assert.NotNil(t, cmd.Args)
	assert.NotNil(t, cmd.RunE)
	assert.NotEmpty(t, cmd.Example)
}

func TestConfigShowCommand_Properties_ShouldRequireOneArgument(t *testing.T) {
	// Arrange & Act
	var configFile string
	cmd := newConfigShowCommand(&configFile, nil)

	// Assert
	assert.Equal(t, "show <name>", cmd.Use)
	assert.Contains(t, cmd.Short, "Show")
	assert.NotNil(t, cmd.Args)
	assert.NotNil(t, cmd.RunE)
	assert.NotEmpty(t, cmd.Example)
}

func TestRootCommand_Help_ShouldContainUsageInformation(t *testing.T) {
	// Arrange
	cmd := newRootCommand("test", nil)
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{"--help"})

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	helpOutput := output.String()
	
	assert.Contains(t, helpOutput, "octopus [command]")
	assert.Contains(t, helpOutput, "Available Commands:")
	assert.Contains(t, helpOutput, "start")
	assert.Contains(t, helpOutput, "config")
	assert.Contains(t, helpOutput, "--verbose")
}