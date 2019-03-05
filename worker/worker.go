package worker

import (
	"log"
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
	TimeStart      time.Time
	TimeEnd        time.Time
}
type JobList chan *Job

type Config struct {
}

type Worker struct {
	sync.Mutex
	jobs        JobList
	jobsLimiter chan struct{}
	logger      *log.Logger
	config      *Config
}

func NewWorker(opts ...OptionFunc) *Worker {
	worker := &Worker{
		jobs:        make(chan *Job, 1000),
		jobsLimiter: make(chan struct{}, 30),
		logger:      DefaultLogger,
		config:      DefaultConfig,
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
	w.jobs <- job
	return job
}

func (w *Worker) log(format string, v ...interface{}) {
	if w.logger != nil {
		w.logger.Printf(format, v...)
	}
}

func (w *Worker) run() {
	for job := range w.jobs {
		w.jobsLimiter <- struct{}{}
		go w.handleJob(job)
	}
}

func (w *Worker) handleJob(job *Job) {
	if job.CurrentAttempt < job.Attempts && !job.IsRunning {
		job.IsRunning = true
		job.CurrentAttempt++
		w.log("Попытка [%d из %d] выполнения задачи [%s]", job.CurrentAttempt, job.Attempts, job.Name)
		if err := job.handler(); err != nil {
			w.log("Задача [%s] выполнена с ошибкой: %s", job.Name, err)
			job.Errors = append(job.Errors, err.Error())
			w.jobs <- job
		} else {
			w.log("Задача [%s] выполнена успешно", job.Name)

		}
		job.IsRunning = false
	}
	<-w.jobsLimiter
}
