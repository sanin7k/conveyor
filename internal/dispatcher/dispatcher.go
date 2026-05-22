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
	statusUpdateChan chan worker.StatusUpdate
	wg sync.WaitGroup
	ctx context.Context
}

func New(ctx context.Context, s *store.Store, queueSize int, numWorkers int) *Dispatcher {
	q := queue.New(queueSize)
	statusUpdateChan:= make(chan worker.StatusUpdate, numWorkers)
	return &Dispatcher{
		s: s,
		q: q,
		wp: worker.NewWorkerPool(ctx, numWorkers, q, statusUpdateChan),
		statusUpdateChan: statusUpdateChan,
		ctx: ctx,
	}
}

func (d *Dispatcher) Start() {
	d.wp.Start()
	
	d.wg.Add(1)

	go d.readStatusUpdate() 

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

	close(d.statusUpdateChan)

	d.wg.Wait()
}

func (d *Dispatcher) readStatusUpdate() {
	defer d.wg.Done()

	for statusUpdate := range d.statusUpdateChan {
		log.Println("[", statusUpdate.JobID, "]: ", statusUpdate.ResultStatus)

		j, err := d.s.GetJob(d.ctx, statusUpdate.JobID)
		if err != nil {
			log.Println(err)
		}
		
		j.Status = statusUpdate.ResultStatus
		j.ErrorMessage = statusUpdate.ErrorMessage
		if j.Status == job.Failed || j.Status == job.Done {
			j.Attempts += 1
		}

		err = d.s.UpdateJob(d.ctx, j)
		if err != nil {
			log.Println(err)
		}
		
		if j.Status == job.Failed && j.Attempts < maxAttempts {
			go func(j job.Job) {
				j.Status = job.Pending
				err = d.s.UpdateJob(d.ctx, j)
				if err != nil {
					log.Println(err)
				}

				time.Sleep(
					time.Duration(1 << j.Attempts) * time.Second,
				)

				err := d.q.Submit(j)
				if err != nil {
					log.Println(err)
				}
			}(j)
		}
	}
}

