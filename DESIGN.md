# Design Decisions

A record of non-obvious architectural choices made during the design of Conveyor.

---

## Dispatcher owns the queue

The queue is not exposed directly to the HTTP layer or any other package. All job submission goes through the dispatcher. This gives the dispatcher a single, controlled entry point for the write path — it can enforce that every job is persisted to PostgreSQL before it enters the in-memory queue. When persistence was added, no other package needed to change.

## Separate dispatcher and worker pool

The worker pool is responsible for concurrency — spinning up goroutines, consuming jobs, and sending status updates. The dispatcher is responsible for coordination — wiring the queue, the store, and the workers together, and owning the retry and recovery logic. Keeping these separate means each has a clear, single responsibility.

## Context cancellation and Shutdown() are separate concerns

Context cancellation (via OS signal) is a signal — it tells workers to stop picking up new jobs. `Shutdown()` is synchronization — it waits for in-flight workers to finish, closes the results channel, and waits for the results goroutine to drain. Without `Shutdown()`, the process would exit before in-flight jobs complete and before all results are written to PostgreSQL.

## Workers communicate via a status update channel

Workers send `StatusUpdate` structs to the dispatcher through a channel rather than calling the store directly. This keeps the worker pool decoupled from PostgreSQL — workers know nothing about persistence. The dispatcher receives all status transitions (`running`, `done`, `failed`) through the same channel and decides what to do with each.

## Attempts are only incremented on terminal states

The dispatcher increments `attempts` only when a job reaches `done` or `failed`, not when it transitions to `running`. This means `attempts` accurately reflects how many times a job has been fully processed, not how many times it was picked up.

## Retry updates status before sleeping

When a failed job is eligible for retry, the dispatcher updates its status to `pending` in PostgreSQL before sleeping for the backoff duration. If the status were updated after the sleep, the dashboard would show the job as `failed` during the entire backoff window, which is misleading. `pending` correctly signals that the job is waiting to be retried.

## Exponential backoff for retries

Retry delay is `2^attempts` seconds. This gives progressively longer gaps between retries without requiring any additional configuration. A job that fails on the first attempt waits 2 seconds before retry, 4 seconds after the second failure.

## watch-jobs hits the HTTP API, not PostgreSQL directly

The dashboard is a client of the server, not a second database consumer. This keeps PostgreSQL access consolidated in the server process, avoids credential distribution, and means the dashboard works across nodes without firewall changes to the database.

## HTTP handlers hold a direct reference to the store for reads

Write operations (job submission) go through the dispatcher. Read operations (`GET /jobs`, `GET /jobs/{id}`) go directly to the store from the HTTP handler. Routing reads through the dispatcher would have required reaching through dispatcher internals (`h.d.s.GetJob()`), which is a law of demeter violation. Reads have no queue involvement, so the dispatcher is not the right owner.

## Payload is opaque

The `Job` struct carries `payload` as `json.RawMessage`. Conveyor does not inspect or validate the payload — that is the responsibility of whoever integrates the worker functions. This keeps the infrastructure generic and the payload schema entirely in the consumer's control.

## Job type is a first-class field, not embedded in payload

`Type` is a top-level field on `Job` rather than a key inside the payload. Conveyor uses the type for routing, the store uses it for indexing, and the dispatcher may use it for per-type configuration in the future. Burying it inside the payload would require every layer to parse JSON just to answer basic questions about a job.

## UUID for job IDs

Jobs are identified by UUID generated at submission time in the API layer, before the job is written to PostgreSQL. This means the ID returned in the HTTP response is stable and does not depend on a database round-trip. PostgreSQL stores the UUID as the primary key.
