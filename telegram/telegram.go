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

type Telegram struct {
	sync.Mutex
	Bot        *tgbotapi.BotAPI
	Config     *Config
	SendErrors []SendError
}

func NewConfig() *Config {
	config := Config{}
	if err := env.Parse(&config); err != nil {
		log.Panicln(err)
	}
	return &config
}

func New(config *Config) *Telegram {

	bot, err := tgbotapi.NewBotAPIWithClient(config.Token, config.httpClient())
	if err != nil {
		log.Panic(err)
	}

	return &Telegram{
		Bot:        bot,
		Config:     config,
		SendErrors: []SendError{},
	}
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
	mu                sync.Mutex
	worker            *worker.Worker
	stateLastMessages map[int64]int
	update            Update
}

func NewBot(botApi *tgbotapi.BotAPI, worker *worker.Worker) *Bot {
	bot := &Bot{
		BotAPI:            botApi,
		worker:            worker,
		stateLastMessages: make(map[int64]int, 100),
	}
	go func() {
		for range time.Tick(time.Hour * 24) {
			bot.mu.Lock()
			bot.stateLastMessages = make(map[int64]int, 100)
			bot.mu.Unlock()
		}
	}()
	return bot
}

func (b Bot) keyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Состояние сервиса", "service_state")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Задачи планировщика", "jobs")),
	)
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


	b.mu.Lock()
	b.update = newUpdate
	b.stateLastMessages[b.update.chatId] = b.update.messageId
	b.mu.Unlock()

	switch newUpdate.command {
	case "start":
		return b.cmdStart()
	case "jobs":
		return b.cmdJobs()
	default:
		return b.cmdDefault()
	}
}

func (b Bot) cmdStart() error {
	msg := tgbotapi.NewMessage(b.update.Message.Chat.ID, "Меню сервиса")
	msg.ReplyMarkup = b.keyboard()
	res, err := b.Send(msg)
	if err != nil {
		return errors.Wrap(err, "cmdStart")
	}
	b.setLastMessageId(res.MessageID)
	return nil
}

func (b Bot) cmdDefault() error {
	msg := tgbotapi.NewMessage(b.update.Message.Chat.ID, "Неизвестная команда")
	_, err := b.Send(msg)
	if err != nil {
		return errors.Wrap(err, "cmdDefault")
	}
	return nil
}

func (b Bot) cmdJobs() error {
	var text strings.Builder
	if len(b.worker.Jobs()) == 0 {
		text.WriteString("Нет запланированных задач")
	} else {
		text.WriteString(fmt.Sprintf("*В очереди выполнения %d задач:* \n", len(b.worker.Jobs())))
	}
	for _, job := range b.worker.Jobs() {
		text.WriteString(fmt.Sprintf("Задача *%s* \n", job.Name))
		if job.IsRunning {
			text.WriteString(fmt.Sprintf("Статус: выполняется. Попытка %d из %d \n", job.CurrentAttempt, job.Attempts))
		} else {
			text.WriteString("Статус: *в очереди* \n")
		}
		text.WriteString(strings.Repeat("-", 15))
		text.WriteString("\n")
	}

	msg := tgbotapi.NewEditMessageText(b.update.chatId, b.getLastMessageId(), text.String())
	msg.ReplyMarkup.InlineKeyboard = append(msg.ReplyMarkup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Состояние сервиса", "service_state")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Задачи планировщика", "jobs")), )
	msg.ParseMode = tgbotapi.ModeMarkdown
	res, err := b.Send(msg)
	if err != nil {
		return errors.Wrap(err, "cmdJobs Send")
	}
	b.setLastMessageId(res.MessageID)

	return nil
}

func (b Bot) getLastMessageId() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	if lastMessageId, ok := b.stateLastMessages[b.update.chatId]; ok {
		return lastMessageId
	} else {
		return 0
	}
}

func (b Bot) setLastMessageId(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stateLastMessages[b.update.chatId] = id
}
