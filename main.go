package main

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"gorepostbot/bin"
	"gorepostbot/config"
	"gorepostbot/lib"
	"gorepostbot/utils"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

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

			time.Sleep(5 * time.Second)
		}
	}()

	for {
		posts, err := vkClient.GetWallPosts(cfg.TargetUser, 20)
		if err != nil {
			log.Printf("Не удалось просмотреть посты: %v", err)
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			continue
		}

		if len(posts) == 0 {
			log.Println("Постов не обнаружено")
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			continue
		}

		sort.Slice(posts, func(i, j int) bool {
			return posts[i].ID < posts[j].ID
		})

		var cacheMutex sync.Mutex
		var wg sync.WaitGroup

		var newPosts []lib.VKPost
		var modifiedPosts []lib.VKPost

		currentTime := time.Now().Unix()
		twoDaysAgo := currentTime - (2 * 24 * 60 * 60)

		for _, post := range posts {
			if int64(post.Date) < twoDaysAgo {
				log.Printf("Пропуск поста ID %d: пост старше 2 дней (дата: %v)", post.ID, time.Unix(int64(post.Date), 0).Format("02.01.2006 15:04:05"))
				continue
			}

			cachedPost := cache.FindPost(post.ID)
			if cachedPost == nil && post.ID > cache.LastPostID {
				newPosts = append(newPosts, post)
			} else if cachedPost != nil {
				// Вычисляем хеш текущего содержимого поста
				currentTextHash := utils.ComputeTextHash(post.Text)

				// Получаем URL фотографий для вычисления хеша
				var currentPhotoURLs []string
				for _, attachment := range post.Attachments {
					if attachment.Type == "photo" && attachment.Photo != nil && len(attachment.Photo.Sizes) > 0 {
						lastSize := attachment.Photo.Sizes[len(attachment.Photo.Sizes)-1]
						currentPhotoURLs = append(currentPhotoURLs, lastSize.URL)
					}
				}
				currentPhotoHash := utils.ComputePhotoURLsHash(currentPhotoURLs)

				// Проверяем, изменился ли текст или фото
				if cachedPost.TextHash != currentTextHash || cachedPost.PhotoHash != currentPhotoHash {
					log.Printf("Обнаружены изменения в посте ID %d (текст: %v, фото: %v)",
						post.ID,
						cachedPost.TextHash != currentTextHash,
						cachedPost.PhotoHash != currentPhotoHash)
					modifiedPosts = append(modifiedPosts, post)
				}
			}
		}

		log.Printf("Найдено %d новых и %d модифицированных постов не старше 2 дней", len(newPosts), len(modifiedPosts))

		for _, post := range newPosts {
			wg.Add(1)
			go func(p lib.VKPost) {
				defer wg.Done()

				log.Printf("Обработка нового поста ID %d", p.ID)

				photoURLs := processAttachments(p, tgClient)

				tgMessageIDs, err := tgClient.SendMessage(p.Text, len(photoURLs))
				if err != nil {
					log.Printf("Не удалось отправить сообщение: %v", err)
					return
				}

				// Вычисляем хеши для нового поста
				textHash := utils.ComputeTextHash(p.Text)
				photoHash := utils.ComputePhotoURLsHash(photoURLs)

				cacheMutex.Lock()
				defer cacheMutex.Unlock()

				for _, tgMessageID := range tgMessageIDs {
					cache.AddPost(bin.Post{
						VKRecordID:   p.ID,
						TGMessageID:  tgMessageID,
						LastModified: p.Date,
						PhotoURLs:    photoURLs,
						TextHash:     textHash,
						PhotoHash:    photoHash,
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

		for _, post := range modifiedPosts {
			wg.Add(1)
			go func(p lib.VKPost) {
				defer wg.Done()

				log.Printf("Обработка измененного поста ID %d", p.ID)

				relatedPosts := cache.FindPostsByVKID(p.ID)
				if len(relatedPosts) == 0 {
					log.Printf("Не найдены связанные посты для ID %d", p.ID)
					return
				}

				// Получаем URL фотографий
				var photoURLs []string
				for _, attachment := range p.Attachments {
					if attachment.Type == "photo" && attachment.Photo != nil && len(attachment.Photo.Sizes) > 0 {
						lastSize := attachment.Photo.Sizes[len(attachment.Photo.Sizes)-1]
						photoURLs = append(photoURLs, lastSize.URL)
					}
				}

				// Вычисляем новые хеши
				newTextHash := utils.ComputeTextHash(p.Text)
				newPhotoHash := utils.ComputePhotoURLsHash(photoURLs)

				// Проверяем, что изменилось
				oldPost := relatedPosts[0]

				// Если хеши не были сохранены ранее (старый кеш), вычисляем их из сохраненных данных
				if oldPost.TextHash == "" {
					// Для старых записей без хеша - нужно инициализировать хеши
					cacheMutex.Lock()
					cache.UpdatePostHashes(p.ID, newTextHash, newPhotoHash)
					cacheMutex.Unlock()
					log.Printf("Инициализированы хеши для существующего поста %d", p.ID)
					return // Пропускаем обновление, т.к. это первая инициализация
				}

				textChanged := oldPost.TextHash != newTextHash
				photoChanged := oldPost.PhotoHash != newPhotoHash

				log.Printf("Изменения в посте %d: текст=%v, фото=%v (старый хеш текста: %s, новый: %s)",
					p.ID, textChanged, photoChanged, oldPost.TextHash[:8], newTextHash[:8])

				// Обрабатываем изменения в фотографиях
				if photoChanged {
					if len(photoURLs) > 0 {
						log.Printf("Обнаружены изменения в фотографиях для поста %d", p.ID)
						if len(photoURLs) > 1 {
							err := tgClient.SendMediaGroup(photoURLs)
							if err != nil {
								log.Printf("Не удалось отправить обновленную группу фото: %v", err)
							} else {
								log.Printf("Отправлена обновленная группа из %d фото для поста %d", len(photoURLs), p.ID)
							}
						} else {
							err := tgClient.SendPhoto(photoURLs[0])
							if err != nil {
								log.Printf("Не удалось отправить обновленное фото: %v", err)
							} else {
								log.Printf("Отправлено обновленное фото для поста %d", p.ID)
							}
						}
					} else {
						log.Printf("Фотографии были удалены из поста %d", p.ID)
					}
				}

				// Обрабатываем изменения в тексте
				if textChanged {
					parts := utils.SplitText(p.Text, 4096)
					for i, relatedPost := range relatedPosts {
						if i < len(parts) {
							messageText := parts[i]
							// Добавляем информацию о количестве фото в последнюю часть
							if i == len(parts)-1 && len(photoURLs) > 0 {
								messageText = fmt.Sprintf("%s\n\n[%dx Photo]", messageText, len(photoURLs))
							}

							err := tgClient.EditMessageWithEditMark(relatedPost.TGMessageID, messageText, p.Date)
							if err != nil {
								log.Printf("Не удалось обновить сообщение %d: %v", relatedPost.TGMessageID, err)
							} else {
								log.Printf("Обновлено сообщение %d для поста %d с пометкой об изменении", relatedPost.TGMessageID, p.ID)
							}
						}
					}
				}

				// Обновляем кеш с новыми хешами
				cacheMutex.Lock()
				cache.UpdatePostWithPhotos(p.ID, p.Date, photoURLs)
				cache.UpdatePostHashes(p.ID, newTextHash, newPhotoHash)
				cacheMutex.Unlock()

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

func processAttachments(p lib.VKPost, tgClient *lib.TGClient) []string {
	var photoURLs []string

	if len(p.Attachments) > 0 {
		var attachWg sync.WaitGroup
		var urlsMutex sync.Mutex

		for _, attachment := range p.Attachments {
			attachWg.Add(1)
			go func(att lib.VKAttachment) {
				defer attachWg.Done()

				switch att.Type {
				case "photo":
					if att.Photo != nil && len(att.Photo.Sizes) > 0 {
						lastSize := att.Photo.Sizes[len(att.Photo.Sizes)-1]
						urlsMutex.Lock()
						photoURLs = append(photoURLs, lastSize.URL)
						urlsMutex.Unlock()
					}
				case "video":
					log.Printf("Видео ещё не поддерживается: %+v", att)
				default:
					log.Printf("Неподдерживаемый вид: %s", att.Type)
				}
			}(attachment)
		}

		attachWg.Wait()

		if len(photoURLs) > 0 {
			if len(photoURLs) > 1 {
				err := tgClient.SendMediaGroup(photoURLs)
				if err != nil {
					log.Printf("Не удалось отправить группу фото: %v", err)
				} else {
					log.Printf("Отправлена группа из %d фото для поста %d", len(photoURLs), p.ID)
				}
			} else {
				err := tgClient.SendPhoto(photoURLs[0])
				if err != nil {
					log.Printf("Не удалось отправить фото: %v", err)
				} else {
					log.Printf("Отправлено фото для поста %d", p.ID)
				}
			}
		}
	}

	return photoURLs
}

func arePhotoURLsEqual(old, new []string) bool {
	if len(old) != len(new) {
		return false
	}

	urlMap := make(map[string]bool)
	for _, url := range old {
		urlMap[url] = true
	}

	for _, url := range new {
		if !urlMap[url] {
			return false
		}
	}

	return true
}
