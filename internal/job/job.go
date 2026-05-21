package job

import "encoding/json"

type JobStatus string

const (
	Pending JobStatus = "pending"
	Running JobStatus = "running"
	Done 		JobStatus = "done"
	Failed 	JobStatus = "failed"
)

type Job struct {
	ID string								`json:"id"`
	Type string 						`json:"type"`
	Payload json.RawMessage `json:"payload"`
	Status JobStatus 				`json:"status"`
	Attempts int 						`json:"attempts"`
	ErrorMessage *string 		`json:"error_message,omitempty"`
}

