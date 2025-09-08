package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	if len(os.Args) >= 3 && (os.Args[1] == "-f" || os.Args[1] == "--config") && len(os.Args) == 3 {
		// Auto-start mode: octopus -f config.toml
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
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "f", "", "config file path (default: configs/default.toml or last used)")
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
				responseTime := string(latency)
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

	// Add follow flag (no short flag to avoid conflict with -f config flag)
	cmd.Flags().BoolVar(&follow, "follow", false, "Follow log output")
	
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
	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()
	
	// Seek to end of file to start following from new content
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to end of file: %w", err)
	}
	
	cmd.Printf("Following logs (Press Ctrl+C to exit)...\n\n")
	
	reader := bufio.NewReader(file)
	
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				// No more content, wait and retry
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return fmt.Errorf("error reading log file: %w", err)
		}
		
		// Print new line (handle partial lines)
		cmd.Printf("%s", string(line))
		if !isPrefix {
			cmd.Printf("\n")
		}
	}
}