package dispatcher

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sanin7k/conveyor/internal/job"
	"github.com/sanin7k/conveyor/internal/queue"
	"github.com/sanin7k/conveyor/internal/store"
	"github.com/sanin7k/conveyor/internal/worker"
)

const maxAttempts = 3

type Dispatcher struct {
	s *store.Store
	q *queue.Queue
	wp *worker.WorkerPool
	results chan worker.Result
	wg sync.WaitGroup
	ctx context.Context
}

func New(ctx context.Context, s *store.Store, queueSize int, numWorkers int) *Dispatcher {
	q := queue.New(queueSize)
	results := make(chan worker.Result, numWorkers)
	return &Dispatcher{
		s: s,
		q: q,
		wp: worker.NewWorkerPool(ctx, numWorkers, q, results),
		results: results,
		ctx: ctx,
	}
}

func (d *Dispatcher) Start() {
	d.wp.Start()
	
	d.wg.Add(1)

	go d.readResults() 

	pendingJobs, err := d.s.GetPendingJobs(d.ctx)
	if err != nil {
		log.Println(err)
	}

	for _, job := range pendingJobs {
		err := d.q.Submit(job)
		if err != nil {
			log.Println(err)
		}
	}
}

func (d *Dispatcher) Submit(job job.Job) error {
	err := d.s.CreateJob(d.ctx, job)
	if err != nil {
		log.Println(err)
		return err
	}

	err = d.q.Submit(job)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (d *Dispatcher) Shutdown() {
	d.wp.Stop()

	close(d.results)

	d.wg.Wait()
}

func (d *Dispatcher) readResults() {
	defer d.wg.Done()

	for res := range d.results {
		log.Println("[", res.JobID, "]: ", res.ResultStatus)

		j, err := d.s.GetJob(d.ctx, res.JobID)
		if err != nil {
			log.Println(err)
		}
		
		j.Status = res.ResultStatus
		j.ErrorMessage = res.ErrorMessage
		j.Attempts += 1

		err = d.s.UpdateJob(d.ctx, j)
		if err != nil {
			log.Println(err)
		}
		
		if j.Status == job.Failed && j.Attempts < maxAttempts {
			go func(j job.Job) {
				time.Sleep(
					time.Duration(1 << j.Attempts) * time.Second,
				)
				j.Status = job.Pending
				err = d.s.UpdateJob(d.ctx, j)
				if err != nil {
					log.Println(err)
				}

				err := d.q.Submit(j)
				if err != nil {
					log.Println(err)
				}
			}(j)
		}
	}
}

