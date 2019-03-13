package main

import (
	"github.com/geekfil/zoom-api-service/app"
	"github.com/pkg/errors"
	"log"
)

func main() {
	if err := app.Build().Run(); err != nil {
		log.Panic(errors.Cause(err))
	}
}
