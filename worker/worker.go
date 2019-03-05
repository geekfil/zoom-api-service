package worker

import (
	"log"
	"os"
	"sync"
	"time"
)

type JobHandler func() error
type Job struct {
	sync.Mutex
	Name           string
	handler        JobHandler
	Attempts       uint
	CurrentAttempt uint
	Errors         []string
	IsRunning      bool
}
type Worker struct {
	mu     sync.Mutex
	jobs   []*Job
	logger *log.Logger
}

var DefaultLogger = log.New(os.Stdout, "Worker Jobs: ", log.LstdFlags|log.Lmicroseconds)

func NewWorker(opts ...OptionFunc) *Worker {
	worker := &Worker{
		jobs: make([]*Job, 0),
	}
	for _, opt := range opts {
		if err := opt(worker); err != nil {
			log.Fatalln(err)
		}
	}
	go worker.run()
	return worker
}

func (w *Worker) AddJob(name string, handler JobHandler, attempts uint) *Job {
	w.log("Добавлена задача [%s]", name)
	job := &Job{
		Name:           name,
		handler:        handler,
		Attempts:       attempts,
		CurrentAttempt: 0,
		Errors:         make([]string, 0),
		IsRunning:      false,
	}

	w.mu.Lock()
	w.jobs = append(w.jobs, job)
	w.mu.Unlock()
	return job
}

func (w *Worker) log(format string, v ...interface{}) {
	if w.logger != nil {
		w.logger.Printf(format, v...)
	}
}

func (w *Worker) Jobs() []*Job {
	return w.jobs
}

func (w *Worker) run() {
	for range time.Tick(time.Second) {
		for index, job := range w.jobs {
			if job.CurrentAttempt < job.Attempts && !job.IsRunning {
				job.IsRunning = true
				job.CurrentAttempt++
				w.log("Попытка [%d из %d] выполнения задачи [%s]", job.CurrentAttempt, job.Attempts, job.Name)
				go func() {
					job.Lock()
					defer func() {
						job.IsRunning = false
						job.Unlock()
					}()
					if err := job.handler(); err != nil {
						w.log("Задача [%s] выполнена с ошибкой: %s", job.Name, err)
						job.Errors = append(job.Errors, err.Error())
					} else {
						w.jobs = append(w.jobs[:index], w.jobs[index+1:]...)
						w.log("Задача [%s] выполнена успешно", job.Name)
					}


				}()
			}
		}
	}
}
