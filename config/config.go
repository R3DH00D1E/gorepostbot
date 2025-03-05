package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	TGToken      string
	VKToken      string
	TargetUser   string
	CacheFile    string
	PollInterval int
	ChatID       string
}

func LoadConfig() (*Config, error) {
	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine user home directory: %w", err)
		}
		envPath = filepath.Join(homeDir, "repostbot", "repostbot.env")
	}

	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			return nil, fmt.Errorf("failed to load .env file (%s): %w", envPath, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to access .env file (%s): %w", envPath, err)
	}

	cfg := Config{
		TGToken:      os.Getenv("TG_TOKEN"),
		VKToken:      os.Getenv("VK_TOKEN"),
		TargetUser:   os.Getenv("TARGET_USER"),
		CacheFile:    os.Getenv("CACHE_FILE"),
		PollInterval: getIntEnv("POLL_INTERVAL"),
		ChatID:       os.Getenv("CHAT_ID"),
	}

	if cfg.TGToken == "" || cfg.VKToken == "" || cfg.TargetUser == "" || cfg.CacheFile == "" || cfg.ChatID == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return &cfg, nil
}

func getIntEnv(key string) int {
	value := os.Getenv(key)
	if value == "" {
		return 0
	}
	var intValue int
	fmt.Sscanf(value, "%d", &intValue)
	return intValue
}
