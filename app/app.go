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
	Echo     *echo.Echo
	Telegram *telegram.Telegram
	Config   *Config
	Worker   *worker.Worker
}

func (app *App) Run() {
	if err := app.Echo.Start(":3000"); err != nil {
		log.Panicln(err)
	}
}

func New(tg *telegram.Telegram, config *Config) *App {
	_echo := echo.New()
	_app := &App{
		_echo,
		tg,
		config,
		worker.NewWorker(worker.WithLogger(worker.DefaultLogger)),
	}
	_app.handlers()
	return _app
}
