package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"errors"

	"github.com/google/uuid"

	"github.com/sanin7k/conveyor/internal/dispatcher"
	"github.com/sanin7k/conveyor/internal/job"
	"github.com/sanin7k/conveyor/internal/store"
)

type Handler struct {
	d *dispatcher.Dispatcher
	s *store.Store
}

func NewHandler(d *dispatcher.Dispatcher, s *store.Store) *Handler {
	return &Handler{
		d: d,
		s: s,
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
	id := r.PathValue("id")

	job, err := h.s.GetJob(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func (h *Handler) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.s.ListJobs(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println(err)
	}
}

