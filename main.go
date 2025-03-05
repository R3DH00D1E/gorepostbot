package main

import (
	"log"
	"time"

	"gorepostbot/bin"
	"gorepostbot/config"
	"gorepostbot/lib"
)

func main() {
	//конфигурация
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	//инициализация клиентов
	vkClient := lib.NewVKClient(cfg.VKToken)
	tgClient := lib.NewTGClient(cfg.TGToken, cfg.ChatID)

	//загрузка кэша
	cache, err := bin.LoadCache(cfg.CacheFile)
	if err != nil {
		log.Printf("Не удалось загрузить конфиг: %v", err)
		cache = &bin.Cache{}
	}

	//цикл
	for {
		posts, err := vkClient.GetWallPosts(cfg.TargetUser, 10)
		if err != nil {
			log.Printf("Ну удалось прочитать новые посты %v", err)
			continue
		}

		for _, post := range posts {
			if post.ID > cache.LastPostID {
				err := tgClient.SendMessage(post.Text)
				if err != nil {
					log.Printf("Не удалось отправить сообщение %v", err)
				}
				cache.LastPostID = post.ID
			}
		}

		// Сохранение кэша
		err = bin.SaveCache(cfg.CacheFile, cache)
		if err != nil {
			log.Printf("Не удалось сохранить кэш %v", err)
		}

		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
	}
}
