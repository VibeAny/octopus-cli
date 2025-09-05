package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := newRootCommand(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// newRootCommand creates the root command for octopus CLI
func newRootCommand(version string) *cobra.Command {
	var configFile string
	var verbose bool

	rootCmd := &cobra.Command{
		Use:   "octopus",
		Short: "Octopus CLI - Dynamic Claude Code API management",
		Long: `Octopus CLI is a command-line tool that provides local API forwarding 
proxy service to solve Claude Code API switching problems.

It allows you to configure multiple API endpoints and keys via TOML 
configuration files, then dynamically switch between them without 
restarting Claude Code or modifying environment variables.`,
		Version: version,
		Example: `  octopus start
  octopus config add official https://api.anthropic.com sk-ant-xxx
  octopus config switch official
  octopus status`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path (default: ~/.config/octopus/octopus.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(newVersionCommand(version))
	rootCmd.AddCommand(newStartCommand())
	rootCmd.AddCommand(newStopCommand())
	rootCmd.AddCommand(newStatusCommand())
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newHealthCommand())
	rootCmd.AddCommand(newLogsCommand())

	return rootCmd
}

func newVersionCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Octopus CLI version %s\n", version)
		},
	}
}

func newStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the proxy service",
		Long:  "Start the Octopus proxy service in the background",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting Octopus proxy service...")
			// TODO: Implement start functionality
		},
	}
}

func newStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy service",
		Long:  "Stop the running Octopus proxy service",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Stopping Octopus proxy service...")
			// TODO: Implement stop functionality
		},
	}
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		Long:  "Display the current status of the Octopus proxy service",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Checking Octopus service status...")
			// TODO: Implement status functionality
		},
	}
}

func newHealthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check API endpoints health",
		Long:  "Check the health status of all configured API endpoints",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Checking API endpoints health...")
			// TODO: Implement health functionality
		},
	}
}

func newLogsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "View service logs",
		Long:  "Display the Octopus service logs",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Showing service logs...")
			// TODO: Implement logs functionality
		},
	}
}

func newConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage API configurations",
		Long:  "Add, remove, list, and switch between API configurations",
	}

	configCmd.AddCommand(newConfigListCommand())
	configCmd.AddCommand(newConfigAddCommand())
	configCmd.AddCommand(newConfigRemoveCommand())
	configCmd.AddCommand(newConfigSwitchCommand())
	configCmd.AddCommand(newConfigShowCommand())

	return configCmd
}

func newConfigListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all API configurations",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Listing API configurations...")
			// TODO: Implement config list functionality
		},
	}
}

func newConfigAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <url> <api-key>",
		Short: "Add a new API configuration",
		Args:  cobra.ExactArgs(3),
		Example: `  octopus config add official https://api.anthropic.com sk-ant-xxx
  octopus config add proxy1 https://api.proxy1.com pk-xxx`,
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			url := args[1]
			apiKey := args[2]
			fmt.Printf("Adding API configuration: %s -> %s\n", name, url)
			// TODO: Implement config add functionality
			_ = apiKey // Avoid unused variable warning
		},
	}
}

func newConfigRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Short:   "Remove an API configuration",
		Aliases: []string{"rm", "delete"},
		Args:    cobra.ExactArgs(1),
		Example: "  octopus config remove proxy1",
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			fmt.Printf("Removing API configuration: %s\n", name)
			// TODO: Implement config remove functionality
		},
	}
}

func newConfigSwitchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch to a specific API configuration",
		Args:  cobra.ExactArgs(1),
		Example: `  octopus config switch official
  octopus config switch proxy1`,
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			fmt.Printf("Switching to API configuration: %s\n", name)
			// TODO: Implement config switch functionality
		},
	}
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "show <name>",
		Short:   "Show details of an API configuration",
		Args:    cobra.ExactArgs(1),
		Example: "  octopus config show official",
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			fmt.Printf("Showing API configuration: %s\n", name)
			// TODO: Implement config show functionality
		},
	}
}