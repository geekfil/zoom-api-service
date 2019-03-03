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
	Error     error     `json:"error"`
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
