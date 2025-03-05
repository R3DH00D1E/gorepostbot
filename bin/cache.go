package bin

import (
	"encoding/json"
	"os"
)

type Cache struct {
	LastPostID int         `json:"last_post_id"`
	Posts      map[int]int `json:"posts"`
}

func LoadCache(path string) (*Cache, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {

		return &Cache{
			LastPostID: 0,
			Posts:      make(map[int]int),
		}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cache Cache
	err = json.NewDecoder(file).Decode(&cache)
	if err != nil {
		return nil, err
	}

	if cache.Posts == nil {
		cache.Posts = make(map[int]int)
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
