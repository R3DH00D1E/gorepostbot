package config

import (
	"fmt"
	"os"
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
