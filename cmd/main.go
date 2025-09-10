package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"octopus-cli/internal/config"
	"octopus-cli/internal/state"
	"octopus-cli/internal/utils"
)

var version = "dev"

// logToServiceFile writes a log entry to the service log file
func logToServiceFile(configPath, message string) error {
	// Load configuration to get log file path
	configManager := config.NewManager(configPath)
	cfg, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get log file path from config
	logFile := cfg.Settings.LogFile
	if logFile == "" {
		logFile = "logs/octopus.log"
	}

	// Convert relative paths to absolute paths based on executable directory
	if !filepath.IsAbs(logFile) {
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			logFile = filepath.Join(execDir, logFile)
		}
	}

	// Open log file for appending
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Create logger and write message
	logger := log.New(file, "", log.LstdFlags)
	logger.Printf("[INFO] %s", message)

	return nil
}

// getConfigPath resolves the configuration file path with state management
func getConfigPath(providedConfigFile string, stateManager *state.Manager) (string, bool, error) {
	return state.ResolveConfigFile(providedConfigFile, stateManager)
}

// handleConfigChange manages daemon restart when config changes
func handleConfigChange(configFile string, configChanged bool) error {
	if !configChanged {
		return nil
	}

	// Check if daemon is running
	serviceManager, err := NewServiceManager(configFile)
	if err != nil {
		return fmt.Errorf("failed to create service manager: %w", err)
	}

	status, err := serviceManager.Status()
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if status.IsRunning {
		fmt.Printf("📝 Configuration changed, restarting daemon...\n")

		// Stop the current daemon
		if err := serviceManager.Stop(); err != nil {
			return fmt.Errorf("failed to stop daemon: %w", err)
		}

		// Start with new configuration
		if err := serviceManager.Start(); err != nil {
			return fmt.Errorf("failed to start daemon with new config: %w", err)
		}

		fmt.Printf("✅ Daemon restarted with new configuration\n")
	}

	return nil
}

// runDaemon runs the service in daemon mode
func runDaemon() {
	// Parse config file from command line args
	configFile := ""
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			configFile = os.Args[i+1]
			break
		}
	}

	// Create service manager
	serviceManager, err := NewServiceManager(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create service manager: %v\n", err)
		os.Exit(1)
	}

	// Write PID file for daemon tracking (use current process PID)
	if err := serviceManager.processManager.WritePIDFile(os.Getpid()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write PID file: %v\n", err)
		os.Exit(1)
	}

	// Cleanup PID file on exit
	defer func() {
		if err := serviceManager.processManager.CleanupPIDFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup PID file: %v\n", err)
		}
	}()

	// Start proxy server
	if err := serviceManager.proxyServer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start proxy server: %v\n", err)
		os.Exit(1)
	}

	// Keep daemon running
	select {}
}

// autoStartService automatically starts the service with the specified config
func autoStartService(configFile string) error {
	// Create service manager
	serviceManager, err := NewServiceManager(configFile)
	if err != nil {
		return fmt.Errorf("failed to create service manager: %w", err)
	}

	// Load configuration to check for active API
	cfg, err := serviceManager.configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find and set active API based on is_active flag
	var activeAPI *config.APIConfig
	for _, api := range cfg.APIs {
		if api.IsActive {
			activeAPI = &api
			break
		}
	}

	if activeAPI == nil {
		return fmt.Errorf("no active API found (is_active = true)")
	}

	// Set the active API if different from config
	if cfg.Settings.ActiveAPI != activeAPI.ID {
		fmt.Printf("Setting active API to: %s (%s)\n", activeAPI.ID, activeAPI.Name)
		if err := serviceManager.configManager.SetActiveAPI(activeAPI.ID); err != nil {
			return fmt.Errorf("failed to set active API: %w", err)
		}
	} else {
		fmt.Printf("Active API: %s (%s)\n", activeAPI.ID, activeAPI.Name)
	}

	// Check if service is already running
	status, err := serviceManager.Status()
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if status.IsRunning {
		fmt.Println(utils.FormatSuccess("✅ Service already running") + utils.FormatDim(fmt.Sprintf(" (PID: %d)", status.PID)))
		fmt.Println(utils.FormatInfo("🌐 Proxy available at:") + " " + utils.FormatHighlight(fmt.Sprintf("http://localhost:%d", cfg.Server.Port)))
		return nil
	}

	// Start the service using ServiceManager (daemon mode)
	fmt.Println(utils.FormatInfo("🚀 Starting proxy service..."))

	if err := serviceManager.Start(); err != nil {
		return fmt.Errorf("failed to start proxy server: %w", err)
	}

	fmt.Println(utils.FormatSuccess("✅ Service started successfully!"))
	fmt.Println(utils.FormatInfo("🌐 Proxy available at:") + " " + utils.FormatHighlight(fmt.Sprintf("http://localhost:%d", cfg.Server.Port)))
	fmt.Println(utils.FormatInfo("📊 Active API:") + " " + utils.FormatBold(activeAPI.Name) + utils.FormatDim(" -> ") + utils.FormatDim(activeAPI.URL))

	return nil
}

