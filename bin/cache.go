package bin

import (
	"encoding/json"
	"fmt"
	"os"
)

type Post struct {
	VKRecordID   int `json:"vk_record_id"`
	TGMessageID  int `json:"tg_message_id"`
	LastModified int `json:"last_modified"`
}

type Cache struct {
	LastPostID int    `json:"last_post_id"`
	Posts      []Post `json:"posts"`
}

func LoadCache(path string) (*Cache, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Cache{
			LastPostID: 0,
			Posts:      []Post{},
		}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %v", err)
	}
	defer file.Close()

	var cache Cache
	err = json.NewDecoder(file).Decode(&cache)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %v", err)
	}

	if cache.Posts == nil {
		cache.Posts = []Post{}
	}

	return &cache, nil
}

func SaveCache(path string, cache *Cache) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(cache)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) AddPost(post Post) {
	c.Posts = append(c.Posts, post)
}

func (c *Cache) UpdatePost(vkRecordID int, lastModified int) bool {
	for i, post := range c.Posts {
		if post.VKRecordID == vkRecordID {
			c.Posts[i].LastModified = lastModified
			return true
		}
	}
	return false
}

func (c *Cache) FindPost(vkRecordID int) *Post {
	for _, post := range c.Posts {
		if post.VKRecordID == vkRecordID {
			return &post
		}
	}
	return nil
}
