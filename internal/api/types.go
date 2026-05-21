package api

import "encoding/json"

type CreateJobRequest struct {
	Type 		string 					`json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type CreateJobResponse struct {
	ID 			string `json:"id,omitempty"`
	Message string `json:"message"`
}

