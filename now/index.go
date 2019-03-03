package main

import (
	"go.uber.org/dig"
	"log"
	"net/http"
	"zoom-api/app"
	"zoom-api/telegram"
)

var container = dig.New()

func init() {
	var err error
	err = container.Provide(telegram.NewConfig)
	err = container.Provide(telegram.New)
	err = container.Provide(app.New)
	if err != nil {
		log.Panicln(err)
	}
}
func Handler(w http.ResponseWriter, r *http.Request) {
	err := container.Invoke(func(app *app.App) {
		app.Echo.ServeHTTP(w, r)
	})
	if err != nil {
		log.Panicln(err)
	}
}
