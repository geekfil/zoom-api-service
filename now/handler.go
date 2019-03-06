package main

import (
	"github.com/geekfil/zoom-api-service/app"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	app.Instance.Echo.ServeHTTP(w, r)
}
