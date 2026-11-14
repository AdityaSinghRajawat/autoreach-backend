# AutoReach Backend

Go service for AutoReach. One binary, two process modes:

| Command | Mode |
|---|---|
| `autoreach` or `autoreach server` | HTTP API (`/api/v1`, `/health`, `/metrics`, `/swagger`) |
| `autoreach worker` | Queue consumers and cron jobs, plus `/health` and `/metrics` |
| `autoreach create-admin <email>` | One-off: creates an administrator. Password is read from `ADMIN_PASSWORD` |

Design documents live in `docs/`:

- `product_requirements_document.md` — what the product does
- `technical_specifications_document.md` — API contract, schema, stack
- `ARCHITECTURE_SPEC.md` — layering, naming and conventions this code follows
- `IMPLEMENTATION_PLAN.md` — milestones, decisions and open questions

## Requirements

- Go 1.26+
- Docker (PostgreSQL 16 with pgvector, Redis 7)

## Run locally

```bash
cp .env.example .env            # then set JWT_SECRET (32+ characters)
make dev-up                     # postgres + redis + migrations
make run-server                 # API on :8080
make run-worker                 # in a second terminal; set PORT=8081 for it
```

Create an administrator and an API key for a scraper:

```bash
ADMIN_PASSWORD='choose-a-long-password' go run ./cmd/autoreach create-admin admin@example.com
# log in at POST /api/v1/auth/login, then POST /api/v1/admin/api-clients
```

API reference: `http://localhost:8080/swagger/index.html` (source: `docs/swagger/openapi.json`).

## Tests

```bash
make test-unit                  # no infrastructure
make test-up                    # throwaway postgres :55432 and redis :56379
make test-integration           # real database, real queue; -p 1
make test-down
make tools                      # install the pinned golangci-lint
make fmt                        # format, wrap at 100 columns, fix blank-line rules
make lint                       # fail on anything fmt would change, plus linters
make check                      # fmt + vet + lint + unit + build
```

Integration tests refuse to run unless the database name ends in `_test`, because they truncate tables. They rebuild the schema from `migrations/` on every run.

## Full stack in Docker

```bash
DB_PASSWORD=... REDIS_PASSWORD=... JWT_SECRET=... docker compose up -d --build
```

Server is published on `127.0.0.1:8080`, worker ops endpoints on `127.0.0.1:8081`.

## Layout

```
cmd/autoreach      process bootstrap and mode selection
config             every env read, constant, error code and table name
routes             route tables per group, under routes/v1
middlewares        request logging, error translation, auth, tenancy, rate limit
pkg/dependencies   the one place repositories, services and handlers are built
pkg/handlers       HTTP in and out; one method per file
pkg/services       use cases; one per file
pkg/repositories   SQL, one file per table
pkg/models         persisted shapes and enums
pkg/customtypes    request, response and result shapes
pkg/integrations   third-party API clients
pkg/db             pools, transactions
pkg/utils          logging, HTTP helpers, hashing, tokens, rate limiter, locks
worker             queue contracts, publisher, asynq adapter, task handlers
cronJobs           scheduled jobs, each guarded by a Redis lock
migrations         golang-migrate SQL pairs
tests              unit, integration, setup, helpers, mocks
```

## Conventions that are easy to miss

- Every response uses one envelope: `{ success, statusCode, data, meta, error }`. JSON keys are camelCase.
- Malformed JSON is `400`; a payload that parses but fails validation is `422` with `error.details[]`.
- Error bodies carry a fixed public message per error code. Internal detail is logged, never returned.
- A candidate-scoped route returns `404`, not `403`, for another tenant's ID.
- Every route under `/api/v1` must be in `docs/swagger/openapi.json`; a unit test fails otherwise, and router tests validate real responses against the spec.
- Every worker task needs a `TaskConfig` and a registered handler; the worker refuses to start if they drift.
- The database is the source of truth for pipeline state. Queue tasks are idempotent, and a reconciler cron re-enqueues rows whose task was lost.
- Go source carries no comments. Reasoning belongs in `docs/`.
- Models in `pkg/models` are database entities only. Each has `New<Model>()` (generates the ID, stamps timestamps and defaults), `UpdateTimestamp()` and `Validate()`. Projections, requests and responses live in `pkg/customtypes`.
- One validation per stage: `utils.BindAndValidate` validates the request from its tags; the request's `To<Model>Model()` receiver only maps fields and never validates; the service calls `UpdateTimestamp()` then the model's `Validate()` before saving.
- A handler only reads the request, calls one service method and writes the response or error. Create endpoints convert the request to a model and pass the model; update endpoints pass the request, and the service loads the existing entity and converts it there.
- Saves are `Upsert(ctx, model)` and return the model they were given. IDs and defaults are set in the constructor, so nothing is read back from the database after a write.
- Repositories hold no duplicated query or scan code: one private finder per table, and `db.FindOne` / `pgx.CollectRows` for row mapping.
- Error messages live in `config/errors.go` as `config.ErrXxx`; no other package calls `errors.New`.
- Independent operations run through `utils.RunConcurrently`; fire-and-forget work through `utils.RunInBackground`. No bare `go` outside `pkg/utils` and bootstrap.
- Formatting is mechanical: run `make fmt` before committing. CI runs `make lint`. Format-on-save is a personal editor choice and is not part of the repository.

## Status

Implemented and covered by tests: foundation, authentication, candidate profiles, API clients, job ingestion with deduplication, batched job embedding, resume upload with PDF/DOCX parsing and embedding, and candidate-to-job matching in both directions (new resume against jobs, new job against resumes). See `docs/IMPLEMENTATION_PLAN.md` for what follows.
