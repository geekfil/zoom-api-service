package app

import (
	"github.com/geekfil/zoom-api-service/telegram"
	"github.com/geekfil/zoom-api-service/worker"
	"github.com/labstack/gommon/log"
	"sync"
)

var (
	Instance *App
	once     sync.Once
)

func init() {
	once.Do(func() {
		tgConf := telegram.NewConfig()
		_worker := worker.NewWorker()
		tg, err := telegram.NewBot(tgConf)
		if err != nil {
			log.Panic(err)
		}
		appConf := NewConfig()
		Instance = New(tg, appConf, _worker)
	})
}
