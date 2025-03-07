package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type VKPost struct {
	ID          int            `json:"id"`
	Text        string         `json:"text"`
	Date        int            `json:"date"`
	Attachments []VKAttachment `json:"attachments"`
}

type VKResponse struct {
	Response struct {
		Items []VKPost `json:"items"`
	} `json:"response"`
}

type VKClient struct {
	Token string
}

type VKAttachment struct {
	Type  string   `json:"type"`
	Photo *VKPhoto `json:"photo,omitempty"`
}

type VKPhoto struct {
	Sizes []VKPhotoSize `json:"sizes"`
}

type VKPhotoSize struct {
	URL string `json:"url"`
}

func NewVKClient(token string) *VKClient {
	return &VKClient{Token: token}
}

func (c *VKClient) GetWallPosts(owner_id string, count int) ([]VKPost, error) {
	baseURL := "https://api.vk.com/method/wall.get"
	params := url.Values{}
	params.Add("access_token", c.Token)
	params.Add("owner_id", owner_id)
	params.Add("count", fmt.Sprintf("%d", count))
	params.Add("v", "5.131")

	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("не удалось отправить HTTP запрос: %v", err)
	}
	defer resp.Body.Close()

	var result VKResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(result.Response.Items) == 0 {
		return nil, fmt.Errorf("не найдены новые посты(либо вообще их нет) owner_id=%s", owner_id)
	}
	fmt.Printf("VK API response: %+v\n", result.Response.Items)
	return result.Response.Items, nil
}
