# Implementation Plan: Lang-Tracker API

**Branch**: `001-lang-tracker-api` | **Date**: 2026-05-01 | **Spec**: `specs/001-lang-tracker-api/`
**Input**: User requirement — Go backend API for language-learning activity tracking with DynamoDB storage and AWS Lambda deployment.

## Summary

Build a well-tested Go backend API that accepts JSON requests from frontends to log
language-learning activity sessions and retrieve per-user, per-language usage statistics.
The service deploys as an AWS Lambda function, stores data in DynamoDB (`lang-tracker` table),
and exposes a single `POST /api` endpoint routed by an `action` field. The codebase uses
interface-based dependency injection throughout, making all services fully unit-testable with
mocks and zero live AWS calls in tests.

**Current state**: Core implementation exists and tests pass. Gaps are: missing test files for
`log_service` and `db` package, and test coverage of all handler validation branches.

## Technical Context

**Language/Version**: Go 1.24.3  
**Primary Dependencies**: `aws-lambda-go v1.49.0`, `aws-sdk-go-v2 v1.38.2`, `google/uuid v1.6.0`,
`joho/godotenv v1.5.1`  
**Storage**: AWS DynamoDB — table `lang-tracker`, PK `userId` (S), SK `logId` (S, UUID)  
**Testing**: `go test ./...` with `mockDynamo` structs; no live AWS calls  
**Target Platform**: AWS Lambda (`provided.al2023`); local HTTP server on `:8080` for dev  
**Project Type**: Web service (JSON API)  
**Performance Goals**: Lambda cold-start < 1s; p95 response < 500ms for stats (application-side
aggregation over typical user item counts)  
**Constraints**: No package-level globals; DI via `db.DynamoDBClient` interface; `.env` MUST NOT
be committed; `go.mod`/`go.sum` tracked via `git add -f`  
**Scale/Scope**: Single-user hobby/personal use to small team; no auth in v1

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate | Status |
|-----------|------|--------|
| I. API-First Design | All logic behind `POST /api`, routed by `action`, correct error shapes | ✅ PASS |
| II. DynamoDB as Sole Data Store | Only DynamoDB used; `TABLE_NAME` env var; UUID sort key; paginated query | ✅ PASS |
| III. DI over Globals | `db.DynamoDBClient` interface injected; no package-level globals | ✅ PASS |
| IV. Test-First | Tests exist for stats service and handler; gaps in log service + db package need addressing | ⚠️ PARTIAL — gaps must be closed |
| V. Simplicity | No auth, no path routing, no GSIs introduced; YAGNI maintained | ✅ PASS |

**Post-Phase 1 re-check**: Principle IV gap (log service + db tests) is a task requirement,
not a design violation. No complexity tracking entries required.

## Project Structure

### Documentation (this feature)

```text
specs/001-lang-tracker-api/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── api.md           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/
├── lambda/
│   └── main.go          # Production Lambda entrypoint
└── local/
    └── main.go          # Local HTTP dev server (:8080)

internal/
├── db/
│   ├── init.go          # db.New(ctx) — creates real DynamoDB client
│   └── dynamo.go        # DynamoDBClient interface + CreateItem + QueryByUserId
├── handler/
│   └── handler.go       # Request routing, validation, response formatting
├── models/
│   └── models.go        # Request, LogItem, StatsResponse structs
└── service/
    ├── log_service.go       # LogService — writes to DynamoDB
    ├── log_service_test.go  # [MISSING — must be added]
    ├── stats_service.go     # StatsService — queries + aggregates
    └── stats_service_test.go

bin/
└── bootstrap            # Lambda binary (GOOS=linux GOARCH=amd64 build output)
```

**Structure Decision**: Single Go module (`lang-tracker`). Existing structure is correct
and complete; no new packages required for v1.

## Complexity Tracking

> No constitution violations requiring justification.
