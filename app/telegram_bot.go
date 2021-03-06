package app

import (
	"encoding/json"
	"fmt"
	"github.com/geekfil/zoom-api-service/telegram"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/url"
)

func (app *App) handlerTelegramBot(g *echo.Group) {
	g.GET("/setwebhook", func(ctx echo.Context) error {
		tg := ctx.Get("tg").(*telegram.Bot)
		webhookUrl, err := url.Parse(fmt.Sprint(ctx.Scheme(), "://", ctx.Request().Host, "/telegram/bot/webhook"))
		if err != nil {
			return echo.NewHTTPError(500, err)
		}

		_, err = tg.SetWebhook(tgbotapi.WebhookConfig{URL: webhookUrl})
		if err != nil {
			return echo.NewHTTPError(400, errors.Wrap(err, "Tg.Bot.SetWebhook"))
		}

		return ctx.JSON(200, echo.Map{
			"message": "OK",
		})
	})
	g.Any("/webhook", func(ctx echo.Context) error {
		var update tgbotapi.Update
		if err := json.NewDecoder(ctx.Request().Body).Decode(&update); err != nil {
			return errors.Wrap(err, "Ошибка декодирования тела запроса")
		}
		if err := app.TelegramBot.Run(update); err != nil {
			return echo.NewHTTPError(500, errors.Wrap(err, "Ошибка выполнения комманды бота"))
		}

		return ctx.NoContent(200)
	})
}
