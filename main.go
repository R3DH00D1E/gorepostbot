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
			} else if cachedPost != nil && cachedPost.LastModified < post.Date {
				modifiedPosts = append(modifiedPosts, post)
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

				cacheMutex.Lock()
				defer cacheMutex.Unlock()

				for _, tgMessageID := range tgMessageIDs {
					cache.AddPost(bin.Post{
						VKRecordID:   p.ID,
						TGMessageID:  tgMessageID,
						LastModified: p.Date,
						PhotoURLs:    photoURLs,
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

				photoURLs := processAttachments(p, tgClient)

				oldPhotoURLs := relatedPosts[0].PhotoURLs
				photoChanged := !arePhotoURLsEqual(oldPhotoURLs, photoURLs)

				if photoChanged && len(photoURLs) > 0 {
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
				}

				parts := utils.SplitText(p.Text, 4096)
				for i, relatedPost := range relatedPosts {
					if i < len(parts) {
						messageText := parts[i]
						if i == len(parts)-1 && len(photoURLs) > 0 {
							messageText = fmt.Sprintf("%s\n\n[%dx Photo]", messageText, len(photoURLs))
						}

						err := tgClient.EditMessageWithEditMark(relatedPost.TGMessageID, messageText, p.Date)
						if err != nil {
							log.Printf("Не удалось обновить сообщение %d: %v", relatedPost.TGMessageID, err)
						} else {
							log.Printf("Обновлено сообщение %d для поста %d с пометкой об изменении", relatedPost.TGMessageID, p.ID)
							cacheMutex.Lock()
							cache.UpdatePostWithPhotos(p.ID, p.Date, photoURLs)
							cacheMutex.Unlock()

							select {
							case cacheUpdates <- struct{}{}:
							default:
							}
						}
					}
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
