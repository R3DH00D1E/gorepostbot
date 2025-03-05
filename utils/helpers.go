package utils

import "strings"

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
