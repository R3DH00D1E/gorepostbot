package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	VKToken      string `json:"vk_token"`
	TGToken      string `json:"tg_token"`
	ChatID       string `json:"chat_id"`
	PollInterval int    `json:"poll_interval"`
	TargetUser   string `json:"target_user"`
	CacheFile    string `json:"cache_file"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
