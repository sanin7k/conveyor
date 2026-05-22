package store

import (
	"database/sql"
	"context"

	"github.com/sanin7k/conveyor/internal/job"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
} 

func (s *Store) CreateJob(ctx context.Context, j job.Job) error {
	query := `
		INSERT INTO jobs(id, job_type, payload)
		VALUES($1, $2, $3)
	`

	_, err := s.db.ExecContext(ctx, query, j.ID, j.Type, j.Payload)
	return err
}

func (s *Store) UpdateJob(ctx context.Context, j job.Job) error {
	query := `
		UPDATE jobs
		SET status = $1, attempts = $2, error_message = $3, updated_at = NOW() 
	`

	switch j.Status {
	case job.Running:
		query += ", started_at = NOW() "
	case job.Done, job.Failed:
		query += ", completed_at = NOW() "
	}

	query += "WHERE id = $4"

	result, err := s.db.ExecContext(ctx, query, j.Status, j.Attempts, j.ErrorMessage, j.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) GetJob(ctx context.Context, id string) (job.Job, error) {
	query := `
		SELECT id, job_type, payload, status, attempts, error_message
		FROM jobs
		WHERE id = $1
	`

	var j job.Job
	var errMsg sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &errMsg)
	if errMsg.Valid {
		j.ErrorMessage = &errMsg.String
	} else {
		j.ErrorMessage = nil
	}

	return j, err
}

func (s *Store) ListJobs(ctx context.Context) ([]job.Job, error) {
	query := `
		SELECT id, job_type, payload, status, attempts, error_message
		FROM jobs
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []job.Job

	for rows.Next() {
		var j job.Job
		var errMsg sql.NullString

		err := rows.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &errMsg)
		if err != nil {
			return nil, err
		}
		if errMsg.Valid {
			j.ErrorMessage = &errMsg.String
		} else {
			j.ErrorMessage = nil
		}

		jobs = append(jobs, j)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (s *Store) GetPendingJobs(ctx context.Context) ([]job.Job, error) {
	query := `
		SELECT id, job_type, payload, status, attempts, error_message
		FROM jobs
		WHERE status IN ('pending', 'running')
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []job.Job

	for rows.Next() {
		var j job.Job
		var errMsg sql.NullString

		err := rows.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &errMsg)
		if err != nil {
			return nil, err
		}
		if errMsg.Valid {
			j.ErrorMessage = &errMsg.String
		} else {
			j.ErrorMessage = nil
		}

		jobs = append(jobs, j)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

