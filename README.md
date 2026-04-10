# Lang-Tracker

A Go backend API for tracking language-learning activity. Stores logs in AWS DynamoDB and computes per-language usage stats. Deploys as an AWS Lambda function behind API Gateway, with a local HTTP server for development.

## Table of Contents

- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Roadmap](#roadmap)

---

## Architecture

```
cmd/
  lambda/   → production entrypoint (AWS Lambda)
  local/    → development entrypoint (HTTP server on :8080)
internal/
  db/       → DynamoDB client interface + query helpers
  handler/  → request routing, validation, response formatting
  models/   → shared request/response structs
  service/  → business logic (LogService, StatsService)
```

All services depend on a `db.DynamoDBClient` interface rather than a concrete AWS client, making them fully testable with mocks.

**DynamoDB table:**
- Table name: `lang-tracker` (overridable via `TABLE_NAME` env var)
- Partition key: `userId` (string)
- Sort key: `logId` (UUID, generated at write time)

---

## Getting Started

**Prerequisites:** Go 1.24+, AWS credentials with DynamoDB access.

1. Clone the repo and install dependencies:
   ```bash
   git clone https://github.com/brian926/lang-tracker.git
   cd lang-tracker
   go mod download
   ```

2. Create a `.env` file in the project root (never commit this):
   ```
   AWS_ACCESS_KEY_ID=your_key_id
   AWS_SECRET_ACCESS_KEY=your_secret
   AWS_REGION=us-east-1
   ```

3. Run the local server:
   ```bash
   go run ./cmd/local
   # Listening at http://localhost:8080/api
   ```

---

## Configuration

| Environment variable | Default | Description |
|---|---|---|
| `AWS_ACCESS_KEY_ID` | — | AWS credentials (required) |
| `AWS_SECRET_ACCESS_KEY` | — | AWS credentials (required) |
| `AWS_REGION` | — | AWS region (required) |
| `TABLE_NAME` | `lang-tracker` | DynamoDB table name |
| `PORT` | `8080` | Local dev server port |

The local server loads `.env` automatically via `godotenv`. In Lambda, set these as environment variables in the function configuration.

---

## API Reference

All requests are `POST /api` with a JSON body. The `action` field determines what the request does.

### Log an activity

Records a language-learning session.

**Request:**
```json
{
  "action": "log",
  "userId": "user123",
  "language": "French",
  "activityType": "Reading",
  "minutes": 45,
  "date": "2026-04-10"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `action` | string | yes | must be `"log"` |
| `userId` | string | yes | identifies the user |
| `language` | string | yes | e.g. `"French"`, `"Spanish"` |
| `activityType` | string | yes | e.g. `"Reading"`, `"Watching"`, `"Speaking"` |
| `minutes` | integer | yes | 1–1440 |
| `date` | string | yes | `YYYY-MM-DD` or `MM/DD/YYYY` |

**Response `200`:**
```json
{ "message": "Log saved" }
```

---

### Get stats

Returns aggregated usage stats for a user and language.

**Request:**
```json
{
  "action": "stats",
  "userId": "user123",
  "language": "French"
}
```

**Response `200`:**
```json
{
  "totalHours": 12.5,
  "today": 0.75,
  "thisWeek": 3.0,
  "thisMonth": 10.25,
  "percentages": {
    "Reading": 60.0,
    "Watching": 30.0,
    "Speaking": 10.0
  }
}
```

| Field | Description |
|---|---|
| `totalHours` | All-time hours logged for this language |
| `today` | Hours logged today |
| `thisWeek` | Hours logged in the current ISO week |
| `thisMonth` | Hours logged in the current calendar month |
| `percentages` | Each activity type's share of total logged time (sums to 100%) |

---

### Error responses

| Status | Body | Cause |
|---|---|---|
| `400` | `{"error": "<message>"}` | Validation failure or unknown action |
| `500` | `{"error": "internal server error"}` | Unexpected server error (details logged server-side) |

---

## Development

```bash
# Run locally (loads .env, serves on :8080)
go run ./cmd/local

# Build
go build ./...

# Vet
go vet ./...
```

Sample requests are in `test_calls.ps1` (PowerShell):
```powershell
# Log an activity
curl -X POST `
    -H "Content-Type: application/json" `
    -d '{"userId":"1","language":"French","activityType":"Watching","action":"log","minutes":60,"date":"2026-04-10"}' `
    http://localhost:8080/api

# Get stats
curl -X POST `
    -H "Content-Type: application/json" `
    -d '{"userId":"1","language":"French","action":"stats"}' `
    http://localhost:8080/api
```

---

## Testing

Tests use mock DynamoDB clients — no AWS connection required.

```bash
# Run all tests
go test ./...

# Run a specific package
go test ./internal/service/...
go test ./internal/handler/...
```

Test coverage:
- `internal/handler` — routing, validation, all actions, error handling, response headers
- `internal/service` — stats aggregation, date parsing, language filtering, percentages, pagination, error propagation

---

## Deployment

Build a Linux binary for Lambda:

```bash
GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./cmd/lambda
```

Then zip and deploy to AWS Lambda:
```bash
zip lambda.zip bin/bootstrap
aws lambda update-function-code --function-name lang-tracker --zip-file fileb://lambda.zip
```

Set the following environment variables on the Lambda function:
- `AWS_REGION`
- `TABLE_NAME` (if not using the default `lang-tracker`)

> **Note:** Full AWS deployment (API Gateway wiring, IAM roles, infrastructure-as-code) is not yet set up. See [Roadmap](#roadmap).

---

## Roadmap

- [ ] Route by HTTP method + URL path — remove the `action` field from request bodies
- [ ] Per-activity time breakdown in stats (`today`, `thisWeek`, `thisMonth` per activity)
- [ ] Full AWS deployment (API Gateway, IAM, infrastructure-as-code)
- [ ] Authentication — currently any caller can read/write any `userId`
