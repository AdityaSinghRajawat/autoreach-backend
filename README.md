# AutoReach Backend

AutoReach automates a job search for candidates. It collects job postings, matches them against a candidate's resume, finds the recruiters and hiring managers behind each matching job, reaches out to them by email and LinkedIn, prepares job applications, and reports everything on a dashboard.

This repository is the Go backend: a REST API and a background worker built from one binary. The Next.js frontend lives in its own repository.

**Contents:** [How it works](#how-it-works) · [Tech stack](#tech-stack) · [Quick start](#quick-start) · [Configuration](#configuration) · [Daily development](#daily-development) · [Testing and quality](#testing-and-quality) · [Project layout](#project-layout) · [Documentation](#documentation) · [Troubleshooting](#troubleshooting) · [Status](#status)

## How it works

```
 job boards (Greenhouse, Lever, Ashby,      external feeds
 SmartRecruiters, Workday, Adzuna,          (POST /api/v1/jobs/ingest)
 career pages)                                       │
        │  polled by the worker                      │
        └──────────────────────┬─────────────────────┘
                               ▼
                         job postings ──► embeddings (AI)     ◄── resumes (PDF/DOCX)
                               │                                      │
                               └──────────► matching (pgvector) ◄─────┘
                                                  │
                                ┌─────────────────┼──────────────────┐
                                ▼                 ▼                  ▼
                       contact enrichment   assisted apply     dashboard
                       (Apollo, SignalHire, (Greenhouse,       (metrics, funnel,
                        ContactOut)          Lever)             live activity)
                                │
                                ▼
                       outreach: email campaigns (SMTP/IMAP)
                       and assisted LinkedIn invites and messages
```

The binary runs in one of three modes:

| Command | What it runs |
|---|---|
| `autoreach` or `autoreach server` | The HTTP API: `/api/v1/*`, `/health`, `/metrics`, `/swagger` |
| `autoreach worker` | Queue consumers and scheduled jobs (polling, sending, reply checks, reconciliation), plus `/health` and `/metrics` |
| `autoreach create-admin <email>` | One-off: creates an administrator; the password is read from `ADMIN_PASSWORD` |

The API only records work and queues tasks; the worker does everything slow or external (embedding, enrichment, sending, polling). The database is the source of truth, so a lost task is re-queued by a reconciler.

## Tech stack

| Concern | Choice |
|---|---|
| Language | Go 1.26 (the toolchain pins 1.26.11; `go` downloads it automatically) |
| HTTP | gin, with an OpenAPI 3 spec in `docs/swagger/openapi.json` |
| Database | PostgreSQL 16 with pgvector, accessed with pgx and plain SQL; migrations with golang-migrate |
| Queue and cache | Redis 7; tasks with asynq; rate limits, locks and counters in Redis |
| AI | Any OpenAI-compatible API (OpenAI by default, or Gemini and others): embeddings for matching and a chat model for resume parsing |
| Files | Local disk or S3-compatible storage |
| Observability | Structured JSON logs (slog) and Prometheus metrics |

## Quick start

### 1. Install the prerequisites

- **Go 1.26 or newer**: <https://go.dev/dl/>
- **Docker** with Compose (Docker Desktop on Windows and macOS)
- **make**: preinstalled on Linux and macOS. On Windows, use Git Bash with make installed (`choco install make` or `winget install GnuWin32.Make`), or run the commands each target prints with `make -n <target>`.
- **An AI provider API key** (OpenAI, Gemini or another OpenAI-compatible API) if you want matching to work (see step 2).

### 2. Configure

```bash
cp .env.example .env
```

Open `.env` and set:

- `JWT_SECRET`: any random string of at least 32 characters, for example the output of `openssl rand -base64 48`. The app refuses to start without it.
- `AI_API_KEY`: without it everything runs, but jobs and resumes cannot be embedded, so no matches appear.

The database and Redis defaults already match the local containers. Every other setting is optional and documented in place in [`.env.example`](.env.example); see [Configuration](#configuration).

### 3. Start Postgres and Redis

```bash
make dev-up
```

This starts PostgreSQL (with pgvector) on `localhost:5432` and Redis on `localhost:6379`, then applies every migration in `migrations/`.

### 4. Start the API and the worker

```bash
make run-server                 # terminal 1: API on http://localhost:8080
PORT=8081 make run-worker       # terminal 2: worker; its /health and /metrics on :8081
```

Both read `.env`. The worker needs its own port because both processes serve `/health` and `/metrics`.

### 5. Check it is running

```bash
curl http://localhost:8080/health          # {"success":true,...,"data":{"status":"ok"}}
```

The interactive API reference is at <http://localhost:8080/swagger/index.html>.

### 6. Create an administrator

```bash
ADMIN_PASSWORD='choose-a-long-password' go run ./cmd/autoreach create-admin admin@example.com
```

Sign in with `POST /api/v1/auth/login` to get an admin token. Admins manage job sources, API keys for job feeds, queues, suppressions and contact erasure under `/api/v1/admin/*`.

### 7. Try the main flow

Using Swagger or curl:

1. **Register a candidate:** `POST /api/v1/auth/register` with `email`, `password` (10+ characters) and `fullName`. The response holds an `accessToken` and the `user.candidateId`.
2. **Upload a resume:** `POST /api/v1/candidates/{candidateId}/resumes` as multipart with the field `file` (PDF or DOCX). The worker parses and embeds it.
3. **Add jobs**, either way:
   - as admin, add a job board: `POST /api/v1/admin/job-sources` with `{"boardType":"GREENHOUSE","boardToken":"stripe","companyName":"Stripe","companyDomain":"stripe.com"}`; the worker polls it right away; or
   - as admin, create an API key (`POST /api/v1/admin/api-clients` with `{"name":"feed","scopes":["job:ingest:any"]}`) and push jobs to `POST /api/v1/jobs/ingest` with the `X-API-Key` header.
4. **See matches:** `GET /api/v1/candidates/{candidateId}/jobs`, and the dashboard at `GET /api/v1/dashboard/funnel`.

Contact enrichment, email outreach and assisted applications need provider keys, a connected mailbox or an application profile; [docs/FEATURES.md](docs/FEATURES.md) explains each.

To stop the containers: `make dev-down`.

## Configuration

All settings come from environment variables, normally through `.env`. [`.env.example`](.env.example) lists every variable the code reads, grouped into numbered sections with an explanation for each, so it is the reference.

What you need to set depends on what you want to use:

| To use | Set (section in `.env.example`) |
|---|---|
| Anything at all | `JWT_SECRET` (6) |
| Matching resumes to jobs | `AI_API_KEY` (9) |
| Finding recruiters and hiring managers | At least one of `HUNTER_API_KEY` (free plan available), `APOLLO_API_KEY`, `SIGNALHIRE_API_KEY` (plus `PUBLIC_BASE_URL` and `SIGNALHIRE_WEBHOOK_SECRET`), `CONTACTOUT_API_KEY` (12) |
| Connecting a candidate's mailbox for email outreach | `CHANNEL_CREDENTIAL_KEYS` (13); `PUBLIC_BASE_URL` (1) for one-click unsubscribe and tracking links |
| Adzuna job searches | `ADZUNA_APP_ID`, `ADZUNA_APP_KEY` (11) and an Apollo key for company domains |
| Submitting applications through an ATS | `GREENHOUSE_BOARD_KEYS` or `LEVER_SITE_KEYS`, issued by the employer (15) |
| Storing resumes in S3, R2 or MinIO | `STORAGE_DRIVER=s3` and the `S3_*` settings (18) |
| Production | `APP_ENV=production`, `INTERNAL_SHARED_SECRET`, `CHANNEL_CREDENTIAL_KEYS`, real database and Redis passwords; see [docs/RUNBOOK.md](docs/RUNBOOK.md) |

Real environment variables override `.env`, which is handy for one-off changes such as `PORT=8081`.

## Daily development

| Task | Command |
|---|---|
| Start or stop Postgres and Redis | `make dev-up` / `make dev-down` |
| Run the API / the worker | `make run-server` / `PORT=8081 make run-worker` |
| Reload the API on every change | `make dev` (needs [air](https://github.com/air-verse/air); `air -- worker` reloads the worker) |
| Apply or roll back one migration | `make migrate-up` / `make migrate-down` |
| Build the binary into `bin/` | `make build` |
| Run the whole production stack in Docker | `make docker-up` (needs the production secrets; see [docs/RUNBOOK.md](docs/RUNBOOK.md)) |

Each table has one pair of files in `migrations/` (`NNNN_name.up.sql` and `.down.sql`) holding the table and all its indexes. Until the first production deploy, change a table by editing its file and recreating your local database; after that, every change is a new pair. New endpoints must be added to `docs/swagger/openapi.json`, or the tests fail. Before you write code, read [docs/CONVENTIONS.md](docs/CONVENTIONS.md): the project follows a strict layout and set of rules.

## Testing and quality

```bash
make test-unit          # unit tests, no infrastructure needed
make test-up            # start throwaway Postgres :55432, Redis :56379, S3 mock :59090, mail server :53025/:53143
make test-integration   # integration tests against those containers (runs one package at a time)
make test-down          # stop and delete the test containers

make fmt                # format code (run before committing)
make lint               # what CI checks: formatting plus linters
make check              # fmt, vet, lint, unit tests and build
make tools              # install the pinned golangci-lint

make vuln               # known vulnerabilities in dependencies (govulncheck)
make secrets-scan       # secrets in git history (gitleaks)
make test-live          # read real public job boards (LIVE_JOB_BOARDS=TYPE:token,...)
make loadtest-api       # API latency against a running stack (scripts/README.md)
make loadtest-apply     # application throughput against a stub ATS
```

Integration tests use `.env.test` and refuse to run unless the database name ends in `_test`, because they empty every table. They rebuild the schema from `migrations/` on every run. CI (`.github/workflows/ci.yml`) runs the build, lint, unit and integration tests, the vulnerability and secret scans, and an image build on every pull request.

## Project layout

```
cmd/autoreach        process entry point and mode selection (server, worker, create-admin)
config               every environment variable, constant, error and table name
routes               route tables, under routes/v1
middlewares          logging, errors, authentication, authorisation, rate limits, CORS, metrics
pkg/dependencies     builds repositories, services and handlers once per process
pkg/handlers         HTTP in and out; one handler per file
pkg/services         use cases; one per file
pkg/repositories     SQL, one file per table
pkg/models           database entities and their enums (pkg/models/helpers)
pkg/customtypes      request, response and other non-entity shapes
pkg/integrations     clients for third-party HTTP APIs (AI provider, enrichment providers, ATS and job boards)
pkg/storage          file storage: local disk or S3
pkg/emailTemplates   HTML templates for outgoing email
pkg/db               connection pools and transactions
pkg/utils            shared helpers: logging, HTTP, crypto, rate limiter, locks, email, web fetching
pkg/validations      custom validation rules
worker               task definitions, queue adapter (asynq) and task handlers
cronJobs             scheduled jobs, each guarded by a Redis lock
migrations           golang-migrate SQL pairs
docs                 OpenAPI spec, features, conventions, runbook, security checklist
scripts              benchmarks, calibration and load tests (scripts/README.md)
tests                unit, integration and live tests, with setup, helpers and mocks
```

A request flows `routes` → `middlewares` → `pkg/handlers` → `pkg/services` → `pkg/repositories` / `pkg/integrations`. Worker tasks in `worker/tasks` call the same services.

## Documentation

| Document | Read it for |
|---|---|
| [docs/FEATURES.md](docs/FEATURES.md) | How each feature behaves: authorisation, rate limits, email and LinkedIn outreach, job sources, applications, dashboard, storage |
| [docs/CONVENTIONS.md](docs/CONVENTIONS.md) | The rules every change follows |
| [docs/RUNBOOK.md](docs/RUNBOOK.md) | Deploying and operating production: secrets, backups, monitoring, incidents, performance |
| [docs/SECURITY.md](docs/SECURITY.md) | The security checklist and the evidence for each control |
| [scripts/README.md](scripts/README.md) | Vector benchmark, match calibration and load tests |
| `/swagger/index.html` | The API reference, served by a running server from `docs/swagger/openapi.json` |

## Troubleshooting

| Symptom | Fix |
|---|---|
| `JWT_SECRET must be at least 32 characters` | Set `JWT_SECRET` in `.env` |
| The worker exits because it cannot listen on port 8080 | The API already uses it; start the worker with `PORT=8081` |
| `connect postgres` or `connect redis` errors | Run `make dev-up` and check `docker ps`; the defaults expect ports 5432 and 6379 to be free |
| Resumes and jobs stay unmatched | Set `AI_API_KEY`, then retry the failed tasks (`POST /api/v1/admin/queues/{queue}/dead/{taskId}/retry`) or upload again |
| Enrichment finds no contacts | No provider key is set, or a provider's daily credit budget is used up (`GET /api/v1/admin/providers`) |
| Connecting a mailbox fails | Set `CHANNEL_CREDENTIAL_KEYS` |
| Migrations report a "dirty" database, or the schema changed under you | A migration failed half way, or an existing migration was edited. During development, reset the local database: `docker compose -f docker-compose.dev.yml down -v` then `make dev-up` (this deletes local data). In production, follow docs/RUNBOOK.md |
| Integration tests are skipped or fail to connect | Run `make test-up` first; set `TEST_REQUIRE_SERVICES=1` to turn a missing service into a failure |
| `go test -race` fails to build on Windows | The race detector needs a 64-bit C compiler; CI runs it on Linux |

## Status

Implemented and covered by tests:

- **Foundation:** authentication with refresh tokens, role-based authorisation checked on every route, rate limits, API clients for job feeds.
- **Jobs:** ingestion with deduplication, polling of Greenhouse, Lever, Ashby, SmartRecruiters and Workday boards, Adzuna searches and career pages (boards checked against live sites with `make test-live`), batched embeddings.
- **Candidates:** resume upload with PDF/DOCX parsing, AI resume parsing, embeddings, and matching in both directions.
- **Contacts:** enrichment through Hunter, Apollo, ContactOut and SignalHire with rate limits, daily credit budgets and circuit breakers; contacts kept current from replies, bounces, LinkedIn responses and re-enrichment; erasure on request and optional anonymizing retention.
- **Outreach:** email campaigns with send windows, daily caps, reply and bounce handling, unsubscribe and optional tracking; assisted LinkedIn invites and messages.
- **Applications:** assisted apply for Greenhouse and Lever.
- **Dashboard and operations:** metrics, funnel and a live activity feed; queue, dead-task, provider and suppression administration; Prometheus metrics per pipeline stage.
- **Production hardening:** hardened Docker image and Compose stack, dependency and secret scans in CI, runbook and security checklist. Measured: 1,800 API requests/s at p95 21 ms, and about 220,000 applications/hour against a stub ATS, with server and worker each under 100 MB.

Not yet verified against live services: the Hunter, Apollo, ContactOut and SignalHire clients and the Adzuna client (tested against stub servers only), and the vector search benchmark at 500,000 jobs, which needs a host with about 12 GB of RAM.
