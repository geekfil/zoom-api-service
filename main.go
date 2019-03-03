package main

import (
	"github.com/geekfil/zoom-api-service/app"
	"github.com/geekfil/zoom-api-service/telegram"
	"go.uber.org/dig"
	"log"
)

func main() {
	var container = dig.New()
	var err error
	err = container.Provide(telegram.NewConfig)
	err = container.Provide(telegram.New)
	err = container.Provide(app.New)
	err = container.Invoke(func(app *app.App) {
		app.Run()
	})
	if err != nil {
		log.Panicln(err)
	}
}
