package dispatcher

import (
	"context"
	"log"
	"sync"

	"github.com/sanin7k/conveyor/internal/job"
	"github.com/sanin7k/conveyor/internal/queue"
	"github.com/sanin7k/conveyor/internal/worker"
)

type Dispatcher struct {
	q *queue.Queue
	wp *worker.WorkerPool
	results chan worker.Result
	wg sync.WaitGroup
}

func New(ctx context.Context, queueSize int, numWorkers int) *Dispatcher {
	q := queue.New(queueSize)
	results := make(chan worker.Result, numWorkers)
	return &Dispatcher{
		q: q,
		wp: worker.NewWorkerPool(ctx, numWorkers, q, results),
		results: results,
	}
}

func (d *Dispatcher) Start() {
	d.wp.Start()
	
	d.wg.Add(1)

	go d.readResults() 
}

func (d *Dispatcher) Submit(job job.Job) error {
	err := d.q.Submit(job)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (d *Dispatcher) Shutdown() {
	d.wp.Stop()

	close(d.results)

	d.wg.Wait()
}

func (d *Dispatcher) readResults() {
	defer d.wg.Done()

	for res := range d.results {
		log.Println("[ ", res.JobID, " ]: ", res.ResultStatus)
	}
}

