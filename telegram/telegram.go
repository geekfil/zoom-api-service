package telegram

import (
	"context"
	"github.com/caarlos0/env"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	ChatId  int64         `env:"TELEGRAM_CHAT_ID"`
	Token   string        `env:"TELEGRAM_TOKEN"`
	Timeout time.Duration `env:"TELEGRAM_CONNECT_TIMEOUT" envDefault:"3s"`
	Proxy   string        `env:"TELEGRAM_PROXY"`
}

type SendError struct {
	Date  time.Time `json:"date"`
	Error error     `json:"error"`
}

type Telegram struct {
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
	proxyUrl, err := url.Parse(config.Proxy)
	if err != nil {
		log.Panicln(err)
	}
	dialer, err := proxy.FromURL(proxyUrl, proxy.Direct)
	if err != nil {
		log.Panicln(err)
	}
	bot, err := tgbotapi.NewBotAPIWithClient(config.Token, &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
				return dialer.Dial(network, addr)
			},
		},
	})
	if err != nil {
		log.Panic(err)
	}

	return &Telegram{
		Bot:        bot,
		Config:     config,
		SendErrors: []SendError{},
	}
}
