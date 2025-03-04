package lib

import (
	"encoding/json"
	"os"
)

type Post struct {
	VkRecordID  int `json:"vk_record_id"`
	TgMessageID int `json:"tg_message_id"`
	LastModify  int `json:"last_modify"`
}

type Cache struct {
	LastRecordID int    `json:"last_record_id"`
	Posts        []Post `json:"posts"`
}

func LoadFromFile(path string) (Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Cache{}, err
	}
	var cache Cache
	err = json.Unmarshal(data, &cache)
	return cache, err
}

func SaveToFile(path string, cache Cache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func AddPost(cache *Cache, post Post) {
	cache.LastRecordID = post.VkRecordID
	cache.Posts = append([]Post{post}, cache.Posts...)
}

func EditPost(cache *Cache, post Post) {
	for i, p := range cache.Posts {
		if p.VkRecordID == post.VkRecordID {
			cache.Posts[i] = post
			break
		}
	}
}

func IteratePosts(cache *Cache, f func(Post)) {
	for _, p := range cache.Posts {
		f(p)
	}
}

func GetPosts(cache *Cache) []Post {
	return cache.Posts
}
