package di

import (
	"github.com/geekfil/zoom-api-service/app"
	"github.com/geekfil/zoom-api-service/telegram"
	"github.com/geekfil/zoom-api-service/worker"
	"go.uber.org/dig"
	"log"
	"sync"
)

var (
	Container *dig.Container
	once      sync.Once
)

func init() {
	once.Do(func() {
		Container = dig.New()
		var err error
		err = Container.Provide(telegram.NewConfig)
		err = Container.Provide(telegram.New)
		err = Container.Provide(worker.NewWorker)
		err = Container.Provide(app.NewConfig)
		err = Container.Provide(app.New)
		if err != nil {
			log.Panicln(err)
		}
	})
}
