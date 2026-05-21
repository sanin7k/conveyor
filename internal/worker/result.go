package worker

import "github.com/sanin7k/conveyor/internal/job"

type Result struct {
	JobID string
	ResultStatus job.JobStatus
}

