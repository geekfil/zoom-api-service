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

func (app *App) Run() {
	if err := app.Echo.Start(":3000"); err != nil {
		log.Panicln(err)
	}
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
