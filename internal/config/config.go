package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Storage      StorageConfig      `yaml:"storage"`
	Defaults     DefaultsConfig     `yaml:"defaults"`
	Notifications NotificationsConfig `yaml:"notifications"`
	Rules        RulesConfig        `yaml:"rules"`
	Items        []ItemConfig       `yaml:"items"`
}

type StorageConfig struct {
	Driver string `yaml:"driver"`
	Path   string `yaml:"path"`
}

type DefaultsConfig struct {
	Currency      string        `yaml:"currency"`
	Timezone      string        `yaml:"timezone"`
	UserAgent     string        `yaml:"user_agent"`
	Retry         RetryConfig   `yaml:"retry"`
	HTTPTimeout   time.Duration `yaml:"http_timeout_sec"`
	CacheTTL      time.Duration `yaml:"cache_ttl_min"`
	Headless      HeadlessConfig `yaml:"headless"`
}

type RetryConfig struct {
	Attempts     int           `yaml:"attempts"`
	BaseDelay    time.Duration `yaml:"base_delay_ms"`
	MaxDelay     time.Duration `yaml:"max_delay_ms"`
}

type HeadlessConfig struct {
	Enabled   bool   `yaml:"enabled"`
	WaitUntil string `yaml:"wait_until"`
}

type NotificationsConfig struct {
	Email    EmailConfig    `yaml:"email"`
	Telegram TelegramConfig `yaml:"telegram"`
	Slack    SlackConfig    `yaml:"slack"`
	Ntfy     NtfyConfig     `yaml:"ntfy"`
}

type EmailConfig struct {
	Enabled bool     `yaml:"enabled"`
	From    string   `yaml:"from"`
	To      []string `yaml:"to"`
}

type TelegramConfig struct {
	Enabled bool   `yaml:"enabled"`
	ChatID  string `yaml:"chat_id"`
}

type SlackConfig struct {
	Enabled bool   `yaml:"enabled"`
	Webhook string `yaml:"webhook"`
}

type NtfyConfig struct {
	Enabled bool   `yaml:"enabled"`
	Topic   string `yaml:"topic"`
}

type RulesConfig struct {
	PercentDrop  float64 `yaml:"percent_drop"`
	TargetPrice  *float64 `yaml:"target_price"`
}

type ItemConfig struct {
	ID           string  `yaml:"id"`
	Name         string  `yaml:"name"`
	URL          string  `yaml:"url"`
	Provider     string  `yaml:"provider"`
	Selector     string  `yaml:"selector"`
	Currency     string  `yaml:"currency"`
	TargetPrice  *float64 `yaml:"target_price"`
	PercentDrop  *float64 `yaml:"percent_drop"`
	Schedule     string  `yaml:"schedule"`
	Regex        string  `yaml:"regex,omitempty"`
	Attr         string  `yaml:"attr,omitempty"`
	Command      string  `yaml:"command,omitempty"`
}

func Load(path string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Storage.Driver == "" {
		cfg.Storage.Driver = "sqlite"
	}
	if cfg.Storage.Path == "" {
		cfg.Storage.Path = "./data/trek.db"
	}
	if cfg.Defaults.Currency == "" {
		cfg.Defaults.Currency = "USD"
	}
	if cfg.Defaults.Timezone == "" {
		cfg.Defaults.Timezone = "UTC"
	}
	if cfg.Defaults.UserAgent == "" {
		cfg.Defaults.UserAgent = "PriceTrek/0.1 (+https://github.com/makalin/pricetrek)"
	}
	if cfg.Defaults.Retry.Attempts == 0 {
		cfg.Defaults.Retry.Attempts = 3
	}
	if cfg.Defaults.Retry.BaseDelay == 0 {
		cfg.Defaults.Retry.BaseDelay = 800 * time.Millisecond
	}
	if cfg.Defaults.Retry.MaxDelay == 0 {
		cfg.Defaults.Retry.MaxDelay = 7 * time.Second
	}
	if cfg.Defaults.HTTPTimeout == 0 {
		cfg.Defaults.HTTPTimeout = 20 * time.Second
	}
	if cfg.Defaults.CacheTTL == 0 {
		cfg.Defaults.CacheTTL = 30 * time.Minute
	}
	if cfg.Defaults.Headless.WaitUntil == "" {
		cfg.Defaults.Headless.WaitUntil = "networkidle"
	}
	if cfg.Rules.PercentDrop == 0 {
		cfg.Rules.PercentDrop = 8.0
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}