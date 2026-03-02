package utils

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

func SplitText(text string, maxLength int) []string {
	var parts []string
	words := strings.Fields(text)
	currentPart := ""

	for _, word := range words {
		if len(currentPart)+len(word)+1 > maxLength {
			parts = append(parts, currentPart)
			currentPart = ""
		}
		if currentPart != "" {
			currentPart += " "
		}
		currentPart += word
	}

	if currentPart != "" {
		parts = append(parts, currentPart)
	}

	return parts
}

// ComputeTextHash вычисляет SHA256 хеш текста
func ComputeTextHash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// ComputePhotoURLsHash вычисляет хеш списка URL фотографий
func ComputePhotoURLsHash(urls []string) string {
	// Сортируем URL для консистентности
	sortedURLs := make([]string, len(urls))
	copy(sortedURLs, urls)
	sort.Strings(sortedURLs)
	
	combined := strings.Join(sortedURLs, "|")
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}
