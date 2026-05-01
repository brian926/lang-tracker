# Copilot Instructions for lang-tracker

## Project Overview

A Go backend API for tracking language-learning activity. The service stores activity logs in AWS DynamoDB and computes per-language usage statistics. It deploys as an AWS Lambda function but also includes a local HTTP server for development.

---

## Build, Test & Development Commands

### Running the local server
```bash
# Starts HTTP server on :8080 (default port, overridable via PORT env var)
# Automatically loads .env file via godotenv
go run ./cmd/local
```

### Running tests
```bash
# Run all tests (no AWS connection required — all tests use mock DynamoDB)
go test ./...

# Run a specific package
go test ./internal/handler/...
go test ./internal/service/...
go test ./internal/db/...
```

### Building & validation
```bash
# Build all packages
go build ./...

# Vet for code issues
go vet ./...

# Build Lambda binary for deployment (requires GOOS=linux)
GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./cmd/lambda
```

### Sample API calls
See `test_calls.ps1` for PowerShell curl examples. Local server listens on `POST /api`.

---

## Architecture

### Entry Points
| Path | Purpose |
|---|---|
| `cmd/lambda/main.go` | Production — AWS Lambda handler |
| `cmd/local/main.go` | Development — HTTP server wrapping the Lambda handler |

### Package Structure
- **`internal/db/`** — DynamoDB client interface (`DynamoDBClient`), wrappers (`CreateItem`, `QueryByUserId`), and initialization
- **`internal/handler/`** — Lambda/HTTP handler, JSON routing by `action` field, request validation, response formatting
- **`internal/models/`** — Shared request/response structs
- **`internal/service/`** — Business logic (`LogService`, `StatsService`), date parsing, stats aggregation
- **`cmd/`** — Entrypoints (both wire up services and the DynamoDB client)

### Key Design Pattern: Dependency Injection
All services receive a `db.DynamoDBClient` interface rather than a concrete `*dynamodb.Client`. This enables:
- Full testability with mock DynamoDB implementations
- No package-level globals or singletons
- Explicit dependency wiring in both entrypoints

**Do not add a package-level `db.Client` global** — the previous global was intentionally removed.

### DynamoDB Details
- **Table name:** `"lang-tracker"` (overridable via `TABLE_NAME` env var)
- **Partition key:** `userId` (string)
- **Sort key:** `logId` (UUID, generated at write time)
- **Query pattern:** `QueryByUserId` paginates automatically following DynamoDB's `LastEvaluatedKey`

---

## Request Routing

All requests are `POST /api` with JSON body. Routing is determined by the `action` field:

### `action: "log"` — Log an activity
Records a language-learning session.

**Required fields:** `userId`, `language`, `activityType`, `minutes` (1–1440), `date`

**Date formats accepted:** `"2006-01-02"` or `"01/02/2006"` (parsed in `stats_service.go`)

**Validation:**
- All string fields must be non-empty after trimming whitespace
- `minutes` must be positive and ≤ 1440
- `date` is required (both formats supported)
- Errors are collected and joined by `"; "` for multi-error responses

### `action: "stats"` — Get aggregated stats
Returns per-language usage stats for a user.

**Required fields:** `userId`, `language`

**Response includes:**
- `totalHours` — All-time hours for the language
- `today`, `thisWeek`, `thisMonth` — Hours in each period
- `percentages` — Activity type breakdown (e.g., Reading: 60%, Watching: 30%)

### Error Handling
- **400** — Validation failure or unknown action; returns `{"error": "<message>"}`
- **500** — Unexpected error (details logged server-side only); returns `{"error": "internal server error"}`

---

## Configuration

| Environment Variable | Default | Behavior |
|---|---|---|
| `AWS_ACCESS_KEY_ID` | — | AWS credentials (required) |
| `AWS_SECRET_ACCESS_KEY` | — | AWS credentials (required) |
| `AWS_REGION` | — | AWS region (required) |
| `TABLE_NAME` | `lang-tracker` | DynamoDB table name |
| `PORT` | `8080` | Local dev server port only |

The local server loads `.env` automatically via `godotenv`. In Lambda, set these as function environment variables. **Never commit `.env`** — it contains real AWS credentials.

---

## Key Conventions

### Error handling
- Use `fmt.Errorf("context: %w", err)` for wrapping errors
- Handler catches errors from services and returns HTTP 500 with `"internal server error"` (full error logged server-side)
- Validation errors are accumulated in `[]string` slices and joined with `"; "`

### Testing
- All tests use mock DynamoDB implementations — no AWS connection needed
- Service tests inject a mock client via the `DynamoDBClient` interface
- Handler tests validate routing, validation, and response formatting

### Response formatting
- All responses use `map[string]interface{}` marshaled to JSON
- Use `jsonResponse(status, body)` helper to set `Content-Type: application/json` and marshal
- Use `errResponse(status, msg)` for error bodies

### Dependencies
- `github.com/aws/aws-sdk-go-v2` — AWS SDK (DynamoDB)
- `github.com/aws/aws-lambda-go` — Lambda handler interface
- `github.com/google/uuid` — UUID generation
- `github.com/joho/godotenv` — .env file loading (dev only)

---

## Roadmap / Known Limitations

- **Routing:** Currently uses `action` field; planned to switch to HTTP method + URL path routing
- **Stats granularity:** Activity type breakdown (`today`, `thisWeek`, `thisMonth` per activity) is planned
- **Authentication:** None currently — any caller can read/write any `userId`
- **Deployment:** Full AWS infrastructure (API Gateway, IAM, IaC) not yet set up

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->
