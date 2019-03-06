package telegram

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/geekfil/zoom-api-service/worker"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Config struct {
	ChatId  int64         `env:"TELEGRAM_CHAT_ID"`
	Token   string        `env:"TELEGRAM_TOKEN"`
	Timeout time.Duration `env:"TELEGRAM_CONNECT_TIMEOUT" envDefault:"3s"`
	Proxy   string        `env:"TELEGRAM_PROXY"`
}

func (config Config) httpClient() *http.Client {
	if config.Proxy != "" {
		proxyUrl, err := url.Parse(config.Proxy)
		if err != nil {
			log.Panicln(err)
		}
		dialer, err := proxy.FromURL(proxyUrl, proxy.Direct)
		if err != nil {
			log.Panicln(err)
		}

		return &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
					return dialer.Dial(network, addr)
				},
			},
		}
	} else {
		return &http.Client{Transport: http.DefaultTransport, Timeout: config.Timeout}
	}
}

type SendError struct {
	Date      time.Time `json:"date"`
	Error     string    `json:"error"`
	TypeError string    `json:"type_error"`
}

func NewConfig() *Config {
	config := Config{}
	if err := env.Parse(&config); err != nil {
		log.Panicln(err)
	}
	return &config
}

type Update struct {
	tgbotapi.Update
	command       string
	chatId        int64
	messageId     int
	userId        int
	lastMessageId int
}

type Bot struct {
	*tgbotapi.BotAPI
	sync.Mutex
	Worker            *worker.Worker
	Config            *Config
	stateLastMessages map[int64]int
	update            Update
}

func NewBot(config *Config) (*Bot, error) {
	botApi, err := tgbotapi.NewBotAPIWithClient(config.Token, config.httpClient())
	if err != nil {
		return nil, err
	}
	bot := &Bot{
		Config: config,
		BotAPI:            botApi,
		stateLastMessages: make(map[int64]int, 100),
	}
	return bot, nil
}

func (b Bot) keyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Состояние сервиса", "sysInfo")),
	)
	return &keyboard
}

func (b Bot) Run(update tgbotapi.Update) error {
	newUpdate := Update{Update: update}
	if update.Message != nil {
		newUpdate.command = update.Message.Command()
		newUpdate.chatId = update.Message.Chat.ID
		newUpdate.messageId = update.Message.MessageID
	} else if update.CallbackQuery != nil {
		newUpdate.command = update.CallbackQuery.Data
		newUpdate.chatId = update.CallbackQuery.Message.Chat.ID
		newUpdate.messageId = update.CallbackQuery.Message.MessageID
	} else {
		newUpdate = Update{Update: update}
	}

	b.Lock()
	b.update = newUpdate
	b.stateLastMessages[b.update.chatId] = b.update.messageId
	b.Unlock()

	switch newUpdate.command {
	case "start":
		return b.cmdStart()
	case "sysInfo":
		return b.cmdSysInfo()
	default:
		return b.cmdDefault()
	}
}

func (b Bot) cmdStart() error {
	msg := tgbotapi.NewMessage(b.update.chatId, "Меню сервиса")
	msg.ReplyMarkup = b.keyboard()
	res, err := b.Send(msg)
	if err != nil {
		return errors.Wrap(err, "cmdStart")
	}
	b.setLastMessageId(res.MessageID)
	return nil
}

func (b Bot) cmdDefault() error {
	msg := tgbotapi.NewMessage(b.update.chatId, "Неизвестная команда")
	_, err := b.Send(msg)
	if err != nil {
		return errors.Wrap(err, "cmdDefault")
	}
	return nil
}

func (b Bot) getLastMessageId() int {
	b.Lock()
	defer b.Unlock()
	if lastMessageId, ok := b.stateLastMessages[b.update.chatId]; ok {
		return lastMessageId
	} else {
		return 0
	}
}

func (b Bot) setLastMessageId(id int) {
	b.Lock()
	defer b.Unlock()
	b.stateLastMessages[b.update.chatId] = id
}

func (b Bot) newMessage(text string) tgbotapi.Chattable {
	if id := b.getLastMessageId(); id != 0 {
		msg := tgbotapi.NewEditMessageText(b.update.chatId, id, text)
		msg.ReplyMarkup = b.keyboard()
		msg.ParseMode = tgbotapi.ModeMarkdown
		return msg
	} else {
		msg := tgbotapi.NewMessage(b.update.chatId, text)
		msg.ReplyMarkup = b.keyboard()
		msg.ParseMode = tgbotapi.ModeMarkdown
		return msg
	}
}

func (b Bot) cmdSysInfo() error {
	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("Количество CPU: %d\n", runtime.NumCPU()))
	textBuilder.WriteString(fmt.Sprintf("Горутин в работе: %d\n", runtime.NumGoroutine()))
	res, err := b.Send(b.newMessage(textBuilder.String()))
	if err != nil {
		return errors.Wrap(err, "cmdSysInfo Send")
	}
	b.setLastMessageId(res.MessageID)
	return nil
}
