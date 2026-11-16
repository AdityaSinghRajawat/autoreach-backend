# Scripts

One-off operational tools. Each has its own `main` and is not part of the service.

## vectorBench

Measures vector search latency and recall on synthetic data, using the same query shape, index parameters and `ef_search` as the service.

```bash
make dev-up
make vector-bench ROWS=50000
```

It creates its own table (`vector_bench_jobs`), fills it with random 1536-dimension vectors generated inside Postgres, builds the same partial HNSW index as `job_postings`, times the searches, compares a sample against an exact scan for recall, and drops the table. Pass `-keep` to keep it.

Sizing: 500,000 rows need roughly 4 GB for the table and 4 GB for the index, and the index has to fit in memory to meet the latency target. Give Postgres at least 12 GB before running at that size.

Random vectors are close to a worst case for an approximate index (no cluster structure), so real embeddings usually search faster and with higher recall.

## calibrate

Derives `MATCH_CALIBRATION_LOW` and `MATCH_CALIBRATION_HIGH` from resume and job pairs you have judged yourself.

```bash
OPENAI_API_KEY=... make calibrate PAIRS=pairs.jsonl
```

`pairs.jsonl` holds one JSON object per line:

```json
{"resume": "full resume text", "job": "full job description", "label": "good"}
{"resume": "full resume text", "job": "unrelated job description", "label": "bad"}
```

Use at least 30 pairs of each label. The tool embeds each pair with the configured model, prints the raw cosine similarity distribution per label, and suggests the two settings so that three quarters of the good pairs score 0.80 or more.

## loadtest

Drives a running stack (`make docker-up`, or the production Compose file) and prints results.

- `make loadtest-api CONCURRENCY=16 DURATION=2m`: registers a candidate and sends a mix of profile, matches, applications and dashboard reads; reports requests per second, p50/p95/p99 latency and status codes. Raise `RATE_LIMIT_USER`/`RATE_LIMIT_IP` on the stack first, or the run measures the rate limiter.
- `make loadtest-apply JOBS=1000`: starts a stub Greenhouse API on `:18090`, seeds that many matched Greenhouse jobs for a new candidate directly in Postgres, queues their applications and reports applications per hour. The worker must point at the stub: `GREENHOUSE_BOARDS_API_URL=http://host.docker.internal:18090/greenhouse`, `GREENHOUSE_BOARD_KEYS=loadtest:loadtest-key` and a high `ATS_RATE_LIMIT`.

Both modes read the stack's database and Redis settings from the environment (`DB_*`, `REDIS_*`, `JWT_SECRET`), as the app does. Watch memory alongside with `docker stats`. Results from 2026-11-16 are in `docs/RUNBOOK.md`.

