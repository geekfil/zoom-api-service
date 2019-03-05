package telegram

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/geekfil/zoom-api-service/worker"
	"github.com/go-telegram-bot-api/telegram-bot-api"
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

func (t Telegram) CmdStart(update tgbotapi.Update) tgbotapi.Chattable {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Состояние сервиса")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Задачи планировщика")),
	)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,"")
	msg.ReplyMarkup = keyboard
	return msg
}

func (t Telegram) CmdHelp(update tgbotapi.Update) tgbotapi.Chattable {
	var text = `
/help - справка по командам
/jobs - текущие задачи планировщика
/goroutines - количество работающих горутин
/cpu - количество ядер процессора
`
	return tgbotapi.NewMessage(update.Message.Chat.ID, text)
}

func (t Telegram) CmdJobs(update tgbotapi.Update, worker *worker.Worker) tgbotapi.Chattable {
	var text strings.Builder
	if len(worker.Jobs()) == 0 {
		text.WriteString("Нет запланированных задач")
	} else {
		text.WriteString(fmt.Sprintf("*В очереди выполнения %d задач:* \n", len(worker.Jobs())))
	}
	for _, job := range worker.Jobs() {
		text.WriteString(fmt.Sprintf("Задача *%s* \n", job.Name))
		if job.IsRunning {
			text.WriteString(fmt.Sprintf("Статус: выполняется. Попытка %d из %d \n", job.CurrentAttempt, job.Attempts))
		} else {
			text.WriteString("Статус: *в очереди* \n")
		}
		text.WriteString(strings.Repeat("-", 15))
		text.WriteString("\n")
	}

	return t.botNewMarkdownMessage(update.Message.Chat.ID, text.String())
}

func (t Telegram) Default(update tgbotapi.Update) tgbotapi.Chattable {
	return tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Отправьте /help для получения справки")
}

func (t Telegram) botNewMarkdownMessage(chatId int64, text string) tgbotapi.Chattable {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	return msg
}
