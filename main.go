package main

import (
	"log"
	"time"

	"gorepostbot/bin"
	"gorepostbot/config"
	"gorepostbot/lib"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	vkClient := lib.NewVKClient(cfg.VKToken)
	tgClient := lib.NewTGClient(cfg.TGToken, cfg.ChatID)

	cache, err := bin.LoadCache(cfg.CacheFile)
	if err != nil {
		log.Fatalf("Не удалось загрузить кэш: %v", err)
	}
	log.Printf("Загружен кэш: %+v", cache)

	for {
		posts, err := vkClient.GetWallPosts(cfg.TargetUser, 10)
		if err != nil {
			log.Printf("Не удалось просмотреть посты: %v", err)
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			continue
		}

		for _, post := range posts {
			if post.ID <= cache.LastPostID {
				continue
			}

			tgMessageIDs, err := tgClient.SendMessage(post.Text)
			if err != nil {
				log.Printf("Failed to send message: %v", err)
				continue
			}

			for _, attachment := range post.Attachments {
				if attachment.Type == "photo" && attachment.Photo != nil {
					lastSize := attachment.Photo.Sizes[len(attachment.Photo.Sizes)-1]
					err := tgClient.SendPhoto(lastSize.URL)
					if err != nil {
						log.Printf("Failed to send photo: %v", err)
					}
				}
			}

			for _, tgMessageID := range tgMessageIDs {
				cache.AddPost(bin.Post{
					VKRecordID:   post.ID,
					TGMessageID:  tgMessageID,
					LastModified: post.Date,
				})
			}

			if post.ID > cache.LastPostID {
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
