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
