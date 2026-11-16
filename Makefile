APP := autoreach
CMD := ./cmd/autoreach
BIN := bin/$(APP)

DB_USER ?= autoreach
DB_PASSWORD ?= autoreach
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= autoreach
DB_SCHEMA ?= autoreach
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable&search_path=$(DB_SCHEMA),public
GOLANGCI_LINT_VERSION ?= v2.15.0
GOLANGCI_LINT ?= golangci-lint
MIGRATE ?= docker run --rm --network host -v "$(CURDIR)/migrations:/migrations:ro" migrate/migrate:v4.17.1

ROWS ?= 50000
PAIRS ?= pairs.jsonl
CONCURRENCY ?= 16
DURATION ?= 2m
JOBS ?= 1000

.PHONY: vector-bench calibrate build run-server run-worker dev tools fmt lint lint-fix vet test test-unit test-integration test-live test-all vuln secrets-scan loadtest-api loadtest-apply cover check \
	migrate-up migrate-down docker-build docker-up docker-down docker-logs dev-up dev-down test-up test-down

build:
	go build -trimpath -o $(BIN) $(CMD)

run-server:
	go run $(CMD) server

run-worker:
	go run $(CMD) worker

dev:
	air

tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

fmt:
	$(GOLANGCI_LINT) fmt
	$(GOLANGCI_LINT) run --fix ./...

lint:
	$(GOLANGCI_LINT) fmt --diff
	$(GOLANGCI_LINT) run ./...

lint-fix: fmt

vet:
	go vet ./...
	go vet -tags=integration ./...

test: test-unit

test-unit:
	go test ./...

test-integration:
	TEST_REQUIRE_SERVICES=1 go test -tags=integration -count=1 -p 1 ./tests/integration/...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...

secrets-scan:
	docker run --rm -v "$(CURDIR):/repo" zricethezav/gitleaks:v8.28.0 git /repo --config /repo/.gitleaks.toml --redact

# Load tests run against a running stack (make docker-up); see scripts/README.md.
loadtest-api:
	go run ./scripts/loadtest -mode api -concurrency $(CONCURRENCY) -duration $(DURATION)

loadtest-apply:
	go run ./scripts/loadtest -mode apply -jobs $(JOBS)

test-live:
	go test -tags=live -count=1 -v ./tests/live/...

test-all: test-unit test-integration

cover:
	go test -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

check: fmt vet lint test-unit build

migrate-up:
	$(MIGRATE) -path=/migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path=/migrations -database "$(DB_URL)" down 1

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f server

dev-up:
	docker compose -f docker-compose.dev.yml up -d

dev-down:
	docker compose -f docker-compose.dev.yml down

test-up:
	docker compose -f docker-compose.test.yml up -d --wait

test-down:
	docker compose -f docker-compose.test.yml down -v

vector-bench:
	go run ./scripts/vectorBench -rows $(ROWS)

calibrate:
	go run ./scripts/calibrate -pairs $(PAIRS)