func main() {
	// Check if running in daemon mode
	if len(os.Args) > 1 && os.Args[1] == "--daemon-mode" {
		runDaemon()
		return
	}

	// Create state manager for config management
	stateManager, err := state.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create state manager: %v\n", err)
		os.Exit(1)
	}

	// Check for auto-start mode (when config file is specified but no command)
	if len(os.Args) >= 3 && (os.Args[1] == "-c" || os.Args[1] == "--config") && len(os.Args) == 3 {
		// Auto-start mode: octopus -c config.toml
		providedConfigFile := os.Args[2]

		configFile, configChanged, err := getConfigPath(providedConfigFile, stateManager)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🚀 Octopus CLI Auto-Start Mode\n")
		fmt.Printf("Config: %s\n", configFile)

		// Handle config change (restart daemon if needed)
		if err := handleConfigChange(configFile, configChanged); err != nil {
			fmt.Fprintf(os.Stderr, "Config change error: %v\n", err)
			os.Exit(1)
		}

		if err := autoStartService(configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Auto-start failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for auto-start mode without config file (use current/default config)
	if len(os.Args) == 1 {
		// Auto-start mode: octopus (no arguments)
		configFile, _, err := getConfigPath("", stateManager)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🚀 Octopus CLI Auto-Start Mode\n")
		fmt.Printf("Config: %s\n", configFile)

		if err := autoStartService(configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Auto-start failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	rootCmd := newRootCommand(version, stateManager)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// newRootCommand creates the root command for octopus CLI
func newRootCommand(version string, stateManager *state.Manager) *cobra.Command {
	var configFile string
	var verbose bool
	var noColor bool

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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Handle color output setting
			if noColor {
				utils.DisableColor()
			} else {
				utils.EnableColor()
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path (default: configs/default.toml or last used)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// Add subcommands - pass state manager reference
	rootCmd.AddCommand(newVersionCommand(version))
	rootCmd.AddCommand(newStartCommand(&configFile, stateManager))
	rootCmd.AddCommand(newStopCommand(&configFile, stateManager))
	rootCmd.AddCommand(newStatusCommand(&configFile, stateManager))
	rootCmd.AddCommand(newConfigCommand(&configFile, stateManager))
	rootCmd.AddCommand(newHealthCommand(&configFile, stateManager))
	rootCmd.AddCommand(newLogsCommand(&configFile, stateManager))
	rootCmd.AddCommand(newUpgradeCommand(&configFile, version))

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

func newStartCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the proxy service",
		Long:  "Start the Octopus proxy service in the background",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve config file path with state management
			cfgPath, configChanged, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}

			if *configFile != "" {
				cmd.Printf("Using config file: %s\n", cfgPath)
			}
			cmd.Println("Starting Octopus proxy service...")

			// Handle config change (restart daemon if needed)
			if err := handleConfigChange(cfgPath, configChanged); err != nil {
				cmd.Printf("Config change error: %v\n", err)
				return err
			}

			// Create service manager
			serviceManager, err := NewServiceManager(cfgPath)
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			// Start the service
			if err := serviceManager.Start(); err != nil {
				cmd.Printf("Failed to start service: %v\n", err)
				return err
			}

			cmd.Println("Service started successfully")
			return nil
		},
	}
}

func newStopCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy service",
		Long:  "Stop the running Octopus proxy service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}

			if *configFile != "" {
				cmd.Printf("Using config file: %s\n", *configFile)
			}
			cmd.Println("Stopping Octopus proxy service...")

			serviceManager, err := NewServiceManager(cfgPath)
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			if err := serviceManager.Stop(); err != nil {
				cmd.Printf("Failed to stop service: %v\n", err)
				return err
			}

			cmd.Println("Service stopped successfully")
			return nil
		},
	}
}

func newStatusCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		Long:  "Display the current status of the Octopus proxy service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}

			if *configFile != "" {
				cmd.Printf("Using config file: %s\n", *configFile)
			}

			serviceManager, err := NewServiceManager(cfgPath)
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			status, err := serviceManager.Status()
			if err != nil {
				cmd.Printf("Failed to get service status: %v\n", err)
				return err
			}

			// Display PID file path for debugging
			pidFilePath := serviceManager.processManager.GetPIDFilePath()
			cmd.Printf("PID file path: %s\n", pidFilePath)

			// Display status information
			if status.IsRunning {
				cmd.Printf("Status: Running\n")
				cmd.Printf("PID: %d\n", status.PID)
			} else {
				cmd.Printf("Status: Stopped\n")
			}

			cmd.Printf("Port: %d\n", status.Port)

			if status.ActiveAPI != "" {
				cmd.Printf("Active API: %s\n", status.ActiveAPI)
			} else {
				cmd.Printf("Active API: (none configured)\n")
			}

			return nil
		},
	}
}

func newHealthCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check API endpoints health",
		Long:  "Check the health status of all configured API endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}

			if *configFile != "" {
				cmd.Printf("Using config file: %s\n", *configFile)
			}
			cmd.Printf("Checking API endpoints health...\n")

			// Load configuration
			configManager := config.NewManager(cfgPath)
			cfg, err := configManager.LoadConfig()
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			// Check if there are any APIs to check
			if len(cfg.APIs) == 0 {
				cmd.Println(utils.FormatWarning("No APIs configured to check"))
				return nil
			}

			cmd.Println(utils.FormatBold("Checking API endpoints health..."))
			cmd.Println()

			// Check health of each API endpoint
			for _, api := range cfg.APIs {
				// Perform actual connectivity check
				status, latency := checkAPIHealth(api.URL, api.APIKey)

				// Determine if healthy based on status
				isHealthy := status == "✅ Healthy"
				responseTime := latency.String()
				if !isHealthy {
					responseTime = "timeout"
				}

				// Format and display API health
				healthDisplay := utils.FormatAPIHealth(api.Name, isHealthy, responseTime)
				cmd.Println(healthDisplay)
				cmd.Println(utils.FormatDim("  URL: " + api.URL))

				// Show if this is the active API
				if api.ID == cfg.Settings.ActiveAPI {
					cmd.Println(utils.FormatHighlight("  Role: [ACTIVE]"))
				}
				cmd.Println()
			}

			return nil
		},
	}
}

func newLogsCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	var follow bool

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View service logs",
		Long:  "Display the Octopus service logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}

			if *configFile != "" {
				cmd.Printf("Using config file: %s\n", *configFile)
			}
			cmd.Printf("Showing service logs...\n")

			// Load configuration
			configManager := config.NewManager(cfgPath)
			cfg, err := configManager.LoadConfig()
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			// Get log file path from config
			logFile := cfg.Settings.LogFile
			if logFile == "" {
				// Default log file location (relative to binary)
				logFile = "logs/octopus.log"
			}

			// Convert relative paths to absolute paths based on executable directory
			if !filepath.IsAbs(logFile) {
				if execPath, err := os.Executable(); err == nil {
					execDir := filepath.Dir(execPath)
					logFile = filepath.Join(execDir, logFile)
				}
			}

			// Check if log file exists
			if _, err := os.Stat(logFile); os.IsNotExist(err) {
				cmd.Printf("Log file not found: %s\n", logFile)
				return fmt.Errorf("log file not found: %s", logFile)
			}

			// Read and display log file
			if follow {
				// Follow mode: tail the file continuously
				if err := followLogFile(cmd, logFile); err != nil {
					cmd.Printf("Failed to follow log file: %v\n", err)
					return err
				}
			} else {
				// Static mode: read and display once
				content, err := os.ReadFile(logFile)
				if err != nil {
					cmd.Printf("Failed to read log file: %v\n", err)
					return err
				}
				cmd.Printf("\n%s", string(content))
			}

			return nil
		},
	}

	// Add follow flag with -f short flag
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")

	return cmd
}

func newConfigCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage API configurations",
		Long:  "Add, remove, list, and switch between API configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show current config info
			if len(args) == 0 {
				// Show current configuration information
				cfgPath, _, err := getConfigPath(*configFile, stateManager)
				if err != nil {
					cmd.Printf("Config error: %v\n", err)
					return err
				}

				cmd.Printf("Current Configuration:\n")
				cmd.Printf("  Config File: %s\n", cfgPath)

				// Load configuration to show active API
				configManager := config.NewManager(cfgPath)
				cfg, err := configManager.LoadConfig()
				if err != nil {
					cmd.Printf("  Status: Failed to load (%v)\n", err)
				} else {
					if cfg.Settings.ActiveAPI != "" {
						cmd.Printf("  Active API: %s\n", cfg.Settings.ActiveAPI)
					} else {
						cmd.Printf("  Active API: (none configured)\n")
					}
					cmd.Printf("  Total APIs: %d\n", len(cfg.APIs))
				}

				cmd.Printf("\nUse 'octopus config --help' to see available subcommands.\n")
				return nil
			}
			// If invalid subcommand is provided, return error
			return fmt.Errorf("unknown subcommand %q for %q", args[0], cmd.CommandPath())
		},
	}

	configCmd.AddCommand(newConfigListCommand(configFile, stateManager))
	configCmd.AddCommand(newConfigAddCommand(configFile, stateManager))
	configCmd.AddCommand(newConfigRemoveCommand(configFile, stateManager))
	configCmd.AddCommand(newConfigSwitchCommand(configFile, stateManager))
	configCmd.AddCommand(newConfigShowCommand(configFile, stateManager))
	configCmd.AddCommand(newConfigEditCommand(configFile, stateManager))

	return configCmd
}

func newConfigListCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all API configurations",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}

			// Load configuration
			configManager := config.NewManager(cfgPath)
			cfg, err := configManager.LoadConfig()
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			// Display API configurations
			cmd.Println(utils.FormatBold("API Configurations:"))

			if len(cfg.APIs) == 0 {
				cmd.Println(utils.FormatDim("No APIs configured"))
				return nil
			}

			// Prepare table data
			headers := []string{"ID", "Name", "Status", "URL"}
			rows := make([][]string, 0, len(cfg.APIs))

			for _, api := range cfg.APIs {
				status := "inactive"
				if api.ID == cfg.Settings.ActiveAPI {
					status = "active"
				}

				// Mask the API key for URL display
				displayURL := api.URL
				if len(displayURL) > 50 {
					displayURL = displayURL[:47] + "..."
				}

				rows = append(rows, []string{
					api.ID,
					api.Name,
					utils.FormatStatus(status),
					displayURL,
				})
			}

			// Display formatted table
			table := utils.FormatTable(headers, rows)
			cmd.Println(table)

			return nil
		},
	}
}

func newConfigAddCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <url> <api-key>",
		Short: "Add a new API configuration",
		Args:  cobra.ExactArgs(3),
		Example: `  octopus config add official https://api.anthropic.com sk-ant-xxx
  octopus config add proxy1 https://api.proxy1.com pk-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}
			name := args[0]
			url := args[1]
			apiKey := args[2]

			// Load existing configuration
			configManager := config.NewManager(cfgPath)
			cfg, err := configManager.LoadConfig()
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			// Create new API config
			newAPI := config.APIConfig{
				ID:         name,
				Name:       name,
				URL:        url,
				APIKey:     apiKey,
				Timeout:    30,
				RetryCount: 3,
			}

			// Add the API
			if err := configManager.AddAPIConfig(&newAPI); err != nil {
				cmd.Printf("Failed to add API configuration: %v\n", err)
				return err
			}

			// Save configuration
			if err := configManager.SaveConfig(cfg); err != nil {
				cmd.Printf("Failed to save configuration: %v\n", err)
				return err
			}

			cmd.Printf("Added API configuration: %s\n", name)
			return nil
		},
	}
}

func newConfigRemoveCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Short:   "Remove an API configuration",
		Aliases: []string{"rm", "delete"},
		Args:    cobra.ExactArgs(1),
		Example: "  octopus config remove proxy1",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}
			name := args[0]

			// Load configuration
			configManager := config.NewManager(cfgPath)
			cfg, err := configManager.LoadConfig()
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			// Check if API exists
			found := false
			for _, api := range cfg.APIs {
				if api.ID == name {
					found = true
					break
				}
			}

			if !found {
				err := fmt.Errorf("API configuration with ID '%s' not found", name)
				cmd.Printf("Error: %v\n", err)
				return err
			}

			// Check if this is the active API before removing
			isActive := cfg.Settings.ActiveAPI == name

			// Remove the API
			if err := configManager.RemoveAPIConfig(name); err != nil {
				cmd.Printf("Failed to remove API configuration: %v\n", err)
				return err
			}

			cmd.Printf("Removed API configuration: %s\n", name)

			// If this was the active API, inform user
			if isActive {
				cmd.Printf("Cleared active API\n")
			}

			return nil
		},
	}
}

func newConfigSwitchCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch to a specific API configuration",
		Args:  cobra.ExactArgs(1),
		Example: `  octopus config switch official
  octopus config switch proxy1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}
			name := args[0]

			// Load configuration
			configManager := config.NewManager(cfgPath)
			cfg, err := configManager.LoadConfig()
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			// Check if the API exists
			var targetAPI *config.APIConfig
			for _, api := range cfg.APIs {
				if api.ID == name {
					targetAPI = &api
					break
				}
			}

			if targetAPI == nil {
				err := fmt.Errorf("API configuration with ID '%s' not found", name)
				cmd.Printf("Error: %v\n", err)
				return err
			}

			// Get the previous active API for logging
			previousAPI := cfg.Settings.ActiveAPI

			// Set active API
			if err := configManager.SetActiveAPI(name); err != nil {
				cmd.Printf("Failed to switch API: %v\n", err)
				return err
			}

			// Save configuration
			if err := configManager.SaveConfig(cfg); err != nil {
				cmd.Printf("Failed to save configuration: %v\n", err)
				return err
			}

			// Log the API switch to service log file
			logMessage := fmt.Sprintf("API switched from '%s' to '%s' (%s -> %s)",
				previousAPI, name, previousAPI, targetAPI.URL)
			if err := logToServiceFile(cfgPath, logMessage); err != nil {
				// Don't fail the command if logging fails, just warn
				cmd.Printf("Warning: Failed to log API switch: %v\n", err)
			}

			// Check if daemon is running and restart it to pick up new configuration
			serviceManager, err := NewServiceManager(cfgPath)
			if err != nil {
				cmd.Printf("Warning: Failed to create service manager: %v\n", err)
			} else {
				status, err := serviceManager.Status()
				if err != nil {
					cmd.Printf("Warning: Failed to check service status: %v\n", err)
				} else if status.IsRunning {
					cmd.Printf("📝 Restarting daemon to apply new API configuration...\n")

					// Stop the current daemon
					if err := serviceManager.Stop(); err != nil {
						cmd.Printf("Warning: Failed to stop daemon: %v\n", err)
					} else {
						// Start with new configuration
						if err := serviceManager.Start(); err != nil {
							cmd.Printf("Warning: Failed to start daemon with new config: %v\n", err)
						} else {
							cmd.Printf("✅ Daemon restarted with new API configuration\n")

							// Log the restart to service log file
							restartMessage := fmt.Sprintf("Daemon restarted to apply API switch to '%s'", name)
							if err := logToServiceFile(cfgPath, restartMessage); err != nil {
								// Don't fail the command if logging fails
								cmd.Printf("Warning: Failed to log daemon restart: %v\n", err)
							}
						}
					}
				}
			}

			cmd.Printf("Switched to API: %s\n", name)
			return nil
		},
	}
}

func newConfigShowCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	return &cobra.Command{
		Use:     "show <name>",
		Short:   "Show details of an API configuration",
		Args:    cobra.ExactArgs(1),
		Example: "  octopus config show official",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}
			name := args[0]

			// Load configuration
			configManager := config.NewManager(cfgPath)
			cfg, err := configManager.LoadConfig()
			if err != nil {
				cmd.Printf("Failed to load configuration: %v\n", err)
				return err
			}

			// Find the API
			var targetAPI *config.APIConfig
			for _, api := range cfg.APIs {
				if api.ID == name {
					targetAPI = &api
					break
				}
			}

			if targetAPI == nil {
				err := fmt.Errorf("API configuration with ID '%s' not found", name)
				cmd.Printf("Error: %v\n", err)
				return err
			}

			// Display API details
			cmd.Printf("API Configuration: %s\n", targetAPI.ID)
			cmd.Printf("  Name: %s\n", targetAPI.Name)
			cmd.Printf("  URL: %s\n", targetAPI.URL)

			// Mask the API key for security
			if targetAPI.APIKey != "" {
				maskedKey := targetAPI.APIKey
				if len(maskedKey) > 5 {
					maskedKey = maskedKey[:3] + "***"
				}
				cmd.Printf("  API Key: %s\n", maskedKey)
			}

			cmd.Printf("  Timeout: %d seconds\n", targetAPI.Timeout)
			cmd.Printf("  Retry Count: %d\n", targetAPI.RetryCount)

			// Show if this is the active API
			if cfg.Settings.ActiveAPI == targetAPI.ID {
				cmd.Printf("  Status: Active\n")
			} else {
				cmd.Printf("  Status: Inactive\n")
			}

			return nil
		},
	}
}

