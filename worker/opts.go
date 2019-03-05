package worker

import (
	"github.com/pkg/errors"
	"log"
	"os"
)

type OptionFunc func(worker *Worker) error

var DefaultLogger = log.New(os.Stdout, "Worker Jobs: ", log.LstdFlags|log.Lmicroseconds)

func WithLogger(logger *log.Logger) OptionFunc {
	return func(worker *Worker) error {
		if logger != nil {
			worker.logger = logger
			return nil
		} else {
			return errors.New("logger is nil")
		}
	}
}

var DefaultConfig = &Config{}

func WithConfig(config *Config) OptionFunc {
	return func(worker *Worker) error {
		if config != nil {
			worker.config = config
			return nil
		} else {
			return errors.New("config is nil")
		}
	}
}
