# conveyor

Async job queue and task runner in Go. Jobs are submitted over HTTP, processed concurrently by a worker pool, and persisted in PostgreSQL.

---

## How it works

- API handlers accept HTTP requests (`POST /jobs`, `GET /jobs`, etc.)
- Submitted jobs are passed to the dispatcher
- Dispatcher stores jobs in PostgreSQL and pushes them into the in-memory queue
- Worker pool consumes jobs concurrently from the queue
- Workers mark jobs as `running`, process the task, then update PostgreSQL with the result
- Failed jobs retry automatically up to 3 times with exponential backoff
- On startup, unfinished jobs are recovered from PostgreSQL back into the queue
- Graceful shutdown waits for in-flight workers to finish before exiting

**Job lifecycle:**

```
pending → running → done
                 ↘ failed → retry (up to 3x)
```

---

## Project structure

```
.
├── cmd
│   ├── server          # HTTP server + worker pool
│   └── watch-jobs      # Terminal dashboard
├── internal
│   ├── api             # HTTP handlers
│   ├── dispatcher      # Coordinates queue, store, and workers
│   ├── job             # Core Job type and status constants
│   ├── queue           # In-memory buffered channel queue
│   ├── store           # PostgreSQL persistence
│   └── worker          # Worker pool and job simulation
└── schema.sql
```

---

## Setup

### 1. Start PostgreSQL

```bash
docker run --name conveyor-pg \
  -e POSTGRES_USER=conveyor \
  -e POSTGRES_PASSWORD=conveyor \
  -e POSTGRES_DB=conveyor \
  -p 5432:5432 \
  -d postgres:16
```

### 2. Apply schema

```bash
psql postgres://conveyor:conveyor@localhost:5432/conveyor?sslmode=disable \
  -f schema.sql
```

### 3. Run the server

```bash
go run ./cmd/server
```

By default, connects to `postgres://conveyor:conveyor@localhost:5432/conveyor?sslmode=disable`.
Override with the `DATABASE_URL` environment variable:

```bash
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable go run ./cmd/server
```

### 4. Run the dashboard

```bash
go run ./cmd/watch-jobs
```

---

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/jobs` | Submit a job |
| `GET` | `/jobs/{id}` | Get job by ID |
| `GET` | `/jobs` | List all jobs |

### Submit a job

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"sendEmail","payload":{"to":"user@example.com"}}'
```

### Response

```json
{
  "id": "89cb107d-...",
  "message": "job created"
}
```

---

## Job types

Workers are intentionally stubbed to demonstrate the infrastructure — queueing, persistence, retries, concurrency, and coordination — rather than real task execution.

| Type | Simulated duration | Failure rate |
|------|--------------------|--------------|
| `sendEmail` | ~2s | 20% |
| `resizeImage` | ~3s | 15% |
| `generateReport` | ~5s | 10% |

---

## watch-jobs

A live terminal dashboard that redraws in-place using ANSI escape codes. No scrolling.

```
Conveyor - 4 jobs

       ID          Job Type          Status    Attempts

 [✓]   89cb107d    sendEmail         done      1/3
 [✓]   fd39dc75    generateReport    done      1/3
 [✗]   929d84c1    sendEmail         failed    3/3
 [/]   090c5cc4    resizeImage       running   1/3
```

- Running jobs show a cycling `|`, `/`, `-`, `\` spinner
- Completed jobs show `✓`
- Permanently failed jobs show `✗`

---

## Notes

- In-memory queue backed by PostgreSQL persistence
- No auth or rate limiting
- No dead-letter queue
- No distributed coordination or horizontal scaling

For production-grade systems, see [Asynq](https://github.com/hibiken/asynq) or [River](https://github.com/riverqueue/river).
