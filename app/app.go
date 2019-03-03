package app

import (
	"github.com/caarlos0/env"
	"github.com/geekfil/zoom-api-service/telegram"
	"github.com/geekfil/zoom-api-service/worker"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"net/http/pprof"
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
	worker   *worker.Worker
}

func (app *App) Run() {
	if err := app.Echo.Start(":3000"); err != nil {
		log.Panicln(err)
	}
}

func (app *App) pprof() {
	r := http.NewServeMux()
	// Регистрация pprof-обработчиков
	r.HandleFunc("debug/pprof/", pprof.Index)
	r.HandleFunc("debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("debug/pprof/profile", pprof.Profile)
	r.HandleFunc("debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("debug/pprof/trace", pprof.Trace)
	app.Echo.GET("/*", echo.WrapHandler(r))
}

func New(tg *telegram.Telegram, config *Config) *App {
	_echo := echo.New()
	_app := &App{
		_echo,
		tg,
		config,
		worker.NewWorker(worker.WithLogger(worker.DefaultLogger)),
	}
	_app.pprof()
	_app.handlers()
	return _app
}
