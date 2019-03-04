package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

type Bot struct {
	tgbotapi.BotAPI
}

func NewBot(token string) *Bot {
	bot := &Bot{}
	bot.Token = token
	bot.Client = &http.Client{}

	return bot
}

func (b Bot) newMsg(chatId int64, text string) tgbotapi.Chattable {
	return tgbotapi.NewMessage(chatId, text)
}

func (b Bot) Run() error {
	updateCh, err := b.GetUpdatesChan(tgbotapi.UpdateConfig{Offset: 0, Timeout: 60})
	if err != nil {
		return errors.Wrap(err, "Run.GetUpdatesChan")
	}
	for update := range updateCh {
		if update.Message == nil || update.Message.IsCommand() {
			continue
		}

		var err error
		switch update.Message.Command() {
		case "start":
			_, err = b.Send(b.cmdStart(update))
		default:
			_, err = b.Send(b.cmdDefault(update))
		}

		if err != nil {
			log.Println(err)
		}

	}

	return nil
}

func (b Bot) cmdStart(update tgbotapi.Update) tgbotapi.Chattable {
	return b.newMsg(update.Message.Chat.ID, "start")
}

func (b Bot) cmdDefault(update tgbotapi.Update) tgbotapi.Chattable {
	return b.newMsg(update.Message.Chat.ID, "default")
}
