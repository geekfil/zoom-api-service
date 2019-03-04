package main

import (
	"github.com/geekfil/zoom-api-service/app"
	"github.com/geekfil/zoom-api-service/di"
	"github.com/labstack/echo"
	"log"
	"net/http"
)

var server *echo.Echo

func init() {
	err := di.Container.Invoke(func(app *app.App) {
		server = app.Echo
	})
	if err != nil {
		log.Panicln(err)
	}
}
func Handler(w http.ResponseWriter, r *http.Request) {
	server.ServeHTTP(w, r)
}
