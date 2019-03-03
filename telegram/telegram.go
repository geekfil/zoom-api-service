package telegram

import (
	"github.com/caarlos0/env"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"time"
)

type Config struct {
	ChatId  int64         `env:"CHAT_ID"`
	Token   string        `env:"TELEGRAM_TOKEN"`
	Timeout time.Duration `env:"TELEGRAM_CONNECT_TIMEOUT" envDefault:"3s"`
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
	bot, err := tgbotapi.NewBotAPIWithClient(config.Token, &http.Client{
		Timeout: config.Timeout,
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

