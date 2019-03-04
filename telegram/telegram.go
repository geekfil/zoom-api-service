package telegram

import (
	"context"
	"github.com/caarlos0/env"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"net/http"
	"net/url"
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

type Cmd struct {
}

type Telegram struct {
	sync.Mutex
	Cmd        Cmd
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

func (t *Telegram) RunBot() error {
	updateCh, err := t.Bot.GetUpdatesChan(tgbotapi.UpdateConfig{Offset: 0, Timeout: 60})
	if err != nil {
		return errors.Wrap(err, "Run.GetUpdatesChan")
	}
	for update := range updateCh {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		var err error
		switch update.Message.Command() {
		case "start":
			_, err = t.Bot.Send(t.Cmd.Start(update))
		default:
			_, err = t.Bot.Send(t.Cmd.Default(update))
		}


		if err != nil {
			log.Println(err)
		}

	}

	return nil
}

func (b Cmd) Start(update tgbotapi.Update) tgbotapi.Chattable {
	return tgbotapi.NewMessage(update.Message.Chat.ID, "Start")
}

func (b Cmd) Default(update tgbotapi.Update) tgbotapi.Chattable {
	return tgbotapi.NewMessage(update.Message.Chat.ID, "Default")
}
