# Security checklist

Status of each control in the implementation plan (section 4.7) and the M10 hardening work, with where it is enforced and what proves it. Reviewed 2026-11-16.

## Identity and access

| Control | Status | Enforced in | Evidence |
|---|---|---|---|
| Passwords hashed with argon2id | done | `pkg/utils/auth.go` | unit tests |
| Short-lived access tokens, rotating hashed refresh tokens | done | `pkg/services/authManagement` | `tests/integration/auth_test.go` (rotation, reuse, `alg: none` token rejected); `TestAccessTokenIsRejectedWhenExpiredTamperedOrForeign` |
| Every candidate-scoped route checks the token's candidate; other candidates' data answers 404 | done | `middlewares/authorize.go` | `TestEveryCandidateRouteHidesOtherCandidatesData` walks every registered route, so new routes are covered automatically |
| Every non-public route requires authentication | done | route groups | `TestEveryNonPublicRouteRequiresAuthentication` |
| Admin routes reject candidates | done | `config/authorizationPolicy.go` grants | `TestEveryAdminRouteRejectsCandidates` |
| API keys shown once, stored as SHA-256 with a lookup prefix, constant-time compare | done | `pkg/services/authManagement/verifyAPIKey.go` | integration tests |
| Webhook signatures and shared secrets compared in constant time | done | `pkg/utils/crypto.go`, `middlewares/authenticate.go` | `TestContactWithoutAnEmailIsCompletedLaterByTheSignalHireCallback` (wrong and missing token rejected) |
| Signed tracking and unsubscribe links | done | `pkg/services/outreachManagement` | `TestTrackingRecordsOpensAndOnlyFollowsSignedLinks`, `TestUnsubscribeLinkStopsTheSenderAndOnlyThePostChangesAnything` |

## Data protection

| Control | Status | Enforced in | Evidence |
|---|---|---|---|
| Mailbox credentials encrypted with AES-256-GCM, random nonce, record ID as AAD, versioned keys for rotation | done | `pkg/utils/credentialCipher.go` | `tests/unit/credentialCipher_test.go` (rotation, tampering, wrong AAD) |
| Production refuses to start without its secrets | done | `config/envConfig.go`, `config/channelAccountPolicy.go`, `docker-compose.yml` (`:?` required variables) | soak run: missing `CHANNEL_CREDENTIAL_KEYS` stopped startup |
| Contact data deleted on request, and the address suppressed so it cannot return | done | `GET /api/v1/admin/contacts?email=`, `DELETE /api/v1/admin/contacts/{contactId}` | `TestAdminFindsAndErasesAContact` |
| Contact retention window (configurable, off by default; anonymizes and keeps history) | done | daily `contactRetentionJob`, `CONTACT_RETENTION_PERIOD` | `TestStaleContactsAreAnonymizedButTheirHistoryIsKept` |
| Suppression honoured before planning and before every send | done | `pkg/services/outreachManagement` | outreach tests; global and per-candidate entries |
| No secrets or PII in logs | done | log calls carry IDs and counts only (erasure logs a count, never the address) | review of every `Log*` call in M10 |
| Uploaded files kept out of images | done | `.dockerignore` excludes `data/` and every `.env*` except `.env.example` | — |

## Abuse and exposure

| Control | Status | Enforced in | Evidence |
|---|---|---|---|
| Rate limits per IP, user, auth endpoint and outbound site (GCRA in Redis) | done | `middlewares/rateLimit.go`, `config/rateLimitPolicy.go` | `tests/unit/rateLimit_test.go`, `tests/integration/rateLimit_test.go` |
| Request body size limit | done | `middlewares.BodyLimit` | `TestBindReturnsPayloadTooLargeWhenTheBodyExceedsTheLimit` |
| CORS allow-list empty by default | done | `CORS_ALLOWED_ORIGINS` | `TestCORSAllowsOnlyListedOrigins` |
| `/metrics` and Swagger behind the internal secret in production | done | `routes/init.go`, worker ops server | — |
| Outbound fetches of career pages cannot reach private networks (checked at connect time, so redirects and DNS tricks fail too) and obey robots.txt | done | `pkg/utils/webFetcher.go`, `pkg/utils/robots.go` | career page tests, robots unit test |
| Ingested job URLs must be public https on the ATS's own host | done | `pkg/validations/jobValidation.go` | ingest validation tests |
| No CAPTCHA bypass or LinkedIn automation | done (by design) | assisted apply and assisted LinkedIn | — |

## Supply chain and runtime

| Control | Status | Enforced in | Evidence |
|---|---|---|---|
| Known vulnerabilities in dependencies and the standard library | 0 found | `govulncheck` in CI (`security` job) | M10 fixed 14 standard-library issues (Go 1.26.6 to 1.26.11) and upgraded quic-go, x/net and x/crypto |
| Secrets in git history | 0 found | `gitleaks` in CI with `.gitleaks.toml` (two test fixtures allow-listed by exact value) | full history scanned |
| Containers run as a non-root user, read-only root filesystem, all capabilities dropped, `no-new-privileges`, process and memory limits | done | `Dockerfile`, `docker-compose.yml` | soak run on the hardened stack |
| Postgres and Redis reachable only on loopback; Redis password; Redis never evicts queue keys | done | `docker-compose.yml` | — |
| Base images pinned (`golang:1.26.11-alpine`, `alpine:3.22`, `pgvector/pgvector:0.8.0-pg16`, `redis:7.2-alpine`, `migrate/migrate:v4.17.1`) | done | `Dockerfile`, `docker-compose.yml` | CI `image` job builds the image and validates Compose |

## Open items

- TLS is terminated by a reverse proxy in front of the stack; it is not part of this repository.
- Enrichment provider clients are tested against stubs only and have not been run against the live APIs.
- Adzuna's response fields are not yet checked against a live response (needs keys).
