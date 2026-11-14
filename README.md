# AutoReach Backend

Go service for AutoReach. One binary, two process modes:

| Command | Mode |
|---|---|
| `autoreach` or `autoreach server` | HTTP API (`/api/v1`, `/health`, `/metrics`, `/swagger`) |
| `autoreach worker` | Queue consumers and cron jobs, plus `/health` and `/metrics` |
| `autoreach create-admin <email>` | One-off: creates an administrator. Password is read from `ADMIN_PASSWORD` |

The product requirements, technical specification, architecture spec and implementation plan are working documents kept outside the repository.

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

## Authorisation

Roles are defined in code, in `config/authorizationPolicy.go`; there is no role enum in the database.

- **Permissions** are `resource:action` strings in one catalogue, for example `resume:update`.
- **Grants** give a permission at a scope, for example `resume:update:own`. `own` means the resource belongs to the caller's candidate; `any` means every resource. The scope is where attributes of the resource decide access (ABAC).
- **Roles** are named lists of grants. A user can hold several roles, stored in the `users.roles` array, and gets the union of their grants. Access tokens carry role names only.
- **API clients** hold grants from a separate allow-list (`job:ingest:any` today), checked by the same middleware.

Routes declare what they need with `middlewares.Authorize(config.PermissionXxx)`. A missing permission is `403`; another candidate's data is `404`. Using an `any` grant on a candidate's data is logged.

To add a role, add an entry to `rolePoliciesI`. To add a permission, add it to the catalogue and grant it to roles. A typo in either stops the server at startup. A new kind of access (for example "candidates assigned to me") is a new scope plus one case in `scopeAllows` in `middlewares/authorize.go`.

## File storage

`STORAGE_DRIVER` chooses where uploaded files go: `local` (the default, under `STORAGE_LOCAL_DIR`) or `s3` (configured by the `S3_*` settings). `S3_ENDPOINT` and `S3_USE_PATH_STYLE` cover S3-compatible services such as Cloudflare R2 or MinIO. Startup fails if `s3` is chosen and the bucket is missing or unreachable, or if the driver name is unknown. Files already on local disk are not copied across automatically.

## Tests

```bash
make test-unit                  # no infrastructure
make test-up                    # throwaway postgres :55432, redis :56379, S3 mock :59090
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
pkg/integrations   clients for third-party HTTP APIs (OpenAI, enrichment providers)
pkg/storage        file storage: local disk (default) or S3, chosen by STORAGE_DRIVER
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
- Go source carries no comments. Naming and small functions carry the meaning.
- Models in `pkg/models` are database entities only. Each has `New<Model>()` (generates the ID, stamps timestamps and defaults), `UpdateTimestamp()` and `Validate()`. Projections, requests and responses live in `pkg/customtypes`.
- One validation per stage: `utils.BindAndValidate` validates the request from its tags; the request's `To<Model>Model()` receiver only maps fields and never validates; the service calls `UpdateTimestamp()` then the model's `Validate()` before saving.
- A handler only reads the request, calls one service method and writes the response or error. Create endpoints convert the request to a model and pass the model; update endpoints pass the request, and the service loads the existing entity and converts it there.
- Saves are `Upsert(ctx, model)` and return the model they were given. IDs and defaults are set in the constructor, so nothing is read back from the database after a write.
- Repositories hold no duplicated query or scan code: one private finder per table, and `db.FindOne` / `pgx.CollectRows` for row mapping.
- Services, handlers, repositories, utilities and integrations declare no constants and no stray types. Constants live in `config` behind getters (headers, vendor wire values, lock names, validation messages); non-entity structs live in `pkg/customtypes/<domain>`. Enum types stay in `pkg/models/helpers`, task and queue names in `worker/configs`, and SQL column lists beside the queries that use them.
- Error messages live in `config/errors.go` as `config.ErrXxx`; no other package calls `errors.New`.
- Independent operations run through `utils.RunConcurrently`; fire-and-forget work through `utils.RunInBackground`. No bare `go` outside `pkg/utils` and bootstrap.
- Formatting is mechanical: run `make fmt` before committing. CI runs `make lint`. Format-on-save is a personal editor choice and is not part of the repository.

## Status

Implemented and covered by tests: foundation, authentication, candidate profiles, API clients, job ingestion with deduplication, batched job embedding, resume upload with PDF/DOCX parsing and embedding, candidate-to-job matching in both directions (new resume against jobs, new job against resumes), and decision-maker contact enrichment through Apollo, ContactOut and SignalHire with rate limits, a daily credit budget and circuit breakers. The three provider clients are tested against a stub server only and have not yet been run against the live APIs.
