package worker

import (
	"context"
	"sync"
	"log"

	"github.com/sanin7k/conveyor/internal/queue"
	"github.com/sanin7k/conveyor/internal/job"
)

type WorkerPool struct {
	numWorkers int
	jobQueue *queue.Queue
	wg sync.WaitGroup	
	results chan Result
	ctx context.Context
}

func NewWorkerPool(ctx context.Context, numWorkers int, jobQueue *queue.Queue, results chan Result) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobQueue: jobQueue,
		results: results,
		ctx: ctx,
	}
}

func (wp *WorkerPool) Start() {
	for range wp.numWorkers {
		wp.wg.Add(1)
		go wp.runWorker()
	}
}

func (wp *WorkerPool) Stop() {
	wp.wg.Wait()
}

func (wp *WorkerPool) runWorker() {
	defer wp.wg.Done()

	log.Println("worker starting")

	for {
		select {
		case <-wp.ctx.Done():
			log.Println("worker shutting down")
			return
		case j, ok := <-wp.jobQueue.Jobs():
			if !ok {
				return
			}
			err := process(j)
			var status job.JobStatus
			if err == nil {
				status = "done"
			} else {
				status = "failed"
			}
			result := Result{j.ID, status}
			wp.results <- result
		}
	}
}

