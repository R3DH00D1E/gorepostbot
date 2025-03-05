package main

import (
	"log"
	"time"

	"gorepostbot/bin"
	"gorepostbot/config"
	"gorepostbot/lib"
)

func main() {

	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	vkClient := lib.NewVKClient(cfg.VKToken)
	tgClient := lib.NewTGClient(cfg.TGToken, cfg.ChatID)

	cache, err := bin.LoadCache(cfg.CacheFile)
	if err != nil {
		log.Printf("Не удалось загрузить конфиг: %v", err)
		cache = &bin.Cache{}
	}

	var photoURLs []string

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
					log.Printf("Failed to send message: %v", err)
				}

				for _, attachment := range post.Attachments {
					if attachment.Type == "photo" && attachment.Photo != nil {
						lastSize := attachment.Photo.Sizes[len(attachment.Photo.Sizes)-1]
						err := tgClient.SendPhoto(lastSize.URL)
						if err != nil {
							log.Printf("Failed to send photo: %v", err)
						}
						photoURLs = append(photoURLs, lastSize.URL)
					}
				}

				if len(photoURLs) > 0 {
					err := tgClient.SendMediaGroup(photoURLs)
					if err != nil {
						log.Printf("Failed to send media group: %v", err)
					}
				}
				cache.LastPostID = post.ID
			}
		}

		err = bin.SaveCache(cfg.CacheFile, cache)
		if err != nil {
			log.Printf("Не удалось сохранить кэш %v", err)
		}

		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
	}
}
