package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	TGToken      string `json:"TG_TOKEN"`
	VKToken      string `json:"VK_TOKEN"`
	TargetUser   string `json:"TARGET_USER"`
	CacheFile    string `json:"CACHE_FILE"`
	PollInterval int    `json:"POLL_INTERVAL"`
	ChatID       string `json:"CHAT_ID"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, "repostbot", "config.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file (%s): %w", configPath, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.TGToken == "" || cfg.VKToken == "" || cfg.TargetUser == "" || cfg.CacheFile == "" || cfg.ChatID == "" {
		return nil, fmt.Errorf("missing required configuration fields")
	}

	return &cfg, nil
}
