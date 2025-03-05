package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TGClient struct {
	Token  string
	ChatID string
}

func NewTGClient(token, chatID string) *TGClient {
	return &TGClient{Token: token, ChatID: chatID}
}

func (c *TGClient) SendMessage(text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.Token)
	payload := map[string]string{
		"chat_id": c.ChatID,
		"text":    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("не удалось отправить сообщение: %s", resp.Status)
	}

	return nil
}
