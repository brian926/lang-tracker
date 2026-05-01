---

description: "Task list for Lang-Tracker API implementation"
---

# Tasks: Lang-Tracker API

**Input**: Design documents from `specs/001-lang-tracker-api/`
**Prerequisites**: plan.md ✅, data-model.md ✅, contracts/api.md ✅, research.md ✅, quickstart.md ✅

**Tests**: Requested — the user requirement states "Code should be well tested and verified."
Tests are included for all services, the handler, and the db package.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm project structure and tooling baseline are correct.

- [X] T001 Verify `go test ./...` passes clean and `go vet ./...` produces no output
- [X] T002 [P] Verify `internal/db/dynamo.go` exposes `DynamoDBClient` interface with `PutItem` and `Query` methods
- [X] T003 [P] Verify `internal/models/models.go` contains `Request`, `LogItem`, and `StatsResponse` structs

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Fill the test coverage gaps identified in the Constitution Check before user-story
work is considered complete. These are the only remaining implementation gaps.

**⚠️ CRITICAL**: All story work depends on these test files existing — Constitution Principle IV
requires `go test ./...` to pass with full coverage of all service and handler paths.

- [X] T004 Create `internal/db/dynamo_test.go` with unit tests for `QueryByUserId` pagination logic using a `mockDynamo` that simulates multi-page results via `LastEvaluatedKey`
- [X] T005 [P] Create `internal/service/log_service_test.go` with tests: successful `LogActivity` write, `PutItem` error propagation, UUID is generated (logId non-empty in marshaled item)
- [X] T006 Confirm `go test ./...` passes with all new test files (T004, T005 complete)

**Checkpoint**: All packages have test coverage — `go test ./...` fully green.

---

## Phase 3: User Story 1 — Log Activity (Priority: P1) 🎯 MVP

**Goal**: A frontend can POST a JSON payload to log a language-learning session; the entry is
persisted to DynamoDB and a success message is returned.

**Independent Test**:
```powershell
curl.exe -X POST http://localhost:8080/api `
  -H "Content-Type: application/json" `
  -d '{"action":"log","userId":"test1","language":"French","activityType":"Reading","minutes":30,"date":"2026-05-01"}'
# Expected: {"message":"Log saved"}
```

### Tests for User Story 1 ⚠️ Write FIRST — verify they fail before implementation

- [X] T007 [P] [US1] Add handler test cases in `internal/handler/handler_test.go` for `action:"log"` happy path (HTTP 200, `{"message":"Log saved"}`)
- [X] T008 [P] [US1] Add handler test cases for all `log` validation failures: missing `userId`, `language`, `activityType`, `minutes ≤ 0`, `minutes > 1440`, missing `date` — each MUST return HTTP 400 with correct error string
- [X] T009 [P] [US1] Add handler test case for malformed JSON body → HTTP 400 `{"error":"invalid JSON body"}`
- [X] T010 [P] [US1] Add handler test case for `LogActivity` service error → HTTP 500 `{"error":"internal server error"}`

### Implementation for User Story 1

- [X] T011 [US1] Verify `internal/service/log_service.go:LogActivity` marshals all fields correctly and calls `db.CreateItem` — cross-check against `data-model.md` LogItem fields
- [X] T012 [US1] Verify `internal/handler/handler.go:validateLogRequest` enforces all rules in `data-model.md` validation table; add any missing checks
- [X] T013 [US1] Run `go test ./internal/handler/... ./internal/service/...` — all T007–T010 tests MUST pass

**Checkpoint**: `POST /api` with `action:"log"` is fully functional and all validation paths
are tested. User Story 1 is independently demonstrable.

---

## Phase 4: User Story 2 — Retrieve Stats (Priority: P2)

**Goal**: A frontend can POST a stats request and receive aggregated hours and activity
percentages for a user+language pair.

**Independent Test**:
```powershell
# First log some data (US1 must be complete)
curl.exe -X POST http://localhost:8080/api `
  -H "Content-Type: application/json" `
  -d '{"action":"stats","userId":"test1","language":"French"}'
# Expected: JSON with totalHours, today, thisWeek, thisMonth, percentages fields
```

### Tests for User Story 2 ⚠️ Write FIRST — verify they fail before implementation

- [X] T014 [P] [US2] Add handler test cases in `internal/handler/handler_test.go` for `action:"stats"` happy path (HTTP 200, response contains all `StatsResponse` fields)
- [X] T015 [P] [US2] Add handler test cases for `stats` validation failures: missing `userId`, missing `language` — each MUST return HTTP 400
- [X] T016 [P] [US2] Add handler test case for `GetStats` service error → HTTP 500 `{"error":"internal server error"}`
- [X] T017 [P] [US2] Add handler test case for unknown `action` value → HTTP 400 `{"error":"unknown action: must be \"log\" or \"stats\""}`

### Implementation for User Story 2

