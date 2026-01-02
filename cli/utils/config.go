package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	configDirName  = ".envoy"
	configFileName = "config.json"
)

var (
	homeDir          string
	configFilePath   string
	DefaultServerURL = "https://envoy.webiliti.com" // Production server URL
)

func init() {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			if runtime.GOOS == "windows" {
				homeDir = os.Getenv("USERPROFILE")
			}
		}
	}
	configFilePath = filepath.Join(homeDir, configDirName, configFileName)
}

type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
}

func GetConfigDir() string {
	return filepath.Join(homeDir, configDirName)
}

func EnsureConfigDir() error {
	configDir := GetConfigDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return os.MkdirAll(configDir, 0700)
	}
	return nil
}

func Load() (*Config, error) {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func GetToken() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.Token, nil
}

func SetToken(token string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.Token = token
	return Save(cfg)
}

func GetServerURL() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	if cfg.ServerURL != "" {
		return cfg.ServerURL, nil
	}

	if envURL := os.Getenv("ENVOY_SERVER_URL"); envURL != "" {
		return envURL, nil
	}

	return DefaultServerURL, nil
}

func ClearToken() error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.Token = ""
	return Save(cfg)
}
