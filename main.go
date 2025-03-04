package main

import (
	"gorepostbot/lib"
	"log"
	"os"
	"time"
)

func main() {
	config := lib.Capture()

	// Проверка обязательных настроек
	if config.TgToken == "" || config.VkToken == "" || config.Targets.VkUser == "" || config.Targets.TgChatID == 0 {
		log.Fatal("Missing required environment variables. Please set TG_TOKEN, VK_TOKEN, TARGET_USER, TARGET_CHAT")
	}

	if config.CacheFile == "" {
		config.CacheFile = "cache.json"
	}

	// Инициализация кэша
	var cache lib.Cache
	if _, err := os.Stat(config.CacheFile); os.IsNotExist(err) {
		cache = lib.Cache{
			LastRecordID: 0,
			Posts:        []lib.Post{},
		}
		if err := lib.SaveToFile(config.CacheFile, cache); err != nil {
			log.Fatal("Failed to create cache file:", err)
		}
	} else {
		var err error
		cache, err = lib.LoadFromFile(config.CacheFile)
		if err != nil {
			log.Fatal("Failed to load cache file:", err)
		}
	}

	// Инициализация сервисов
	vkService, err := lib.NewVKService(config.VkToken)
	if err != nil {
		log.Fatal("Failed to initialize VK service:", err)
	}

	tgService, err := lib.NewTelegramService(config.TgToken)
	if err != nil {
		log.Fatal("Failed to initialize Telegram service:", err)
	}

	log.Println("Bot started successfully!")
	if config.Debug {
		log.Println("Debug mode enabled")
	}

	// Основной цикл
	for {
		if config.Debug {
			log.Println("Checking for new posts...")
		}

		// Получение новых постов из ВК
		posts, err := vkService.GetWallPosts(config.Targets.VkUser, cache.LastRecordID)
		if err != nil {
			log.Println("Error getting VK posts:", err)
			time.Sleep(time.Duration(config.Interval) * time.Second)
			continue
		}

		// Обработка новых постов
		for _, post := range posts {
			if config.Debug {
				log.Println("Processing post ID:", post.ID)
			}

			// Отправка поста в Telegram
			messageID, err := tgService.SendPost(config.Targets.TgChatID, post)
			if err != nil {
				log.Println("Error sending post to Telegram:", err)
				continue
			}

			// Сохранение информации о посте в кэше
			lib.AddPost(&cache, lib.Post{
				VkRecordID:  post.ID,
				TgMessageID: messageID,
				LastModify:  int(time.Now().Unix()),
			})

			// Сохранение кэша
			if err := lib.SaveToFile(config.CacheFile, cache); err != nil {
				log.Println("Error saving cache:", err)
			}
		}

		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}
