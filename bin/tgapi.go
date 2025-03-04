package lib

import (
	"context"
	"errors"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type TelegramService struct {
	bot *telego.Bot
}

func NewTelegramService(token string) (*TelegramService, error) {
	if token == "" {
		return nil, errors.New("telegram token is required")
	}

	bot, err := telego.NewBot(token)
	if err != nil {
		return nil, err
	}

	return &TelegramService{
		bot: bot,
	}, nil
}

func (tg *TelegramService) SendPost(chatID int, post VKPost) (int, error) {
	messageText := post.Text
	if messageText == "" {
		messageText = "Новый пост из ВКонтакте"
	}

	date := fmt.Sprintf("\n\nДата публикации: %s",
		fmt.Sprintf("<i>%s</i>", formatUnixTime(post.Date)))
	messageText += date

	vkService := &VKService{}
	attachments := vkService.GetAttachments(post)

	var sentMessage *telego.Message
	var err error
	ctx := context.TODO()

	if len(attachments) > 0 {
		photoURL := attachments[0]
		params := tu.Photo(
			tu.ID(int64(chatID)),
			tu.FileFromURL(photoURL),
		).WithCaption(messageText).WithParseMode(telego.ModeHTML)

		sentMessage, err = tg.bot.SendPhoto(ctx, params)
		if err != nil {
			msgParams := tu.Message(
				tu.ID(int64(chatID)),
				messageText,
			).WithParseMode(telego.ModeHTML)

			sentMessage, err = tg.bot.SendMessage(ctx, msgParams)
		}
	} else {
		msgParams := tu.Message(
			tu.ID(int64(chatID)),
			messageText,
		).WithParseMode(telego.ModeHTML)

		sentMessage, err = tg.bot.SendMessage(ctx, msgParams)
	}

	if err != nil {
		return 0, err
	}

	return sentMessage.MessageID, nil
}

func formatUnixTime(unixTime int) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		1970+unixTime/(60*60*24*365),
		(unixTime/(60*60*24*30))%12+1,
		(unixTime/(60*60*24))%30+1,
		(unixTime/(60*60))%24,
		(unixTime/60)%60,
		unixTime%60,
	)
}