- [X] T018 [US2] Verify `internal/service/stats_service.go:GetStats` correctly filters by language, computes `TotalHours`, `Today`, `ThisWeek`, `ThisMonth`, `Percentages` — existing tests in `stats_service_test.go` already cover this
- [X] T019 [US2] Verify `internal/handler/handler.go:validateStatsRequest` enforces `userId` and `language` required rules
- [X] T020 [US2] Run `go test ./internal/handler/... ./internal/service/...` — all T014–T017 tests MUST pass

**Checkpoint**: `POST /api` with `action:"stats"` is fully functional. Both User Stories 1 and 2
are independently working end-to-end.

---

## Phase 5: User Story 3 — End-to-End Integration (Priority: P3)

**Goal**: The full round-trip (log → retrieve stats) works correctly via the local server
and all existing test vectors pass, confirming the integration between handler, services, and db.

**Independent Test**: Follow `specs/001-lang-tracker-api/quickstart.md` validation checklist —
all 5 items MUST pass.

### Tests for User Story 3 ⚠️ Write FIRST — verify they fail before implementation

- [X] T021 [P] [US3] Add `internal/handler/handler_test.go` test for `ToAPIGatewayRequest` round-trip: marshal a `Request`, wrap via `ToAPIGatewayRequest`, pass to handler, assert correct routing
- [X] T022 [P] [US3] Add `internal/db/dynamo_test.go` test for `CreateItem`: confirm `PutItem` is called with the correct `TableName` and item content

### Implementation for User Story 3

- [X] T023 [US3] Run full `go test ./...` — every package MUST report `ok` or `[no test files]`
- [X] T024 [US3] Run `go vet ./...` — MUST produce no output
- [X] T025 [US3] Start `go run ./cmd/local` and execute all smoke test calls from `quickstart.md` — all MUST succeed against a live DynamoDB table

**Checkpoint**: All user stories complete. Full test suite green. Local server validated end-to-end.

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, build artifact, and code hygiene.

- [X] T026 [P] Build Lambda binary: `GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./cmd/lambda` — MUST succeed with zero errors
- [X] T027 [P] Update `test_calls.ps1` to include working examples for both `log` and `stats` actions matching the request shapes in `contracts/api.md`
- [X] T028 Review `README.md` API Reference section against `contracts/api.md` — update any discrepancies
- [X] T029 Force-add module files if changed: `git add -f go.mod go.sum` (see `.gitignore` note in `AGENTS.md`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 verification
- **User Stories (Phase 3, 4, 5)**: Depend on Phase 2 completion
  - US1 (Phase 3) has no dependency on US2/US3
  - US2 (Phase 4) has no dependency on US1/US3
  - US3 (Phase 5) depends on US1 + US2 being complete (tests the integration)
- **Polish (Phase N)**: Depends on all user story phases

### User Story Dependencies

- **US1 (P1)**: Independent after Phase 2
- **US2 (P2)**: Independent after Phase 2
- **US3 (P3)**: Depends on US1 + US2 completion

### Within Each User Story

- Tests MUST be written and FAIL before implementation tasks begin
- Handler tests (T007–T010, T014–T017, T021–T022) can run in parallel [P]
- Implementation verification tasks follow tests

### Parallel Opportunities

- T002 and T003 can run in parallel (Phase 1)
- T004 and T005 can run in parallel (Phase 2)
- All T007–T010 can run in parallel (US1 tests)
- All T014–T017 can run in parallel (US2 tests)
- T021 and T022 can run in parallel (US3 tests)
- T026 and T027 can run in parallel (Polish)

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together (write and verify they fail first):
Task: "Add handler test for log happy path in internal/handler/handler_test.go"       # T007
Task: "Add handler tests for log validation failures in internal/handler/handler_test.go"  # T008
Task: "Add handler test for malformed JSON in internal/handler/handler_test.go"        # T009
Task: "Add handler test for LogActivity error in internal/handler/handler_test.go"     # T010
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup verification
2. Complete Phase 2: Foundational tests (T004–T006)
3. Complete Phase 3: User Story 1 — log endpoint fully tested
4. **STOP and VALIDATE**: `go test ./...` green, smoke test log action ✅
5. Demo/deploy if ready

### Incremental Delivery

1. Phase 1 + 2 → Foundation verified
2. Phase 3 (US1) → Log endpoint tested → Demo
3. Phase 4 (US2) → Stats endpoint tested → Demo
4. Phase 5 (US3) → End-to-end validated → Lambda binary built
5. Phase N → Polish complete

### Parallel Team Strategy

With 2 developers after Phase 2:
- Developer A: US1 (Phase 3) — log tests + verification
- Developer B: US2 (Phase 4) — stats handler tests + verification
- Both merge → US3 (Phase 5) together

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Tests MUST fail before implementation — red-green-refactor (Constitution Principle IV)
- `go test ./...` and `go vet ./...` MUST pass clean at every checkpoint
- The core implementation already exists; most tasks are test-coverage completions
- Commit after each phase or logical group
- Do NOT introduce new abstractions, GSIs, auth, or path-routing — out of scope (Constitution Principle V)