func newConfigEditCommand(configFile *string, stateManager *state.Manager) *cobra.Command {
	var customEditor string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration file with system editor",
		Long:  "Open the current configuration file with the system's default editor or a specified editor",
		Example: `  octopus config edit
  octopus config edit --editor vim
  octopus config edit --editor code`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current config file path
			cfgPath, _, err := getConfigPath(*configFile, stateManager)
			if err != nil {
				cmd.Printf("Config error: %v\n", err)
				return err
			}

			// Verify config file exists
			if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
				cmd.Printf("Configuration file not found: %s\n", cfgPath)
				return fmt.Errorf("configuration file not found: %s", cfgPath)
			}

			cmd.Printf("Opening configuration file: %s\n", cfgPath)

			// Open file in editor
			if err := utils.OpenFileInEditor(cfgPath, customEditor); err != nil {
				cmd.Printf("Failed to open configuration file in editor: %v\n", err)
				return err
			}

			// After editor closes, validate configuration
			cmd.Printf("Configuration file closed. Validating configuration...\n")

			// Load and validate the modified configuration
			configManager := config.NewManager(cfgPath)
			_, err = configManager.LoadConfig()
			if err != nil {
				cmd.Printf("⚠️  Configuration validation failed: %v\n", err)
				cmd.Printf("Please fix the configuration errors and run 'octopus config edit' again if needed.\n")
				return err
			}

			cmd.Printf("✅ Configuration validated successfully!\n")

			// Check if service is running and suggest restart
			serviceManager, err := NewServiceManager(cfgPath)
			if err != nil {
				cmd.Printf("Warning: Could not check service status: %v\n", err)
				return nil
			}

			status, err := serviceManager.Status()
			if err != nil {
				cmd.Printf("Warning: Could not check service status: %v\n", err)
				return nil
			}

			if status.IsRunning {
				cmd.Printf("💡 Service is currently running. To apply configuration changes, run:\n")
				cmd.Printf("   octopus stop && octopus start\n")
			}

			return nil
		},
	}

	// Add editor flag
	cmd.Flags().StringVar(&customEditor, "editor", "", "Specify editor to use (e.g., vim, code, nano)")

	return cmd
}

// checkAPIHealth performs a health check on an API endpoint
func checkAPIHealth(apiURL, apiKey string) (status string, latency time.Duration) {
	startTime := time.Now()

	// Create a simple health check request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "❌ Invalid URL", 0
	}

	// Add API key if provided
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Set proper headers for Anthropic API
	req.Header.Set("User-Agent", "Octopus-CLI/1.0")
	req.Header.Set("Accept", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: nil, // Disable proxy to avoid system proxy interference
		},
	}

	resp, err := client.Do(req)
	latency = time.Since(startTime)

	if err != nil {
		return "❌ Connection failed", latency
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "✅ Healthy", latency
	} else if resp.StatusCode == 401 {
		return "⚠️ Unauthorized (API key issue)", latency
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return "⚠️ Client error", latency
	} else if resp.StatusCode >= 500 {
		return "❌ Server error", latency
	}

	return "⚠️ Unknown status", latency
}

// followLogFile implements tail-like functionality for log files
func followLogFile(cmd *cobra.Command, logFile string) error {
	// Set up signal handling for graceful exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cmd.Printf("\n\nStopping log following...\n")
		cancel()
	}()

	// First, display the last 20 lines of existing content
	if err := displayRecentLogLines(cmd, logFile, 20); err != nil {
		cmd.Printf("Warning: Could not display recent log lines: %v\n", err)
	}

	cmd.Printf("\n--- Following logs (Press Ctrl+C to exit) ---\n")

	// Use a file watcher approach with stat checking
	var lastSize int64 = -1
	var lastModTime time.Time

	// Get initial file info
	if info, err := os.Stat(logFile); err == nil {
		lastSize = info.Size()
		lastModTime = info.ModTime()
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Check if file has been modified
			info, err := os.Stat(logFile)
			if err != nil {
				// File might have been removed or rotated, wait for it to reappear
				lastSize = -1
				continue
			}

			currentSize := info.Size()
			currentModTime := info.ModTime()

			// Check if file has new content
			if currentSize > lastSize || currentModTime.After(lastModTime) {
				if err := readNewContent(cmd, logFile, lastSize, currentSize); err != nil {
					cmd.Printf("Error reading new content: %v\n", err)
				}
				lastSize = currentSize
				lastModTime = currentModTime
			} else if currentSize < lastSize {
				// File was truncated or rotated
				cmd.Printf("\n--- Log file was rotated ---\n")
				lastSize = 0
				// Read from beginning
				if err := readNewContent(cmd, logFile, 0, currentSize); err != nil {
					cmd.Printf("Error reading rotated content: %v\n", err)
				}
				lastSize = currentSize
				lastModTime = currentModTime
			}
		}
	}
}

