package config

import (
	"encoding/json"
	"errors"
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

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, errors.New("не удалось открыть файл конфигурации: " + err.Error())
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, errors.New("ошибка разбора JSON: " + err.Error())
	}

	if config.VKToken == "" {
		return nil, errors.New("не указан токен VK")
	}
	if config.TGToken == "" {
		return nil, errors.New("не указан токен Telegram")
	}
	if config.ChatID == "" {
		return nil, errors.New("не указан ID чата")
	}
	if config.TargetUser == "" {
		return nil, errors.New("не указан ID пользователя ВК")
	}

	if config.PollInterval <= 0 {
		config.PollInterval = 10
	}
	if config.CacheFile == "" {
		config.CacheFile = "cache.json"
	}

	return &config, nil
}
