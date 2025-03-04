package lib

import (
	"os"
	"strconv"
)

type Targets struct {
	VkUser   string
	TgChatID int
}

type Config struct {
	TgToken   string
	VkToken   string
	Targets   Targets
	CacheFile string
	Debug     bool
	Interval  int
}

func Capture() Config {
	tgToken := os.Getenv("TG_TOKEN")
	vkToken := os.Getenv("VK_TOKEN")

	vkUser := os.Getenv("TARGET_USER")
	tgChatID, _ := strconv.Atoi(os.Getenv("TARGET_CHAT"))

	targets := Targets{
		VkUser:   vkUser,
		TgChatID: tgChatID,
	}

	cacheFile := os.Getenv("CACHE_FILE")
	debug := os.Getenv("DEBUG") == "1" || os.Getenv("DEBUG") == "true"
	interval, err := strconv.Atoi(os.Getenv("INTERVAL"))
	if err != nil {
		interval = 60 * 2
	}

	return Config{
		TgToken:   tgToken,
		VkToken:   vkToken,
		Targets:   targets,
		CacheFile: cacheFile,
		Debug:     debug,
		Interval:  interval,
	}
}
