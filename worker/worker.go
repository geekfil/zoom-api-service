package worker

import (
	"log"
	"os"
	"runtime"
	"sync"
)

type JobHandler func() error
type Job struct {
	name           string
	handler        JobHandler
	attempts       uint
	currentAttempt uint
	errors         []string
}
type Worker struct {
	mu     sync.Mutex
	jobs   []*Job
	logger *log.Logger
}

var DefaultLogger = log.New(os.Stdout, "Worker Jobs: ", log.LstdFlags)

func NewWorker(opts ...OptionFunc) *Worker {
	worker := &Worker{
		jobs: make([]*Job, 0),
	}
	for _, opt := range opts {
		if err := opt(worker); err != nil {
			log.Fatalln(err)
		}
	}
	return worker
}

func (w *Worker) AddJob(name string, handler JobHandler, attempts uint) {
	w.log("Добавлена задача [%s]", name)
	w.mu.Lock()
	defer w.mu.Unlock()
	w.jobs = append(w.jobs, &Job{
		name,
		handler,
		attempts,
		0,
		make([]string, 0),
	})
}

func (w *Worker) log(format string, v ...interface{}) {
	if w.logger != nil {
		w.logger.Printf(format, v...)
	}
}

func (w *Worker) Run() {

	for {
		if len(w.jobs) > 0 {
			w.mu.Lock()
			lastIndex := len(w.jobs) - 1
			currentJob := w.jobs[lastIndex]

			if currentJob.currentAttempt < currentJob.attempts {
				currentJob.currentAttempt++
				w.log("Попытка [%d из %d] выполнения задачи [%s]", currentJob.currentAttempt, currentJob.attempts, currentJob.name)
				if err := currentJob.handler(); err != nil {
					w.log("Задача [%s] выполнена с ошибкой: %s", currentJob.name, err)
					currentJob.errors = append(currentJob.errors, err.Error())
				} else {
					w.log("Задача [%s] выполнена успешно", currentJob.name)
					w.jobs = w.jobs[:lastIndex]
				}
			}
			runtime.Gosched()
			w.mu.Unlock()
		}

	}
}
