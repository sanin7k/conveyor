package worker

import "github.com/sanin7k/conveyor/internal/job"

type StatusUpdate struct {
	JobID string
	ResultStatus job.JobStatus
	ErrorMessage *string
}

