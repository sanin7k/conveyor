# conveyor

Async job queue and task runner in Go. Jobs are submitted over HTTP, processed by a worker pool, and persisted in PostgreSQL.

---

## How it works

```
POST /jobs → [channel] → [worker pool] → PostgreSQL
                                              │
                                         GET /jobs/:id
```

- Jobs move through: `pending → running → done / failed`
- Failed jobs retry up to 3 times automatically
- On startup, unfinished jobs are recovered from PostgreSQL
- Graceful shutdown waits for in-flight workers before exiting

---

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/jobs` | Submit a job |
| GET | `/jobs/:id` | Get job status |
| GET | `/jobs` | List all jobs |

**Submit a job:**
```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type": "send_email", "payload": {"to": "user@example.com"}}'
```

---

## Job types

Workers are intentionally stubbed — the project demonstrates the infrastructure around job execution, not the execution itself.

| Type | Simulated duration | Failure rate |
|------|--------------------|--------------|
| `send_email` | ~2s | 20% |
| `resize_image` | ~3s | 15% |
| `generate_report` | ~5s | 10% |

---

## watch-jobs

A live terminal dashboard that updates in-place. One row per job, no scrolling.

```
CONVEYOR — 4 jobs

[✓]  send_email       def456   done      1.8s
[/]  resize_image     abc123   running   attempt 1/3
[✗]  send_email       jkl012   failed    attempt 3/3
[|]  generate_report  ghi789   running   attempt 1/3
```

Running jobs show a cycling `\|/-` spinner. Completed jobs show `✓`, permanently failed jobs show `✗`. Rewrites in place using ANSI escape codes.

---

## Setup

```bash
# Configure
export DATABASE_URL="postgres://localhost/conveyor?sslmode=disable"
export WORKER_COUNT=5
export PORT=8080

# Run
go run ./cmd/server

# Watch
go run ./cmd/watch-jobs
```

---

## Structure

```
conveyor/
├── cmd/
│   ├── server/      # HTTP server + worker pool
│   └── watch-jobs/  # Terminal dashboard
├── internal/
│   ├── queue/       # In-memory channel
│   ├── worker/      # Worker pool and job execution
│   ├── store/       # PostgreSQL persistence
│   └── api/         # HTTP handlers
└── schema.sql
```

---

## Design notes

- **In-memory channel over DB polling** — workers pick up jobs immediately. On crash, startup recovery reloads `pending` and `running` jobs from PostgreSQL.
- **`running` state written immediately** — distinguishes interrupted jobs from unstarted ones on recovery.
- **Fixed worker count** — explicit concurrency bound, predictable DB connection usage.
- **Graceful shutdown** — on `SIGINT`, stops intake, drains in-flight jobs, then exits cleanly.

---

> Not production-ready. No auth, no horizontal scaling, no dead-letter queue. For production, see [Asynq](https://github.com/hibiken/asynq) or [River](https://github.com/riverqueue/river).
