package api

import (
	"encoding/json"
	"net/http"
	"log"
	"github.com/google/uuid"

	"github.com/sanin7k/conveyor/internal/dispatcher"
	"github.com/sanin7k/conveyor/internal/job"
)

type Handler struct {
	d *dispatcher.Dispatcher
}

func NewHandler(d *dispatcher.Dispatcher) *Handler {
	return &Handler{
		d: d,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /jobs", h.handleSubmit)
	mux.HandleFunc("GET /jobs/{id}", h.handleGetJob)
	mux.HandleFunc("GET /jobs", h.handleListJobs)
}

func (h *Handler) handleSubmit(w http.ResponseWriter, r *http.Request) {
	var jobReq CreateJobRequest

	err := json.NewDecoder(r.Body).Decode(&jobReq)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	job := job.Job{
		ID: uuid.NewString(),
		Type: jobReq.Type,
		Payload: jobReq.Payload,
		Status: "pending",
		Attempts: 0,
	}

	err = h.d.Submit(job)	
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, CreateJobResponse{
			ID: job.ID,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, CreateJobResponse{
		ID: job.ID,
		Message: "job created",
	})

}

func (h *Handler) handleGetJob(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) handleListJobs(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println(err)
	}
}