// displayRecentLogLines shows the last N lines from the log file
func displayRecentLogLines(cmd *cobra.Command, logFile string, maxLines int) error {
	file, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read file content and get last N lines
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		cmd.Printf("Log file is empty\n")
		return nil
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	startIdx := len(lines) - maxLines
	if startIdx < 0 {
		startIdx = 0
	}

	cmd.Printf("--- Last %d lines of log ---\n", len(lines)-startIdx)
	for i := startIdx; i < len(lines); i++ {
		cmd.Printf("%s\n", lines[i])
	}

	return nil
}

// readNewContent reads new content from the file starting at offset
func readNewContent(cmd *cobra.Command, logFile string, startOffset, endOffset int64) error {
	if startOffset >= endOffset {
		return nil
	}

	file, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Seek to the start offset
	if startOffset > 0 {
		_, err = file.Seek(startOffset, io.SeekStart)
		if err != nil {
			return err
		}
	}

	// Read only the new content
	buffer := make([]byte, endOffset-startOffset)
	n, err := io.ReadFull(file, buffer)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return err
	}

	// Output new content, ensuring proper line handling
	content := string(buffer[:n])
	if content != "" {
		// Remove trailing newline to avoid double newlines
		content = strings.TrimRight(content, "\n")
		if content != "" {
			cmd.Printf("%s\n", content)
		}
	}

	return nil
}

