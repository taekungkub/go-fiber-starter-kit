# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go run cmd/main.go              # run the app (listens on :8080)
go build ./...                  # build everything
go test ./...                   # run all tests
go test ./pkg/common/...        # run tests for one package
go test -run TestHashPassword ./pkg/common/...  # run a single test
docker-compose up -d            # run Postgres + Redis + app via Docker
```

There is no separate lint config; `go vet ./...` is the closest available check. Environment variables are loaded from `.env` then `.env.<APP_ENV>` (see `config/config.go`); copy `.env.example` to `.env` before running locally.

Module path is `go-fiber-stater-kit` (note the typo — this is intentional and used in every import path, do not "fix" it).

## Architecture

Clean Architecture, organized by feature module under `internal/api/<module>/`. Each module (`auth`, `user`) follows the same 4-file layering, wired together in `cmd/main.go`:

- `router.go` — registers Fiber routes, applies `middleware.JWTMiddleware()` / `middleware.RequireRole(...)`
- `handler.go` — parses request, calls usecase, writes response via `pkg/core` helpers
- `usecase.go` — business logic
- `repository.go` — sqlx/pgx queries against Postgres
- `<module>.go` (e.g. `auth.go`, `user.go`) — DTOs and domain structs for that module

`cmd/main.go` is the composition root: it builds `db`/`redis` connections, runs migrations, starts background workers/ingest, then constructs each module's repository → usecase → handler → router chain manually (no DI framework).

### Auth & authorization

- JWT access (15 min) + refresh (7 days) tokens, config in `pkg/core/jwt.go`, secrets/expiry set from `config.Config` in `main.go` at startup (`core.AccessTokenSecret`, etc. are package-level vars).
- `middleware.JWTMiddleware()` validates the bearer token and puts `userID`, `email`, `role`, `token` into Fiber `c.Locals`.
- `middleware.RequireRole(roles...)` reads `c.Locals("role")` and gates access; roles are `ADMIN`, `STAFF`, `CUSTOMER`.
- Redis-backed token blacklist/session logic exists in comments/design but is not currently wired up in `main.go` (redis client construction is commented out) — check before assuming it's active.

### Ingest → Queue → Worker pipeline (background job processing)

This is a separate concern from the HTTP API, started in `main.go` alongside the Fiber app:

```
MQTT (mocked) → internal/ingest → internal/queue.JobQueue (buffered chan, cap 1000) → internal/worker (N goroutines) → batch (size 10) → DB
```

- `internal/ingest/mqtt_consumer.go`: `StartMQTTConsumerMock` ticks on an interval and pushes a fake `worker.Job` onto `queue.JobQueue`. There is no real MQTT client wired in yet — this simulates the producer.
- `internal/queue/queue.go`: exposes the single package-level `JobQueue` channel shared between ingest and workers.
- `internal/worker/worker.go`: `StartWorkers(db, jobQueue, n)` spawns `n` goroutines, each accumulating jobs into a local batch and flushing (currently just logs, no actual DB insert) once the batch reaches 10.
- See `MQTT.md` for the intended data-flow diagram; treat it as a design note, not a guarantee of current implementation completeness (e.g. actual MQTT and batch DB writes are still stubbed).

### Data access

- Postgres via `sqlx` (`internal/database/postgres.go`); connection and table migration (`migrations/tb_users.go`, ad-hoc `CREATE TABLE IF NOT EXISTS`, no versioned migration tool) run at startup in `main.go`.
- Users table has soft delete (`deleted_at`), a generated `name` column from `first_name`/`last_name`, and a `user_role` enum.
- Redis client setup exists in `internal/database/redis.go` but is currently commented out in `main.go`.

### Request/response conventions

- Use `pkg/core.SendSuccess` / `pkg/core.SendError` / `pkg/core.SendValidationError` for all handler responses — keeps the `{error, message, data}` / `{error, message, errors}` envelope consistent.
- Use `pkg/core.ValidateStruct` (wraps `go-playground/validator`) for DTO validation in handlers before calling into usecases.
- List endpoints use `pkg/core.Pagination` (page/limit/offset helper returning `Paging{List, Page, Limit, TotalPage, Total}`) and `pkg/core/sorting.go`/`params.go` for query param parsing.
