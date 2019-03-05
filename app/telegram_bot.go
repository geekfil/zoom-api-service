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
		tg := ctx.Get("tg").(*telegram.Telegram)
		webhookUrl, err := url.Parse(fmt.Sprint(ctx.Scheme(), "://", ctx.Request().Host, "/telegram/bot/webhook"))
		if err != nil {
			return echo.NewHTTPError(500, err)
		}

		_, err = tg.Bot.SetWebhook(tgbotapi.WebhookConfig{URL: webhookUrl})
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

		if update.Message == nil {
			return errors.New("Сообщение пустое")
		}
		if update.Message.IsCommand() {
			var err error
			switch update.Message.Command() {
			case "start":
				_, err = app.Telegram.Bot.Send(app.Telegram.CmdStart(update))
			case "help":
				_, err = app.Telegram.Bot.Send(app.Telegram.CmdHelp(update))
			case "jobs":
				_, err = app.Telegram.Bot.Send(app.Telegram.CmdJobs(update, app.Worker))

			default:
				_, err = app.Telegram.Bot.Send(app.Telegram.Default(update))
			}

			if err != nil {
				return echo.NewHTTPError(500, errors.Wrap(err, "Ошибка выполнения команды telegram"))
			}
		} else {

		}

		return ctx.NoContent(200)
	})
}
