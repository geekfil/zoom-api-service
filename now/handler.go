package main

import (
	"github.com/geekfil/zoom-api-service/app"
	"github.com/geekfil/zoom-api-service/di"
	"log"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	err := di.Container.Invoke(func(app *app.App) {
		app.Echo.ServeHTTP(w, r)
	})
	if err != nil {
		log.Panicln(err)
	}
}
