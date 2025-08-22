package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	Redis RedisConfig `json:"redis"`
	UI    UIConfig    `json:"ui"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	Password string    `json:"password"`
	DB       int       `json:"db"`
	Timeout  int       `json:"timeout"`
	PoolSize int       `json:"pool_size"`
	TLS      TLSConfig `json:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	CertFile           string `json:"cert_file"`
	KeyFile            string `json:"key_file"`
	CAFile             string `json:"ca_file"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

// UIConfig holds UI configuration
type UIConfig struct {
	Theme           string `json:"theme"`
	RefreshInterval int    `json:"refresh_interval"`
	MaxKeys         int    `json:"max_keys"`
	ShowMemory      bool   `json:"show_memory"`
	ShowTTL         bool   `json:"show_ttl"`
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			Timeout:  5000,
			PoolSize: 10,
			TLS: TLSConfig{
				Enabled:            false,
				CertFile:           "",
				KeyFile:            "",
				CAFile:             "",
				InsecureSkipVerify: false,
			},
		},
		UI: UIConfig{
			Theme:           "default",
			RefreshInterval: 1000,
			MaxKeys:         1000,
			ShowMemory:      true,
			ShowTTL:         true,
		},
	}
}

// Load loads configuration from file or returns default
func Load() (*Config, error) {
	cfg := Default()

	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./config.json"
	}
	return filepath.Join(homeDir, ".redis-valkey-tui", "config.json")
}
