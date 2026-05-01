# Test calls for lang-tracker local server (http://localhost:8080/api)
# Start server first: go run ./cmd/local
# All requests: POST /api with JSON body

# --- action: log -------------------------------------------------------
# Log 60 minutes of Watching French (slash date format)
curl.exe -X POST `
    -H "Content-Type: application/json" `
    -d '{"action":"log","userId":"user1","language":"French","activityType":"Watching","minutes":60,"date":"10/15/2025"}' `
    http://localhost:8080/api

# Log 30 minutes of Reading French (ISO date format)
curl.exe -X POST `
    -H "Content-Type: application/json" `
    -d '{"action":"log","userId":"user1","language":"French","activityType":"Reading","minutes":30,"date":"2026-05-01"}' `
    http://localhost:8080/api

# Log 45 minutes of Speaking Spanish
curl.exe -X POST `
    -H "Content-Type: application/json" `
    -d '{"action":"log","userId":"user1","language":"Spanish","activityType":"Speaking","minutes":45,"date":"2026-05-01"}' `
    http://localhost:8080/api

# --- action: stats -----------------------------------------------------
# Get stats for user1 / French
curl.exe -X POST `
    -H "Content-Type: application/json" `
    -d '{"action":"stats","userId":"user1","language":"French"}' `
    http://localhost:8080/api

# Get stats for user1 / Spanish
curl.exe -X POST `
    -H "Content-Type: application/json" `
    -d '{"action":"stats","userId":"user1","language":"Spanish"}' `
    http://localhost:8080/api

# --- validation errors (should return HTTP 400) -------------------------
# Missing activityType
curl.exe -X POST `
    -H "Content-Type: application/json" `
    -d '{"action":"log","userId":"user1","language":"French","minutes":30,"date":"2026-05-01"}' `
    http://localhost:8080/api

# minutes out of range
curl.exe -X POST `
    -H "Content-Type: application/json" `
    -d '{"action":"log","userId":"user1","language":"French","activityType":"Reading","minutes":1500,"date":"2026-05-01"}' `
    http://localhost:8080/api

# Unknown action
curl.exe -X POST `
    -H "Content-Type: application/json" `
    -d '{"action":"delete","userId":"user1"}' `
    http://localhost:8080/api
