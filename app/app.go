package app

import (
	"github.com/caarlos0/env"
	"github.com/geekfil/zoom-api-service/telegram"
	"github.com/geekfil/zoom-api-service/worker"
	"github.com/labstack/echo"
	"log"
)

type Config struct {
	Token string `env:"APP_TOKEN"`
	Dev   bool   `env:"APP_DEV"`
}

func NewConfig() *Config {
	config := &Config{Token: ""}
	if err := env.Parse(config); err != nil {
		log.Panicln(err)
	}
	return config
}

type App struct {
	Echo        *echo.Echo
	Config      *Config
	Worker      *worker.Worker
	TelegramBot *telegram.Bot
}

func (app *App) Run() error {
	if err := app.Echo.Start(":3000"); err != nil {
		return err
	}
	return nil
}

func New(tgBot *telegram.Bot, config *Config, worker *worker.Worker) *App {
	_echo := echo.New()
	_app := &App{
		Echo:        _echo,
		Config:      config,
		Worker:      worker,
		TelegramBot: tgBot,
	}
	_app.handlers()
	return _app
}

func Build() *App {
	tgConf := telegram.NewConfig()
	_worker := worker.NewWorker()
	tg, err := telegram.NewBot(tgConf)
	if err != nil {
		log.Panic(err)
	}
	appConf := NewConfig()
	return New(tg, appConf, _worker)
}
