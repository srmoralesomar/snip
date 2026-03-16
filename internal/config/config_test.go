package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/srmoralesomar/snip/internal/config"
)

func newViper() *viper.Viper {
	return viper.New()
}

func TestLoad_Defaults(t *testing.T) {
	// Use a viper pointed at a non-existent config directory so no file is read.
	v := newViper()
	v.AddConfigPath(t.TempDir()) // empty dir — no config.yaml

	cfg, err := config.Load(v)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.MaxHistory != config.DefaultMaxHistory {
		t.Errorf("MaxHistory = %d, want %d", cfg.MaxHistory, config.DefaultMaxHistory)
	}
	if cfg.PollIntervalMs != config.DefaultPollIntervalMs {
		t.Errorf("PollIntervalMs = %d, want %d", cfg.PollIntervalMs, config.DefaultPollIntervalMs)
	}
	if cfg.StoragePath == "" {
		t.Error("StoragePath should not be empty")
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	yaml := "max_history: 100\npoll_interval_ms: 250\nstorage_path: /tmp/test.db\n"
	if err := os.WriteFile(cfgFile, []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	v := newViper()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)

	cfg, err := config.Load(v)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.MaxHistory != 100 {
		t.Errorf("MaxHistory = %d, want 100", cfg.MaxHistory)
	}
	if cfg.PollIntervalMs != 250 {
		t.Errorf("PollIntervalMs = %d, want 250", cfg.PollIntervalMs)
	}
	if cfg.StoragePath != "/tmp/test.db" {
		t.Errorf("StoragePath = %q, want /tmp/test.db", cfg.StoragePath)
	}
}

func TestLoad_FlagOverridesFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	yaml := "max_history: 100\npoll_interval_ms: 250\n"
	if err := os.WriteFile(cfgFile, []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	v := newViper()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)

	// Simulate a CLI flag override by setting directly on viper.
	v.Set("max_history", 999)

	cfg, err := config.Load(v)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.MaxHistory != 999 {
		t.Errorf("MaxHistory = %d, want 999 (flag override)", cfg.MaxHistory)
	}
	// poll_interval_ms should still come from file.
	if cfg.PollIntervalMs != 250 {
		t.Errorf("PollIntervalMs = %d, want 250 (from file)", cfg.PollIntervalMs)
	}
}

func TestLoad_MissingFileUsesDefaults(t *testing.T) {
	v := newViper()
	// Point at an empty temp dir — no config.yaml exists.
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(t.TempDir())

	cfg, err := config.Load(v)
	if err != nil {
		t.Fatalf("Load() should not error when config file is missing: %v", err)
	}

	if cfg.MaxHistory != config.DefaultMaxHistory {
		t.Errorf("MaxHistory = %d, want default %d", cfg.MaxHistory, config.DefaultMaxHistory)
	}
}
