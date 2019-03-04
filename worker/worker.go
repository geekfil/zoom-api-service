package worker

import (
	"log"
	"os"
	"runtime"
	"sync"
)

type JobHandler func() error
type Job struct {
	Name           string
	handler        JobHandler
	Attempts       uint
	CurrentAttempt uint
	Errors         []string
	Done           chan bool
}
type Worker struct {
	mu     sync.Mutex
	jobs   chan *Job
	_jobs  chan Job
	logger *log.Logger
}

var DefaultLogger = log.New(os.Stdout, "Worker CmdJobs: ", log.LstdFlags|log.Lmicroseconds)

func NewWorker(opts ...OptionFunc) *Worker {
	worker := &Worker{
		jobs: make(chan *Job, 100),
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
	done := make(chan bool)
	job := &Job{
		name,
		handler,
		attempts,
		0,
		make([]string, 0),
		done,
	}
	w.jobs <- job
	return job
}

func (w *Worker) handleJob(job *Job) {
	if job.CurrentAttempt < job.Attempts {
		job.CurrentAttempt++
		w.log("Попытка [%d из %d] выполнения задачи [%s]", job.CurrentAttempt, job.Attempts, job.Name)
		if err := job.handler(); err != nil {
			w.log("Задача [%s] выполнена с ошибкой: %s", job.Name, err)
			job.Errors = append(job.Errors, err.Error())
			w.jobs <- job
			if job.CurrentAttempt == job.Attempts {
				job.Done <- false
			}
		} else {
			job.Done <- true
			w.log("Задача [%s] выполнена успешно", job.Name)
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
		w._jobs <- *job
		go w.handleJob(job)
	}
}

func (w *Worker) Jobs() <-chan Job {
	return w._jobs
}
