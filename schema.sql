DROP TABLE IF EXISTS jobs;

CREATE TABLE jobs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  job_type TEXT NOT NULL,
  payload JSONB NOT NULL,

  status TEXT NOT NULL DEFAULT 'pending',

  attempts INTEGER NOT NULL DEFAULT 0,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,

  error_message TEXT
);

CREATE INDEX idx_jobs_status
ON jobs(status);

CREATE INDEX idx_jobs_job_type_status
ON jobs(job_type, status);

