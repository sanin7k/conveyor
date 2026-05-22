# conveyor

Async job queue and task runner in Go. Jobs are submitted over HTTP, processed by a worker pool, and persisted in PostgreSQL.

---

## How it works

```
POST /jobs → [channel] → [worker pool] → PostgreSQL
                                              │
                                         GET /jobs/{id}
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
| GET | `/jobs/{id}` | Get job status |
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
Conveyor -  4  jobs

       ID          Job Type          Status    Attempts  

 [✓]   89cb107d    sendEmail         done      1/3
 [✓]   fd39dc75    generateReport    done      1/3
 [✗]   929d84c1    sendEmail         failed    3/3
 [/]   090c5cc4    resizeImage       running   0/3
```

Running jobs show a cycling `\|/-` spinner. Completed jobs show `✓`, permanently failed jobs show `✗`. Rewrites in place using ANSI escape codes.

---

## Setup

```bash
# Run
go run ./cmd/server

# Watch
go run ./cmd/watch-jobs
```

---

> Not production-ready. No auth, no horizontal scaling, no dead-letter queue. For production, see [Asynq](https://github.com/hibiken/asynq) or [River](https://github.com/riverqueue/river).
