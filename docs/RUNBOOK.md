# AutoReach backend runbook

Operating the production stack in `docker-compose.yml`: API server, worker, Postgres with pgvector, Redis and the migration job. Commands assume the repository root on the host.

## 1. First deployment

1. Copy `.env.example` to `.env` and fill in at least:

   | Variable | How to generate |
   |---|---|
   | `DB_PASSWORD`, `REDIS_PASSWORD` | `openssl rand -hex 24` |
   | `JWT_SECRET` (32+ characters) | `openssl rand -hex 32` |
   | `INTERNAL_SHARED_SECRET` | `openssl rand -hex 24`; protects `/metrics` and Swagger |
   | `CHANNEL_CREDENTIAL_KEYS` | `1:$(openssl rand -base64 32)` with `CHANNEL_CREDENTIAL_KEY_VERSION=1` |
   | `PUBLIC_BASE_URL` | the public address of the API; tracking and unsubscribe links in emails point at it |
   | `OPENAI_API_KEY`, provider keys | from each vendor |

   Compose refuses to start without the variables in the first four rows, and so does the app in production.
2. `docker compose up -d --build --wait`. The `migrate` job applies every migration before the server and worker start.
3. Create the first administrator:

   ```bash
   docker compose run --rm -e ADMIN_PASSWORD='a-long-password' server create-admin admin@example.com
   ```
4. Put a TLS-terminating reverse proxy in front of `127.0.0.1:8080` (the server is bound to loopback only) and list its address in `TRUSTED_PROXIES` so client IPs and rate limits are right. Route `/t/*` and `/u/*` (tracking and unsubscribe links) publicly as well as `/api/v1/*`.
5. Point Prometheus at `server:8080/metrics` and `worker:8080/metrics` with the header `X-Internal-Secret: <INTERNAL_SHARED_SECRET>`.

Files uploaded with `STORAGE_DRIVER=local` live in the `uploads` volume shared by server and worker. For more than one host, use `STORAGE_DRIVER=s3`.

## 2. Upgrades and migrations

1. Back up first (section 3).
2. `docker compose pull` (or `AUTOREACH_IMAGE=<registry>/autoreach-backend:<tag>`), then `docker compose up -d --wait`. New migrations run before the new containers start; server and worker drain in-flight work within `stop_grace_period` (45 s).
3. If a migration fails, the `migrate` container exits non-zero and the old containers keep running. Read `docker compose logs migrate`, fix, and run `up` again. golang-migrate marks the version dirty on failure: after fixing by hand, `docker compose run --rm migrate -path=/migrations -database=<url> force <version>`.
4. Rolling back the schema (`... down 1`) loses data in dropped columns; restore from backup instead when in doubt.

## 3. Backups and restore

| Data | Backup | Restore |
|---|---|---|
| Postgres | `docker compose exec -T postgres pg_dump -U autoreach -Fc autoreach > autoreach-$(date +%F).dump` daily, kept off-host | `pg_restore --clean --if-exists -d autoreach` into a stopped stack |
| Redis | AOF in the `redis-data` volume. Holds queued tasks and counters only; losing it loses queued work, which the reconciler re-queues from Postgres state | restore the volume, or start empty |
| Uploads | `uploads` volume (local driver) or the S3 bucket's versioning | copy the volume back |

Test a restore into a scratch stack (`docker compose -p restore-test ...`) at least monthly.

## 4. Rotating secrets

| Secret | Procedure | Effect |
|---|---|---|
| `JWT_SECRET` | change and restart server and worker; rotate only if it may be compromised | every session ends. Tracking and unsubscribe links are signed with keys derived from it, so links in emails already sent stop working: unsubscribes then arrive only as replies (the sweeper honours a reply with the subject "unsubscribe"), so watch the mailboxes for a few weeks |
| `CHANNEL_CREDENTIAL_KEYS` | append `2:<new key>`, set `CHANNEL_CREDENTIAL_KEY_VERSION=2`, restart | new and re-saved mailbox credentials use key 2; older ones stay readable with key 1. Remove key 1 only when `SELECT COUNT(*) FROM candidate_channel_accounts WHERE key_version = 1` is 0 |
| `INTERNAL_SHARED_SECRET` | change, restart, update the Prometheus scrape header | none for users |
| API keys (ingest clients) | create a new client (`POST /api/v1/admin/api-clients`), switch the feed, revoke the old one | none if done in that order |
| `DB_PASSWORD`, `REDIS_PASSWORD` | change in the database/Redis and in `.env`, then `up -d` | brief restart |

## 5. Monitoring

Useful Prometheus expressions:

