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
	statusUpdateChan chan StatusUpdate
	ctx context.Context
}

func NewWorkerPool(ctx context.Context, numWorkers int, jobQueue *queue.Queue, statusUpdateChan chan StatusUpdate) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobQueue: jobQueue,
		statusUpdateChan: statusUpdateChan,
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

			wp.statusUpdateChan <- StatusUpdate{j.ID, job.Running, nil}

			err := process(j)

			var statusUpdate StatusUpdate

			if err == nil {
				statusUpdate = StatusUpdate{j.ID, job.Done, nil}
			} else {
				errmsg := err.Error()
				statusUpdate = StatusUpdate{j.ID, job.Failed, &errmsg}

			}
			wp.statusUpdateChan <- statusUpdate
		}
	}
}

