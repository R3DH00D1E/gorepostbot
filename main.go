package main

import (
	"fmt"
	"log"
	"time"

	"gorepostbot/bin"
	"gorepostbot/config"
	"gorepostbot/lib"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Выводим значения переменных окружения
	fmt.Println("TG_TOKEN:", cfg.TGToken)
	fmt.Println("VK_TOKEN:", cfg.VKToken)
	fmt.Println("CHAT_ID:", cfg.ChatID)
	fmt.Println("CACHE_FILE:", cfg.CacheFile)
	fmt.Println("POLL_INTERVAL:", cfg.PollInterval)
	fmt.Println("TARGET_USER:", cfg.TargetUser)

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
				log.Printf("Не удалось отправить сообщение: %v", err)
				continue
			}

			for _, attachment := range post.Attachments {
				switch attachment.Type {
				case "photo":
					if attachment.Photo != nil {
						lastSize := attachment.Photo.Sizes[len(attachment.Photo.Sizes)-1]
						err := tgClient.SendPhoto(lastSize.URL)
						if err != nil {
							log.Printf("Не удалось отправить фото: %v", err)
						}
					}
				case "video":
					log.Printf("Видео ещё не поддерживается: %+v", attachment)
				default:
					log.Printf("Неподдерживаемый вид: %s", attachment.Type)
				}
			}

			for _, tgMessageID := range tgMessageIDs {
				cache.AddPost(bin.Post{
					VKRecordID:   post.ID,
					TGMessageID:  tgMessageID,
					LastModified: post.Date,
				})
			}

			if cache.LastPostID == 0 {
				cache.LastPostID = posts[0].ID
			}
		}

		err = bin.SaveCache(cfg.CacheFile, cache)
		if err != nil {
			log.Printf("Не удалось сохранить кэш %v", err)
		}

		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		if err != nil {
			fmt.Println("Ошибка:", err)
		}
	}
}
