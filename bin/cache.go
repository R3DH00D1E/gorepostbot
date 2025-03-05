package bin

import (
	"encoding/json"
	"os"
)

type Cache struct {
	LastPostID int `json:"last_post_id"`
}

func LoadCache(path string) (*Cache, error) {
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

	return &cache, nil
}

func SaveCache(path string, cache *Cache) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cache)
	if err != nil {
		return err
	}

	return nil
}
