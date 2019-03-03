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
	jobs   chan *Job
	logger *log.Logger
}

var DefaultLogger = log.New(os.Stdout, "Worker Jobs: ", log.LstdFlags)

func NewWorker(opts ...OptionFunc) *Worker {
	worker := &Worker{
		jobs: make(chan *Job, 0),
	}
	for _, opt := range opts {
		if err := opt(worker); err != nil {
			log.Fatalln(err)
		}
	}
	go worker.run()
	return worker
}

func (w *Worker) AddJob(name string, handler JobHandler, attempts uint) {
	w.log("Добавлена задача [%s]", name)
	w.jobs <- &Job{
		name,
		handler,
		attempts,
		0,
		make([]string, 0),
	}
}

func (w *Worker) handleJob(job *Job) {
	if job.currentAttempt < job.attempts {
		job.currentAttempt++
		w.log("Попытка [%d из %d] выполнения задачи [%s]", job.currentAttempt, job.attempts, job.name)
		if err := job.handler(); err != nil {
			w.log("Задача [%s] выполнена с ошибкой: %s", job.name, err)
			job.errors = append(job.errors, err.Error())
			w.jobs <- job
		} else {
			w.log("Задача [%s] выполнена успешно", job.name)
		}
	}
	runtime.Gosched()
}

func (w *Worker) log(format string, v ...interface{}) {
	if w.logger != nil {
		w.logger.Printf(format, v...)
	}
}

func (w *Worker) run() {
	for job := range w.jobs {
		go w.handleJob(job)
	}
}
