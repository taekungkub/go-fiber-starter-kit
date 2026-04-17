# Go Fiber Starter Kit — Project Structure

A clean, modular REST API starter kit using **Go + Fiber + PostgreSQL**, following a layered architecture pattern per domain module.

---

## Folder Structure

```
go-fiber-starter-kit/
├── cmd/
│   └── main.go                  # Entry point — wires all modules, middleware, and starts server
├── config/
│   └── config.go                # Loads env vars into Config struct via envconfig
├── internal/
│   ├── api/
│   │   ├── auth/                # Auth module (login, refresh, logout)
│   │   │   ├── auth.go          # Model, DTOs, Response types
│   │   │   ├── handler.go       # HTTP handlers (interface + struct)
│   │   │   ├── repository.go    # DB queries (interface + struct)
│   │   │   ├── usecase.go       # Business logic (interface + struct)
│   │   │   └── router.go        # Route registration
│   │   ├── user/                # User CRUD module
│   │   │   ├── user.go
│   │   │   ├── handler.go
│   │   │   ├── repository.go
│   │   │   ├── usecase.go
│   │   │   └── router.go
│   │   └── ask/                 # AI/LLM module (OpenAI)
│   │       ├── ask.go
│   │       ├── handler.go
│   │       ├── repository.go
│   │       ├── usecase.go
│   │       └── router.go
│   ├── database/
│   │   ├── postgres.go          # sqlx PostgreSQL connection
│   │   └── redis.go             # Redis store + client
│   └── rag/                     # RAG (Retrieval-Augmented Generation) utilities
│       ├── rag.go
│       ├── embed.go
│       ├── search.go
│       └── store.go
├── middleware/
│   ├── jwt_middleware.go        # Bearer token validation, attaches userID/email/role to ctx
│   └── role_middleware.go       # RequireRole("ADMIN") guard
├── migrations/
│   └── 001_init.sql             # SQL schema — run manually or via migrate tool
├── pkg/
│   ├── core/
│   │   ├── response.go          # SuccessResponse, ErrorResponse, SendSuccess/SendError helpers
│   │   ├── paging.go            # Paging struct + Pagination() helper
│   │   ├── params.go            # Query param helpers
│   │   ├── request.go           # PagingRequest(), SortingRequest()
│   │   ├── sorting.go           # Sorting struct
│   │   └── jwt.go               # JWT generation + validation (access & refresh)
│   ├── common/
│   │   ├── common.go            # Shared utilities
│   │   ├── common_test.go
│   │   ├── bcrypt.go            # HashPassword / ComparePassword
│   │   └── bcrypt_test.go
│   └── timex/
│       ├── timex.go             # Time utilities
│       └── timex_test.go
├── .env.example                 # Environment variable template
├── .air.toml                    # Air live-reload config
├── docker-compose.yml           # PostgreSQL + Redis services
├── Dockerfile
└── go.mod
```

---

## Module Pattern (per domain)

Every domain under `internal/api/<module>/` follows the same 5-file structure:

| File | Purpose |
|---|---|
| `<module>.go` | Model, constants, DTOs (Create/Update), Response type, `ToResponse()` |
| `repository.go` | `Repository` interface + `repository` struct — raw DB queries only |
| `usecase.go` | `UseCase` interface + `useCase` struct — business logic, calls repo |
| `handler.go` | `Handler` interface + `handler` struct — HTTP layer, calls usecase |
| `router.go` | Route registration function, applies middleware per route |

### Dependency flow

```
router.go
  └── handler.go    (depends on UseCase)
        └── usecase.go  (depends on Repository)
              └── repository.go  (depends on *sqlx.DB)
```

### Wiring in `main.go` (manual DI)

```go
repo    := user.NewRepository(db)
useCase := user.NewUseCase(repo)
handler := user.NewHandler(useCase)
user.UserRouter(api, handler)
```

---

