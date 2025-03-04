package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type VKPost struct {
	ID          int            `json:"id"`
	Text        string         `json:"text"`
	Date        int            `json:"date"`
	Attachments []VKAttachment `json:"attachments"`
}

type VKAttachment struct {
	Type  string  `json:"type"`
	Photo VKPhoto `json:"photo,omitempty"`
	Video VKVideo `json:"video,omitempty"`
	Doc   VKDoc   `json:"doc,omitempty"`
}

type VKPhoto struct {
	Sizes []VKPhotoSize `json:"sizes"`
}

type VKPhotoSize struct {
	Type   string `json:"type"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type VKVideo struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Image       []VKVideoImage `json:"image"`
	AccessKey   string         `json:"access_key"`
	OwnerID     int            `json:"owner_id"`
	ID          int            `json:"id"`
}

type VKVideoImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type VKDoc struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Size  int    `json:"size"`
	Ext   string `json:"ext"`
}

type VKResponse struct {
	Response VKItems `json:"response"`
}

type VKItems struct {
	Items []VKPost `json:"items"`
	Count int      `json:"count"`
}

type VKService struct {
	token      string
	httpClient *http.Client
}

func NewVKService(token string) (*VKService, error) {
	if token == "" {
		return nil, errors.New("VK token is required")
	}

	return &VKService{
		token:      token,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (vk *VKService) GetWallPosts(userID string, lastID int) ([]VKPost, error) {
	apiURL := "https://api.vk.com/method/wall.get"

	params := url.Values{}
	params.Add("access_token", vk.token)
	params.Add("domain", userID)
	params.Add("count", "10")
	params.Add("v", "5.131")

	resp, err := vk.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var vkResponse VKResponse
	if err := json.NewDecoder(resp.Body).Decode(&vkResponse); err != nil {
		return nil, err
	}

	var result []VKPost
	for _, post := range vkResponse.Response.Items {
		if post.ID > lastID {
			result = append(result, post)
		}
	}

	return result, nil
}

func (vk *VKService) GetBestPhotoURL(post VKPost) string {
	for _, attachment := range post.Attachments {
		if attachment.Type == "photo" {
			bestSize := attachment.Photo.Sizes[0]
			for _, size := range attachment.Photo.Sizes {
				if size.Width > bestSize.Width {
					bestSize = size
				}
			}
			return bestSize.URL
		}
	}
	return ""
}

func (vk *VKService) GetVideoURL(video VKVideo) string {
	return fmt.Sprintf("https://vk.com/video%d_%d?access_key=%s",
		video.OwnerID, video.ID, video.AccessKey)
}

func (vk *VKService) GetAttachments(post VKPost) []string {
	var attachments []string

	for _, attachment := range post.Attachments {
		switch attachment.Type {
		case "photo":
			url := vk.GetBestPhotoURL(post)
			if url != "" {
				attachments = append(attachments, url)
			}
		case "video":
			url := vk.GetVideoURL(attachment.Video)
			if url != "" {
				attachments = append(attachments, url)
			}
		case "doc":
			if attachment.Doc.URL != "" {
				attachments = append(attachments, attachment.Doc.URL)
			}
		}
	}

	return attachments
}
