package config

// Config represents the main configuration structure
type Config struct {
	Server   ServerConfig `toml:"server"`
	APIs     []APIConfig  `toml:"apis"`
	Settings Settings     `toml:"settings"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port     int    `toml:"port"`
	LogLevel string `toml:"log_level"`
	Daemon   bool   `toml:"daemon"`
}

// APIConfig represents an API configuration
type APIConfig struct {
	ID         string `toml:"id"`
	Name       string `toml:"name"`
	URL        string `toml:"url"`
	APIKey     string `toml:"api_key"`
	IsActive   bool   `toml:"is_active"`
	Timeout    int    `toml:"timeout"`
	RetryCount int    `toml:"retry_count"`
}

// Settings represents global settings
type Settings struct {
	ActiveAPI    string `toml:"active_api"`
	LogFile      string `toml:"log_file"`
	ConfigBackup bool   `toml:"config_backup"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	pm := GetDefaultPathManager()
	return &Config{
		Server: ServerConfig{
			Port:     8080,
			LogLevel: "info",
			Daemon:   true,
		},
		APIs: []APIConfig{
			{
				ID:         "official-example",
				Name:       "Anthropic Official API (Example)",
				URL:        "https://api.anthropic.com",
				APIKey:     "sk-ant-your-api-key-here",
				IsActive:   false,
				Timeout:    30,
				RetryCount: 3,
			},
			{
				ID:         "proxy-example",
				Name:       "Proxy Service Example",
				URL:        "https://api.proxy-service.com",
				APIKey:     "your-proxy-api-key-here",
				IsActive:   false,
				Timeout:    30,
				RetryCount: 3,
			},
		},
		Settings: Settings{
			ActiveAPI:    "",
			LogFile:      pm.LogFile(),
			ConfigBackup: true,
		},
	}
}
