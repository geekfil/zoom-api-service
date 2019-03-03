package app

import (
	"github.com/geekfil/zoom-api-service/telegram"
	"github.com/labstack/echo"
	"log"
)

type App struct {
	Echo     *echo.Echo
	Telegram *telegram.Telegram
}

func (app *App) Run() {
	app.Echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("tg", app.Telegram)
			return next(ctx)
		}
	})
	app.handlers()
	if err := app.Echo.Start(":3000"); err != nil {
		log.Panicln(err)
	}
}


func New(tg *telegram.Telegram) *App {
	_echo := echo.New()
	_app := &App{
		_echo,
		tg,
	}
	return _app
}
