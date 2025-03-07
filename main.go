package main

import (
	"fmt"
	"log"
	"sync"
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

	cacheUpdates := make(chan struct{}, 10)

	go func() {
		for range cacheUpdates {
			err := bin.SaveCache(cfg.CacheFile, cache)
			if err != nil {
				log.Printf("Не удалось сохранить кэш: %v", err)
			} else {
				log.Printf("Кэш успешно сохранен")
			}

			// Ограничиваем частоту сохранений
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		posts, err := vkClient.GetWallPosts(cfg.TargetUser, 10)
		if err != nil {
			log.Printf("Не удалось просмотреть посты: %v", err)
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			continue
		}

		if len(posts) == 0 {
			log.Println("Новых постов не обнаружено")
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			continue
		}

		var cacheMutex sync.Mutex
		var wg sync.WaitGroup

		for _, post := range posts {
			if post.ID <= cache.LastPostID {
				continue
			}

			wg.Add(1)
			go func(p lib.VKPost) {
				defer wg.Done()

				log.Printf("Обработка поста ID %d", p.ID)
				tgMessageIDs, err := tgClient.SendMessage(p.Text)
				if err != nil {
					log.Printf("Не удалось отправить сообщение: %v", err)
					return
				}

				if len(p.Attachments) > 0 {
					var attachWg sync.WaitGroup

					for _, attachment := range p.Attachments {
						attachWg.Add(1)
						go func(att lib.VKAttachment) {
							defer attachWg.Done()

							switch att.Type {
							case "photo":
								if att.Photo != nil {
									lastSize := att.Photo.Sizes[len(att.Photo.Sizes)-1]
									err := tgClient.SendPhoto(lastSize.URL)
									if err != nil {
										log.Printf("Не удалось отправить фото: %v", err)
									} else {
										log.Printf("Отправлено фото для поста %d", p.ID)
									}
								}
							case "video":
								log.Printf("Видео ещё не поддерживается: %+v", att)
							default:
								log.Printf("Неподдерживаемый вид: %s", att.Type)
							}
						}(attachment)
					}

					attachWg.Wait()
				}

				cacheMutex.Lock()
				defer cacheMutex.Unlock()

				for _, tgMessageID := range tgMessageIDs {
					cache.AddPost(bin.Post{
						VKRecordID:   p.ID,
						TGMessageID:  tgMessageID,
						LastModified: p.Date,
					})
				}

				if p.ID > cache.LastPostID {
					cache.LastPostID = p.ID
					log.Printf("Обновлен LastPostID: %d", cache.LastPostID)
				}

				select {
				case cacheUpdates <- struct{}{}:
				default:
				}
			}(post)
		}

		wg.Wait()

		select {
		case cacheUpdates <- struct{}{}:
		default:
		}

		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
	}
}