## Key Packages

### `pkg/core`

| Helper | Usage |
|---|---|
| `SendSuccess(c, msg, data)` | Standard `200` success response |
| `SendError(c, status, msg)` | Standard error response |
| `SendValidationError(c, errs)` | `400` with field-level errors |
| `ValidateStruct(&dto)` | Returns `[]ValidationError` via go-playground/validator |
| `PagingRequest(c, defaultLimit)` | Parses `?page=&limit=` from query |
| `SortingRequest(c, col, dir)` | Parses `?sort=&order=` from query |
| `Pagination(page, limit, countFn, dataFn)` | Returns `Paging` struct |

### `pkg/core` — Response shapes

```json
// Success
{ "error": false, "message": "...", "data": { ... } }

// Error
{ "error": true, "message": "..." }

// Validation error
{ "error": true, "message": "Validation failed", "errors": [{ "field": "Email", "message": "..." }] }

// Paged list
{ "list": [...], "page": 1, "limit": 20, "total_page": 5, "total": 100 }
```

---

## Middleware

```go
// Protect all routes in a group
users := app.Group("/users", middleware.JWTMiddleware())

// Protect specific routes by role
users.Post("/", middleware.RequireRole("ADMIN"), handler.Create)
```

After `JWTMiddleware` runs, locals are available in handlers:
```go
userID := c.Locals("userID").(string)
role   := c.Locals("role").(string)
```

---

## Config & Environment

`config.go` uses `envconfig` + `godotenv`. Loads `.env` then `.env.<APP_ENV>`.

```env
APP_ENV=dev
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=password
DB_NAME=mydb
JWT_SECRET=change-me
REFRESH_SECRET=change-me
JWT_EXPIRES_IN_ACCESS_TOKEN=15       # minutes
JWT_EXPIRES_IN_REFRESH_TOKEN=10080   # minutes (7 days)
REDIS_HOST=localhost
REDIS_PORT=6379
OPENAI_KEY=sk-...
```

---

## Database

- **ORM**: `sqlx` with raw SQL — no ORM abstraction
- **Driver**: `pgx/v5` (PostgreSQL)
- **Soft delete**: `deleted_at TIMESTAMPTZ` — all queries filter `WHERE deleted_at IS NULL`
- **Primary key**: `UUID` via `gen_random_uuid()`
- **Generated column**: `name` = `first_name || ' ' || last_name` (STORED)
- **Migrations**: Plain `.sql` files in `migrations/` — run manually

---

## Adding a New Module

1. Create folder `internal/api/<module>/`
2. Add `<module>.go` — define Model, DTOs, Response, `ToResponse()`
3. Add `repository.go` — define interface + DB queries
4. Add `usecase.go` — define interface + business logic
5. Add `handler.go` — define interface + HTTP handlers
6. Add `router.go` — register routes with middleware
7. Wire in `cmd/main.go`:
```go
repo    := mymodule.NewRepository(db)
useCase := mymodule.NewUseCase(repo)
handler := mymodule.NewHandler(useCase)
mymodule.MyRouter(api, handler)
```

---

## Tech Stack

| Library | Version | Purpose |
|---|---|---|
| `gofiber/fiber/v2` | v2.52 | HTTP framework |
| `jmoiron/sqlx` | v1.4 | SQL with struct scanning |
| `jackc/pgx/v5` | v5.8 | PostgreSQL driver |
| `golang-jwt/jwt/v5` | v5.3 | JWT access + refresh tokens |
| `go-playground/validator/v10` | v10.30 | Struct validation |
| `redis/go-redis/v9` | v9.17 | Redis client + rate limit store |
| `joho/godotenv` | v1.5 | `.env` file loading |
| `kelseyhightower/envconfig` | v1.4 | Env → struct mapping |
| `sashabaranov/go-openai` | v1.41 | OpenAI API client |
| `air` | — | Live reload (dev) |
