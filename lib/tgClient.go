package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorepostbot/utils"
	"net/http"
)

type TGClient struct {
	Token  string
	ChatID string
}

func NewTGClient(token, chatID string) *TGClient {
	return &TGClient{Token: token, ChatID: chatID}
}

func (c *TGClient) SendMessage(text string) ([]int, error) {

	parts := utils.SplitText(text, 4096)
	var messageIDs []int

	for _, part := range parts {
		url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.Token)
		payload := map[string]string{
			"chat_id": c.ChatID,
			"text":    part,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %v", err)
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("failed to send HTTP request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var respBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&respBody)
			return nil, fmt.Errorf("failed to send message: status=%d, body=%v", resp.StatusCode, respBody)
		}

		var result struct {
			Result struct {
				MessageID int `json:"message_id"`
			} `json:"result"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}

		messageIDs = append(messageIDs, result.Result.MessageID)
	}

	return messageIDs, nil
}

func (c *TGClient) SendPhoto(photoURL string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", c.Token)
	payload := map[string]string{
		"chat_id": c.ChatID,
		"photo":   photoURL,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)
		return fmt.Errorf("failed to send photo: status=%d, body=%v", resp.StatusCode, respBody)
	}
	fmt.Printf("Sending message to Telegram: %+v\n", payload)
	return nil
}

func (c *TGClient) SendMediaGroup(photoURLs []string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMediaGroup", c.Token)

	media := make([]map[string]string, len(photoURLs))
	for i, photoURL := range photoURLs {
		media[i] = map[string]string{
			"type":  "photo",
			"media": photoURL,
		}
	}

	payload := map[string]interface{}{
		"chat_id": c.ChatID,
		"media":   media,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)
		return fmt.Errorf("failed to send media group: status=%d, body=%v", resp.StatusCode, respBody)
	}
	fmt.Printf("Sending message to Telegram: %+v\n", payload)
	return nil
}

func (c *TGClient) EditMessage(messageID int, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", c.Token)
	payload := map[string]interface{}{
		"chat_id":    c.ChatID,
		"message_id": messageID,
		"text":       text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)
		return fmt.Errorf("failed to edit message: status=%d, body=%v", resp.StatusCode, respBody)
	}
	fmt.Printf("Sending message to Telegram: %+v\n", payload)
	return nil
}
