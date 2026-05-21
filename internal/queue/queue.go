package queue

import (
	"github.com/sanin7k/conveyor/internal/job"
	"errors"
)

var ErrQueueFull = errors.New("queue full")

type Queue struct {
	jobs chan job.Job
}

func New(size int) *Queue {
	return &Queue{
		jobs: make(chan job.Job, size),
	}
}

func (q *Queue) Submit(job job.Job) error {
	select {
	case q.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Jobs() <-chan job.Job {
	return q.jobs
}

