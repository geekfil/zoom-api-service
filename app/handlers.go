package app

import (
	"github.com/geekfil/zoom-api-service/telegram"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/labstack/echo"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"
)

func (app App) handlers() {
	web := app.Echo
	web.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("tg", app.Telegram)
			ctx.Set("worker", app.Worker)
			return next(ctx)
		}
	})

	app.handlerTelegramBot(web.Group("/telegram/bot"))

	web.GET("/", func(context echo.Context) error {
		return context.String(200, "ZOOM PRIVATE API")
	})

	app.handlerWebSys(web.Group("/sys"))

	apiGroup := web.Group("/api")
	apiGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if app.Config.Token == ctx.QueryParam("token") {
				return next(ctx)
			}
			return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}
	})
	app.handlerWebApi(apiGroup)

}

func (app *App) handlerWebApi(group *echo.Group) {
	telegramGroup := group.Group("/telegram")

	telegramGroup.GET("/send", func(ctx echo.Context) error {

		var text string
		if text = ctx.QueryParam("text"); len(text) == 0 {
			return echo.NewHTTPError(400, "text is required")
		}

		app.Worker.AddJob("Отправка уведомления в Telegram", func() error {
			if _, err := app.Telegram.Bot.Send(tgbotapi.NewMessage(app.Telegram.Config.ChatId, text)); err != nil {
				app.Telegram.Lock()
				app.Telegram.SendErrors = append(app.Telegram.SendErrors, telegram.SendError{
					Date: time.Now(),
					Error: func(err error) string {
						switch e := err.(type) {
						case *url.Error:
							return e.Error()
						default:
							return e.Error()
						}
					}(err),
					TypeError: reflect.TypeOf(err).String(),
				})
				app.Telegram.Unlock()
				return err
			}

			return nil
		}, 5)

		return ctx.JSON(200, map[string]string{
			"message": "OK",
		})

	})
	telegramGroup.GET("/send/errors", func(ctx echo.Context) error {
		var tg = ctx.Get("tg").(*telegram.Telegram)
		return ctx.JSON(200, tg.SendErrors)
	})
	telegramGroup.GET("/send/errors/clear", func(ctx echo.Context) error {
		var tg = ctx.Get("tg").(*telegram.Telegram)
		tg.SendErrors = []telegram.SendError{}
		return ctx.JSON(200, map[string]string{
			"message": "OK",
		})
	})
}

func (app *App) handlerWebSys(g *echo.Group) {
	g.GET("/info", func(context echo.Context) error {
		return context.JSON(200, echo.Map{
			"NumGoroutine": runtime.NumGoroutine(),
			"NumCPU":       runtime.NumCPU(),
		})
	})
}
