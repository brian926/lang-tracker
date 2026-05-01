# AGENTS.md

## What this repo is

Go backend for a language-learning activity tracker. Deploys as an AWS Lambda function. `cmd/local` wraps it in a plain HTTP server for local development.

## Two entrypoints

| Path | Purpose |
|---|---|
| `cmd/lambda/main.go` | Production — starts `lambda.Start(svc.Handler)` |
| `cmd/local/main.go` | Local dev — HTTP server on `:8080`, loads `.env`, mimics API Gateway |

## Key commands

```bash
# Run locally
go run ./cmd/local

# Build Lambda binary (Linux target required for AWS)
GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./cmd/lambda

# Run all tests
go test ./...

# Run a single package's tests
go test ./internal/service/...
go test ./internal/handler/...

# Vet
go vet ./...
```

On Windows, set env vars inline or use `.env` (loaded automatically by `cmd/local` via `godotenv`).

## Dependency injection

The handler, log service, and stats service all receive a `db.DynamoDBClient` interface — **not** a concrete `*dynamodb.Client`. Tests inject `mockDynamo` structs. The real client is wired up in `cmd/lambda/main.go` and `cmd/local/main.go` via `db.New(ctx)`.

Do not add a package-level `db.Client` global. The previous global was removed intentionally.

## DynamoDB

- Table name defaults to `"lang-tracker"`, overridable via `TABLE_NAME` env var.
- Partition key: `userId` (string). `logId` is the sort key (UUID generated at write time).
- `db.New(ctx)` returns `(*dynamodb.Client, error)` — callers must handle the error; it no longer panics.
- `db.QueryByUserId` paginates automatically using `LastEvaluatedKey`.
- Credentials come from `.env` (ignored by git). The `.env` file contains real AWS credentials — **do not commit it**.

## Request shape

All requests go to `POST /api` with JSON body. Routing is done by the `action` field in the handler (`internal/handler/handler.go`):

- `action: "log"` — write a log entry (requires `userId`, `language`, `activityType`, `minutes` (1–1440), `date`)
- `action: "stats"` — get stats for a user+language (requires `userId`, `language`)

Date formats accepted: `"2006-01-02"` and `"01/02/2006"` (both parsed in `stats_service.go`).

All validation errors return HTTP 400 with `{"error": "<message>"}`. Internal errors return HTTP 500 with `{"error": "internal server error"}` — full errors are logged server-side only.

## Handler wiring

`handler.Services` holds `*service.LogService` and `*service.StatsService`. The entry point creates them:

```go
svc := &handler.Services{
    Log:   &service.LogService{DB: client, TableName: table},
    Stats: &service.StatsService{DB: client, TableName: table},
}
lambda.Start(svc.Handler)
```

## Test calls

`test_calls.ps1` has working PowerShell `curl` examples for both actions against the local server.

## Roadmap / known issues (from README)

- Handler routing by URL path is planned (remove `action` param)
- Stats are currently per-language only; per-activity breakdown is planned
- AWS deployment not yet done
- No authentication — any caller can read/write any `userId`

## .gitignore notes

`go.mod` and `go.sum` are listed in `.gitignore` — this is unusual. They **are** present in the repo but will show as untracked after changes. Restore or force-add them if needed: `git add -f go.mod go.sum`.

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
at `specs/001-lang-tracker-api/plan.md`.
<!-- SPECKIT END -->
