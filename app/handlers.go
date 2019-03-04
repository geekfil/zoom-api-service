package app

import (
	"github.com/geekfil/zoom-api-service/telegram"
	"github.com/geekfil/zoom-api-service/worker"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
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
			ctx.Set("worker", app.worker)
			return next(ctx)
		}
	})

	webTelegramBot(web.Group("/telegram/bot"))

	web.GET("/", func(context echo.Context) error {
		return context.String(200, "ZOOM PRIVATE API")
	})

	webSys(web.Group("/sys"))

	apiGroup := web.Group("/api")
	apiGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if app.Config.Token == ctx.QueryParam("token") {
				return next(ctx)
			}
			return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}
	})
	webApi(apiGroup)

}

func webApi(group *echo.Group) {
	telegramGroup := group.Group("/telegram")

	telegramGroup.GET("/send", func(ctx echo.Context) error {

		var tg = ctx.Get("tg").(*telegram.Telegram)
		var workerJob = ctx.Get("worker").(*worker.Worker)
		var text string
		if text = ctx.QueryParam("text"); len(text) == 0 {
			return echo.NewHTTPError(400, "text is required")
		}

		workerJob.AddJob("Отправка уведомления в Telegram", func() error {
			if _, err := tg.Bot.Send(tgbotapi.NewMessage(tg.Config.ChatId, text)); err != nil {
				tg.Lock()
				tg.SendErrors = append(tg.SendErrors, telegram.SendError{
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
				tg.Unlock()
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

func webSys(g *echo.Group) {
	g.GET("/info", func(context echo.Context) error {
		return context.JSON(200, echo.Map{
			"NumGoroutine": runtime.NumGoroutine(),
			"NumCPU":       runtime.NumCPU(),
		})
	})
}

func webTelegramBot(g *echo.Group) {
	g.GET("/setwebhook", func(ctx echo.Context) error {
		tg := ctx.Get("tg").(*telegram.Telegram)
		url := ctx.Request().URL.String()
		if _, err := tg.Bot.SetWebhook(tgbotapi.NewWebhook("/telegram/bot/webhook")); err != nil {
			return echo.NewHTTPError(400, errors.Wrap(err, "Tg.Bot.SetWebhook"))
		}
		return ctx.JSON(200, echo.Map{
			"message": "OK",
		})
	})
	g.GET("/test", func(ctx echo.Context) error {
		return ctx.JSON(200,ctx.Request())
	})
}
