# Code conventions

Rules every change follows. CI enforces formatting and linting (`make lint`); the rest is checked in review and partly by tests.

## Rules that are easy to miss

- Every response uses one envelope: `{ success, statusCode, data, meta, error }`. JSON keys are camelCase.
- Malformed JSON is `400`; a payload that parses but fails validation is `422` with `error.details[]`.
- Error bodies carry a fixed public message per error code. Internal detail is logged, never returned.
- A candidate-scoped route returns `404`, not `403`, for another tenant's ID.
- Every route under `/api/v1` must be in `docs/swagger/openapi.json`; a unit test fails otherwise, and router tests validate real responses against the spec.
- Every worker task needs a `TaskConfig` and a registered handler; the worker refuses to start if they drift.
- The database is the source of truth for pipeline state. Queue tasks are idempotent, and a reconciler cron re-enqueues rows whose task was lost.
- Comments are rare: one line, only where the logic is not obvious from the names (an algorithm, a security reason, a protocol quirk).
- Models in `pkg/models` are database entities only. Each has `New<Model>()` (generates the ID, stamps timestamps and defaults), `UpdateTimestamp()` and `Validate()`. Projections, requests and responses live in `pkg/customtypes`.
- One validation per stage: `utils.BindAndValidate` validates the request from its tags; the request's `To<Model>Model()` receiver only maps fields and never validates; the service calls `UpdateTimestamp()` then the model's `Validate()` before saving.
- A handler only reads the request, calls one service method and writes the response or error. Create endpoints convert the request to a model and pass the model; update endpoints pass the request, and the service loads the existing entity and converts it there.
- Saves are `Upsert(ctx, model)` and return the model they were given. IDs and defaults are set in the constructor, so nothing is read back from the database after a write.
- Repositories hold no duplicated query or scan code: one private finder per table, and `db.FindOne` / `pgx.CollectRows` for row mapping.
- Services, handlers, repositories, utilities and integrations declare no constants and no stray types. Constants live in `config` behind getters (headers, vendor wire values, lock names, validation messages); non-entity structs live in `pkg/customtypes/<domain>`. Enum types stay in `pkg/models/helpers`, task and queue names in `worker/configs`, and SQL column lists beside the queries that use them.
- Error messages live in `config/errors.go` as `config.ErrXxx`; no other package calls `errors.New`.
- Independent operations run through `utils.RunConcurrently`; fire-and-forget work through `utils.RunInBackground`. No bare `go` outside `pkg/utils` and bootstrap.
- Formatting is mechanical: run `make fmt` before committing. CI runs `make lint`. Format-on-save is a personal editor choice and is not part of the repository.
