package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultMaxHistory     = 500
	DefaultPollIntervalMs = 500
)

// Config holds the effective snip configuration.
type Config struct {
	MaxHistory     int    `mapstructure:"max_history"`
	PollIntervalMs int    `mapstructure:"poll_interval_ms"`
	StoragePath    string `mapstructure:"storage_path"`
}

// DefaultStoragePath returns the default path for history.db.
func DefaultStoragePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".snip", "history.db"), nil
}

// DefaultConfigPath returns the default config file path (~/.snip/config.yaml).
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".snip", "config.yaml"), nil
}

// Load reads configuration from ~/.snip/config.yaml (if present) and returns
// a Config populated with effective values (defaults < file < env).
// Missing config file is not an error.
func Load(v *viper.Viper) (*Config, error) {
	storagePath, err := DefaultStoragePath()
	if err != nil {
		return nil, err
	}

	v.SetDefault("max_history", DefaultMaxHistory)
	v.SetDefault("poll_interval_ms", DefaultPollIntervalMs)
	v.SetDefault("storage_path", storagePath)

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home dir: %w", err)
	}

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(filepath.Join(home, ".snip"))

	// Ignore "file not found" — defaults are used instead.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
