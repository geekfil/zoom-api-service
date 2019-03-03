package main

import (
	"github.com/geekfil/zoom-api-service/app"
	"github.com/geekfil/zoom-api-service/di"
	"log"
)

func main() {
	err := di.Container.Invoke(func(app *app.App) {
		app.Run()
	})
	if err != nil {
		log.Panicln(err)
	}
}
