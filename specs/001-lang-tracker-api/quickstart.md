# Quickstart: Lang-Tracker API

Get the local development server running and verified in under 5 minutes.

---

## Prerequisites

- Go 1.24+ installed (`go version`)
- AWS account with a DynamoDB table named `lang-tracker`
  - Partition key: `userId` (String)
  - Sort key: `logId` (String)
- AWS credentials with DynamoDB read/write access

---

## 1. Configure environment

Copy the example and fill in your AWS credentials:

```bash
# Create .env from scratch (it is gitignored — do NOT commit it)
cat > .env <<'EOF'
AWS_ACCESS_KEY_ID=YOUR_KEY
AWS_SECRET_ACCESS_KEY=YOUR_SECRET
AWS_REGION=us-east-1
TABLE_NAME=lang-tracker
EOF
```

On Windows (PowerShell):
```powershell
@"
AWS_ACCESS_KEY_ID=YOUR_KEY
AWS_SECRET_ACCESS_KEY=YOUR_SECRET
AWS_REGION=us-east-1
TABLE_NAME=lang-tracker
"@ | Set-Content .env
```

---

## 2. Start the local server

```bash
go run ./cmd/local
```

Expected output:
```
{"time":"...","level":"INFO","msg":"local server starting","addr":"http://localhost:8080/api"}
```

---

## 3. Smoke tests

### Log an activity

```powershell
curl.exe -X POST http://localhost:8080/api `
  -H "Content-Type: application/json" `
  -d '{"action":"log","userId":"user1","language":"French","activityType":"Reading","minutes":30,"date":"2026-05-01"}'
```

Expected response:
```json
{"message":"Log saved"}
```

### Retrieve stats

```powershell
curl.exe -X POST http://localhost:8080/api `
  -H "Content-Type: application/json" `
  -d '{"action":"stats","userId":"user1","language":"French"}'
```

Expected response (values depend on data):
```json
{
  "totalHours": 0.5,
  "today": 0.0,
  "thisWeek": 0.0,
  "thisMonth": 0.5,
  "percentages": {"Reading": 100}
}
```

> See `test_calls.ps1` at the repo root for more ready-to-run examples.

---

## 4. Run the test suite

```bash
go test ./...
```

All packages should report `ok` or `[no test files]`. No live AWS calls are made by the tests.

```bash
go vet ./...
```

Must produce no output (zero warnings).

---

## 5. Build the Lambda binary (optional)

```bash
GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./cmd/lambda
```

On Windows (PowerShell):
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o bin/bootstrap ./cmd/lambda
```

Upload `bin/bootstrap` to your Lambda function as a `provided.al2` or `provided.al2023` runtime.

---

## Validation checklist

- [ ] `go run ./cmd/local` starts without error
- [ ] `curl` log request returns `{"message":"Log saved"}`
- [ ] `curl` stats request returns a JSON object with `totalHours` field
- [ ] `go test ./...` — all packages pass
- [ ] `go vet ./...` — no output
