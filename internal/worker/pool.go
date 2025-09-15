package worker

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Job func()

type Pool struct {
	workers  int
	jobQueue chan Job
	quit     chan bool
	wg       sync.WaitGroup
	logger   *logrus.Logger
}

func NewPool(workers int, logger *logrus.Logger) *Pool {
	return &Pool{
		workers:  workers,
		jobQueue: make(chan Job, 100), // Buffer for 100 jobs
		quit:     make(chan bool),
		logger:   logger,
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	p.logger.Infof("Worker %d started", id)

	for {
		select {
		case job := <-p.jobQueue:
			p.logger.Debugf("Worker %d processing job", id)
			job()
		case <-p.quit:
			p.logger.Infof("Worker %d stopping", id)
			return
		}
	}
}

func (p *Pool) AddJob(job Job) {
	select {
	case p.jobQueue <- job:
		p.logger.Debug("Job added to queue")
	default:
		p.logger.Warn("Job queue is full, dropping job")
	}
}

func (p *Pool) Stop() {
	p.logger.Info("Stopping worker pool...")
	close(p.quit)
	p.wg.Wait()
	p.logger.Info("Worker pool stopped")
}
