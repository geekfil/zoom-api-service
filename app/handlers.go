package app

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

func (app App) handlers() {
	web := app.Echo
	web.Use(middleware.Logger())
	web.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("tg", app.TelegramBot)
			ctx.Set("worker", app.Worker)
			return next(ctx)
		}
	})
	web.GET("/", func(context echo.Context) error {
		return context.String(200, "ZOOM PRIVATE API")
	})

	app.handlerTelegramBot(web.Group("/telegram/bot"))
	app.handlerWebApi(web.Group("/api"))

}

func (app *App) handlerWebApi(group *echo.Group) {
	group.Group("/api")
	group.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if app.Config.Token == ctx.QueryParam("token") {
				return next(ctx)
			}
			return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}
	})

	tgGroup := group.Group("/telegram")

	tgGroup.GET("/send", func(ctx echo.Context) error {

		var text string
		if text = ctx.QueryParam("text"); len(text) == 0 {
			return echo.NewHTTPError(400, "text is required")
		}

		app.Worker.AddJob("Отправка уведомления в Telegram", func() error {
			if _, err := app.TelegramBot.Send(tgbotapi.NewMessage(app.TelegramBot.Config.ChatId, text)); err != nil {
				return err
			}

			return nil
		}, 5)

		return ctx.JSON(200, map[string]string{
			"message": "OK",
		})

	})

}