func newUpgradeCommand(configFile *string, version string) *cobra.Command {
	var checkOnly bool
	var force bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade to the latest version",
		Long:  "Check for the latest version of Octopus CLI and upgrade to it",
		Example: `  octopus upgrade
  octopus upgrade --check
  octopus upgrade --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create version checker
			versionChecker := utils.NewVersionChecker("VibeAny/octopus-cli", version)

			cmd.Printf("🔍 Checking for upgrades...\n")

			// Check if upgrade is available
			isAvailable, latestRelease, err := versionChecker.IsUpdateAvailable()
			if err != nil {
				cmd.Printf("❌ Failed to check for upgrades: %v\n", err)
				return err
			}

			if !isAvailable {
				cmd.Printf("✅ You are using the latest version (%s)\n", utils.FormatHighlight(version))
				return nil
			}

			// Display upgrade information
			upgradeInfo := utils.FormatUpdateInfo(version, latestRelease.TagName, latestRelease)
			// Update the text to use "upgrade"
			upgradeInfo = strings.Replace(upgradeInfo, "octopus update", "octopus upgrade", -1)
			cmd.Printf("%s\n", upgradeInfo)

			// If check-only mode, just return
			if checkOnly {
				return nil
			}

			// Ask for confirmation unless forced
			if !force {
				cmd.Printf("\nDo you want to upgrade now? [y/N]: ")

				var response string
				_, _ = fmt.Scanln(&response)

				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					cmd.Printf("Upgrade cancelled.\n")
					return nil
				}
			}

			// Proceed with upgrade
			cmd.Printf("\n🚀 Starting upgrade process...\n")

			// Create update manager
			updateManager := utils.NewUpdateManager("VibeAny/octopus-cli", version)
			defer updateManager.Cleanup()

			// Get current platform
			platform := utils.GetCurrentPlatform()
			cmd.Printf("📋 Platform: %s-%s\n", platform.OS, platform.Arch)

			// Find asset for current platform
			asset, err := updateManager.FindAssetForPlatform(latestRelease, platform)
			if err != nil {
				cmd.Printf("❌ Failed to find upgrade for your platform: %v\n", err)
				return err
			}

			cmd.Printf("📦 Found upgrade: %s (%.1f MB)\n", asset.Name, float64(asset.Size)/1024/1024)

			// Create backup of current binary
			cmd.Printf("💾 Creating backup of current version...\n")
			backupPath, err := updateManager.BackupCurrentBinary()
			if err != nil {
				cmd.Printf("❌ Failed to create backup: %v\n", err)
				return err
			}

			// Progress callback
			var lastPercent int
			progressCallback := func(progress utils.DownloadProgress) {
				percent := int(progress.Percentage)
				if percent > lastPercent && percent%10 == 0 {
					cmd.Printf("⬇️  Downloaded: %.1f%% (%s) - %s\n",
						progress.Percentage,
						formatBytes(progress.Downloaded),
						progress.Speed)
					lastPercent = percent
				}
			}

			// Download upgrade
			cmd.Printf("⬇️  Downloading upgrade...\n")
			downloadPath, err := updateManager.DownloadUpdate(asset, progressCallback)
			if err != nil {
				cmd.Printf("❌ Failed to download upgrade: %v\n", err)
				return err
			}

			// Verify download
			cmd.Printf("🔍 Verifying download...\n")
			if err := updateManager.VerifyDownload(downloadPath, asset.Size); err != nil {
				cmd.Printf("❌ Download verification failed: %v\n", err)
				return err
			}

			// Install upgrade
			cmd.Printf("🔄 Installing upgrade...\n")
			if err := updateManager.InstallUpdate(downloadPath); err != nil {
				cmd.Printf("❌ Failed to install upgrade: %v\n", err)

				// Try to restore from backup
				cmd.Printf("🔄 Attempting to restore from backup...\n")
				if restoreErr := updateManager.RestoreFromBackup(backupPath); restoreErr != nil {
					cmd.Printf("❌ Failed to restore from backup: %v\n", restoreErr)
					cmd.Printf("⚠️  Please restore manually from: %s\n", backupPath)
				} else {
					cmd.Printf("✅ Restored from backup successfully\n")
				}

				return err
			}

			// Clean up backup (keep it commented for safety)
			// os.Remove(backupPath)

			cmd.Printf("✅ Upgrade completed successfully!\n")
			cmd.Printf("🎉 Octopus CLI has been upgraded to %s\n", utils.FormatHighlight(latestRelease.TagName))

			// Check if service was running and restart if needed
			cmd.Printf("🔄 Checking for running service...\n")
			if err := handleServiceRestart(cmd, *configFile); err != nil {
				cmd.Printf("⚠️  Warning: Failed to restart service: %v\n", err)
				cmd.Printf("💡 Please manually restart the service with 'octopus start'\n")
			}

			cmd.Printf("💡 Restart your terminal or run 'octopus version' to verify the upgrade.\n")

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for upgrades, don't install")
	cmd.Flags().BoolVar(&force, "force", false, "Install upgrade without confirmation")

	return cmd
}

// handleServiceRestart checks if service is running and restarts it after upgrade
func handleServiceRestart(cmd *cobra.Command, configFile string) error {
	// Create state manager for config management
	stateManager, err := state.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create state manager: %w", err)
	}

	// Resolve config file path
	cfgPath, _, err := getConfigPath(configFile, stateManager)
	if err != nil {
		return fmt.Errorf("failed to resolve config file: %w", err)
	}

	// Create service manager to check current status
	serviceManager, err := NewServiceManager(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to create service manager: %w", err)
	}

	// Check if service is currently running
	status, err := serviceManager.Status()
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if !status.IsRunning {
		cmd.Printf("📋 Service was not running - no restart needed\n")
		return nil
	}

	cmd.Printf("🔄 Service is running - restarting with upgraded binary...\n")

	// Stop the current service
	cmd.Printf("⏹️  Stopping current service...\n")
	if err := serviceManager.Stop(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Brief pause to ensure cleanup
	time.Sleep(1 * time.Second)

	// Start the service with the new binary
	cmd.Printf("▶️  Starting service with upgraded binary...\n")
	if err := serviceManager.Start(); err != nil {
		return fmt.Errorf("failed to start upgraded service: %w", err)
	}

	// Verify the service started successfully
	newStatus, err := serviceManager.Status()
	if err != nil {
		return fmt.Errorf("failed to verify service restart: %w", err)
	}

	if newStatus.IsRunning {
		cmd.Printf("✅ Service restarted successfully with upgraded binary\n")
		cmd.Printf("📋 Service running on port %d with PID %d\n", newStatus.Port, newStatus.PID)
		
		// Log the upgrade to service log file
		logMessage := fmt.Sprintf("Service upgraded and restarted with new binary version")
		if err := logToServiceFile(cfgPath, logMessage); err != nil {
			// Don't fail if logging fails
			cmd.Printf("Warning: Failed to log upgrade: %v\n", err)
		}
	} else {
		return fmt.Errorf("service failed to start after upgrade")
	}

	return nil
}

// formatBytes formats bytes as human readable string (helper for CLI use)
func formatBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}

	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(1024), 0
	for n := bytes / 1024; n >= 1024; n /= 1024 {
		div *= 1024
		exp++
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp+1])
}
