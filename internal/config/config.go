package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultPort is the default server port if not specified in config.
	DefaultPort = 4000
	// DefaultConfigPath is the default configuration file path.
	DefaultConfigPath = "./config.toml"
)

// Load reads and parses the TOML configuration file at the given path.
func Load(path string) (*Config, error) {
	if err := validateConfigPath(path); err != nil {
		return nil, fmt.Errorf("config path validation failed: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func validateConfigPath(path string) error {
	if path == "" {
		return fmt.Errorf("config path cannot be empty")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", absPath)
	}

	return nil
}

func validateConfig(cfg *Config) error {
	if cfg.UseDirectory == "" {
		return fmt.Errorf("use_directory cannot be empty")
	}

	absDir, err := filepath.Abs(cfg.UseDirectory)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for use_directory: %w", err)
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return fmt.Errorf("failed to create uploads directory: %w", err)
	}
	cfg.UseDirectory = absDir

	if cfg.Port <= 0 || cfg.Port > 65535 {
		cfg.Port = DefaultPort
	}

	if len(cfg.Users) == 0 {
		return fmt.Errorf("at least one user must be configured")
	}

	for i, user := range cfg.Users {
		if user.Name == "" {
			return fmt.Errorf("user %d: name cannot be empty", i)
		}
		if user.Auth == "" {
			return fmt.Errorf("user %d: auth code cannot be empty", i)
		}
	}

	return nil
}
