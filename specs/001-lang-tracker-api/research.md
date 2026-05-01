# Research: Lang-Tracker API (001-lang-tracker-api)

All unknowns resolved from existing codebase analysis. No NEEDS CLARIFICATION items remain.

---

## Language & Runtime

**Decision**: Go 1.24.3  
**Rationale**: Already in use; `go.mod` declares `go 1.24.3`. Matches `aws-lambda-go` and
`aws-sdk-go-v2` SDK versions already pinned.  
**Alternatives considered**: None — language already chosen and deployed.

---

## Primary Dependencies

**Decision**: `aws-lambda-go v1.49.0`, `aws-sdk-go-v2 v1.38.2` (full suite), `google/uuid v1.6.0`,
`joho/godotenv v1.5.1`  
**Rationale**: Already present in `go.mod`. AWS SDK v2 is the current generation (v1 deprecated);
godotenv handles local `.env` loading without changing Lambda behaviour.  
**Alternatives considered**: None — already locked.

---

## Storage

**Decision**: AWS DynamoDB via `aws-sdk-go-v2/service/dynamodb`  
**Rationale**: Serverless, zero-ops, scales to zero, no connection pools needed in Lambda. The
partition-key-per-user model fits the access pattern (write single item; query all by userId).  
**Alternatives considered**:
- RDS/Aurora Serverless — rejected: connection management overhead in Lambda cold starts.
- S3 + Athena — rejected: poor latency for real-time stat queries.

---

## DynamoDB Access Pattern Analysis

**Decision**: Single-table design with `userId` (PK) + `logId` (SK, UUID)  
**Rationale**: All queries are by userId. Stats are computed in-application by scanning all
items for a user then filtering by language. This avoids GSIs for the current scope.  
**Known limitation**: Query retrieves ALL items for a user regardless of language; stats filter
happens in Go. Acceptable at small scale; a `language` GSI would optimize at high item counts.  
**Alternatives considered**:
- GSI on `(userId, language)` — deferred per YAGNI principle; add if profiling shows need.
- Sort key as `language#date` composite — rejected: complicates UUID-based idempotency.

---

## Testing Strategy

**Decision**: `go test ./...` with mock-injected `db.DynamoDBClient`  
**Rationale**: Interface-based DI already in place. Tests use `mockDynamo` structs in
`_test` packages — no live DynamoDB calls. Handler tests use the same pattern.
Pattern is established in `internal/service/stats_service_test.go`.  
**Coverage gaps identified** (to address in tasks):
- `internal/handler/handler.go` has a test file in `internal/handler/` — confirmed by `go test` passing, but coverage of all validation branches needs verification.
- `internal/service/log_service.go` has no test file yet — needs `log_service_test.go`.
- `internal/db/dynamo.go` has no tests — `QueryByUserId` pagination logic needs unit coverage.
**Alternatives considered**: Integration tests against DynamoDB Local — deferred, out of scope for v1.

---

## Architecture: Lambda vs Local Dev

**Decision**: Dual entrypoint pattern — `cmd/lambda` (production) and `cmd/local` (dev)  
**Rationale**: Lambda handler (`events.APIGatewayProxyRequest`) is wrapped by a thin HTTP
adapter in `cmd/local/main.go` via `handler.ToAPIGatewayRequest`. Business logic is
completely shared. `godotenv` loaded only in local mode.  
**Alternatives considered**: Single binary with build tags — rejected: adds complexity with no benefit.

---

## Error Handling

**Decision**: HTTP 400 for validation errors, HTTP 500 for internal errors; full error logged
server-side only  
**Rationale**: Prevents information leakage. Pattern already implemented in `handler.go`.  
**Alternatives considered**: Problem+JSON (RFC 7807) — deferred per YAGNI.

---

## Date Parsing

**Decision**: Accept both `"2006-01-02"` (ISO) and `"01/02/2006"` (US slash) formats  
**Rationale**: Already implemented in `stats_service.go`. Both covered by existing test
`TestGetStats_AcceptsSlashDateFormat`.  
**Note**: Date is stored as-is (string) in DynamoDB; parsing happens at read time in stats.

---

## Roadmap Items (Out of Scope — Constitution V)

The following are acknowledged but MUST NOT be implemented until spec-approved:

| Item | Status |
|------|--------|
| URL-path routing (remove `action` param) | Planned |
| Per-activity stats breakdown | Planned |
| AWS deployment / CDK / Terraform | Not started |
| Authentication / user identity | Not started |
| Language GSI on DynamoDB | Deferred (performance) |
