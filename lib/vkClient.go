package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type VKPost struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Date int    `json:"date"`
}

type VKResponse struct {
	Response struct {
		Items []VKPost `json:"items"`
	} `json:"response"`
}

type VKClient struct {
	Token string
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
		return nil, err
	}
	defer resp.Body.Close()

	var result VKResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Response.Items, nil
}
