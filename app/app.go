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
	_echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("tg", tg)
			return next(ctx)
		}
	})
	_app.httpHandler(_echo.Group("/api"))
	return _app
}

func (app App) httpHandler(http *echo.Group) {
	var httpTelegram = http.Group("/telegram")
	httpTelegram.GET("/send", func(ctx echo.Context) error {
		var tg = ctx.Get("tg").(*telegram.Telegram)
		var text string
		if text = ctx.QueryParam("text"); len(text) == 0 {
			return echo.NewHTTPError(400, "text is required")
		}
		go tg.Send(text)
		return ctx.JSON(200, map[string]string{
			"message": "Notification sent",
		})
	})
	httpTelegram.GET("/send/errors", func(ctx echo.Context) error {
		var tg = ctx.Get("tg").(*telegram.Telegram)
		return ctx.JSON(200, tg.SendErrors)
	})
	httpTelegram.GET("/send/errors/clear", func(ctx echo.Context) error {
		var tg = ctx.Get("tg").(*telegram.Telegram)
		tg.SendErrors = []telegram.SendError{}
		return ctx.JSON(200, map[string]string{
			"message": "OK",
		})
	})
}
