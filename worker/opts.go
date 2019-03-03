package worker

import (
	"github.com/pkg/errors"
	"log"
)

type OptionFunc func(worker *Worker) error

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