| Signal | Expression | Act when |
|---|---|---|
| API errors | `sum(rate(autoreach_http_requests_total{status=~"5.."}[5m])) / sum(rate(autoreach_http_requests_total[5m]))` | above 1% |
| API latency | `histogram_quantile(0.95, sum by (le) (rate(autoreach_http_request_duration_seconds_bucket[5m])))` | above 0.5 s |
| Failing stage | `sum by (stage) (rate(autoreach_worker_tasks_total{outcome="failed"}[15m]))` | sustained above 0 |
| Retry storm | `sum by (stage) (rate(autoreach_worker_tasks_total{outcome="retry"}[15m]))` | rising for 30 min |
| Queue backlog | `sum by (queue) (autoreach_queue_tasks{state="pending"})` | growing for 15 min |
| Dead tasks | `sum by (queue) (autoreach_queue_tasks{state="dead"})` | any new |

Container health: `docker compose ps` (both app containers have health checks on `/health`) and `docker stats` (see section 8 for normal memory).

## 6. Routine operations (admin API)

All under `/api/v1/admin`, with an admin token.

- **Dead tasks:** `GET /queues` shows `dead` per queue; `GET /queues/{queue}/dead` shows the payload and last error; fix the cause, then `POST /queues/{queue}/dead/{taskId}/retry`.
- **Provider status:** `GET /providers` for enrichment credits used today against each daily budget, and cold-email deliverability over 7 days. Below 95% deliverability, pause campaigns and review bounces.
- **Suppressions:** `GET/POST /suppressions`, `DELETE /suppressions/{id}`. Bounces and unsubscribes are added automatically.
- **Erasure requests** (a contact asks to be forgotten): find their records with `GET /contacts?email=...` (or `linkedinUrl=...`), then `DELETE /contacts/{contactId}` for each. Each delete removes the contact with their job links, messages and events, and adds the email to the global suppression list so it is never enriched into outreach again. Keep the request ticket; nothing else about the person remains. AutoReach logs only route templates, never query strings, but configure the reverse proxy not to log query strings either.
- **Contacts are kept up to date, not deleted:** replies confirm an address (or replace it with a new one at the same company), LinkedIn acceptances and replies confirm the profile, a bounce drops the address and queues `task:refresh_contact` to find the current one, and enrichment records every changed title, email, phone or company in the contact's history. An address that bounced is never adopted again.
- **Retention (off by default):** with `CONTACT_RETENTION_PERIOD` set (for example `17520h` for EU contacts), the worker daily removes the personal details of contacts not refreshed, messaged or heard from within that period and blanks the text sent to them; the rows stay, so outreach history and metrics are unchanged.
- **Job sources:** `GET /job-sources` shows each board's last poll, job count and last error.

## 7. Incidents

| Symptom | Likely cause | Response |
|---|---|---|
| Enrichment stops finding contacts | provider credit budget used up, rate limited or circuit open | `GET /providers`; raise `<PROVIDER>_DAILY_CREDITS` or wait for the daily reset; check the vendor's status page |
| Bounce rate jumps | bad contact data or a burned mailbox | pause campaigns (`PATCH` campaign `status: PAUSED`), check `GET /providers` deliverability, review the mailbox in the provider's console |
| `outreach_queue` backlog | send windows or daily caps holding messages | expected outside send hours; check `OUTREACH_SEND_WINDOW_*` and the 30/day cap |
| `embedding_queue` dead tasks | OpenAI key, quota or outage | fix the key or wait, then retry the dead tasks |
| Worker restarts in a loop | missing production setting | `docker compose logs worker`; the first line names the missing variable |
| Applications stuck in `QUEUED` | prepare tasks exhausted their retries while an ATS was down | the stuck-work reconciler re-queues them; or retry the dead tasks |
| Redis lost | volume loss | start Redis empty, then restart server and worker so the queues are listed again in `GET /admin/queues` (tasks are processed either way); the reconciler and cron jobs rebuild queued work from Postgres |

Scaling the worker: run a second worker container with no host port (copy the `worker` service without `ports`) or move to an orchestrator; every cron job takes a Redis lock, so several workers do not double-run them.

## 8. Performance baselines (2026-11-16)

Measured on one host (8 CPUs, 4 GB RAM for Docker) with the production Compose stack and `scripts/loadtest`:

| Measure | Result | Target |
|---|---|---|
| API reads, 16 clients, 2 min | 1,843 req/s, p50 6.7 ms, p95 20.5 ms, p99 39 ms, no errors | — |
| API reads while the worker applies | 592 req/s, p95 71 ms, no errors | — |
| Applications against a stub ATS (N4) | 1,000 in 16 s (about 220,000/hour); 3,000 in 13 s | 1,500/hour |
| Server memory (N5) | 35 MB idle, 97 MB peak | under 350 MB |
| Worker memory (N5) | 37 MB idle, 42 MB peak | under 350 MB |

Against real ATSs the limit is their rate limits and `ATS_RATE_LIMIT` (30 per minute per site by default), not AutoReach.
