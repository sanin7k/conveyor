package job

import "encoding/json"

type JobStatus string

const (
	Pending JobStatus = "pending"
	Running JobStatus = "running"
	Done JobStatus = "done"
	Failed JobStatus = "failed"
)

type Job struct {
	ID string
	Type string
	Payload json.RawMessage
	Status JobStatus
	Attempts int
}

