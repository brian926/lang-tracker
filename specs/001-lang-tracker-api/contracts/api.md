# API Contract: POST /api

**Endpoint**: `POST /api`  
**Content-Type**: `application/json`  
**Platform**: AWS Lambda (API Gateway Proxy) / Local HTTP server on `:8080`

All requests share the same endpoint and are routed by the `action` field.

---

## Action: `log` — Write Activity Log

### Request

```json
{
  "action": "log",
  "userId": "string (required, non-empty)",
  "language": "string (required, non-empty)",
  "activityType": "string (required, non-empty)",
  "minutes": "integer (required, 1–1440)",
  "date": "string (required, YYYY-MM-DD or MM/DD/YYYY)"
}
```

**Example**:
```json
{
  "action": "log",
  "userId": "user-abc",
  "language": "French",
  "activityType": "Reading",
  "minutes": 45,
  "date": "2026-05-01"
}
```

### Success Response — HTTP 200

```json
{
  "message": "Log saved"
}
```

### Validation Error — HTTP 400

One or more validation rules violated. Multiple failures are reported in a single response,
concatenated with `"; "`.

```json
{
  "error": "userId is required; minutes must be a positive integer"
}
```

Possible error messages:
- `userId is required`
- `language is required`
- `activityType is required`
- `minutes must be a positive integer`
- `minutes cannot exceed 1440 (24 hours)`
- `date is required (formats: YYYY-MM-DD or MM/DD/YYYY)`

### Internal Error — HTTP 500

```json
{
  "error": "internal server error"
}
```

---

## Action: `stats` — Get Aggregated Statistics

### Request

```json
{
  "action": "stats",
  "userId": "string (required, non-empty)",
  "language": "string (required, non-empty)"
}
```

**Example**:
```json
{
  "action": "stats",
  "userId": "user-abc",
  "language": "French"
}
```

### Success Response — HTTP 200

```json
{
  "totalHours": 12.5,
  "today": 0.75,
  "thisWeek": 3.0,
  "thisMonth": 8.25,
  "percentages": {
    "Reading": 60.0,
    "Watching": 30.0,
    "Speaking": 10.0
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `totalHours` | float64 | All-time total hours for the user+language combination |
| `today` | float64 | Hours logged on today's calendar date (UTC wall-clock) |
| `thisWeek` | float64 | Hours in the current ISO 8601 week |
| `thisMonth` | float64 | Hours in the current calendar month |
| `percentages` | object | Activity type → % share of total time (sums to 100 when logs exist; empty object when no logs) |

### Validation Error — HTTP 400

```json
{
  "error": "userId is required"
}
```

Possible error messages:
- `userId is required`
- `language is required`

### Internal Error — HTTP 500

```json
{
  "error": "internal server error"
}
```

---

## Action: Unknown

Any `action` value other than `"log"` or `"stats"`:

### Response — HTTP 400

```json
{
  "error": "unknown action: must be \"log\" or \"stats\""
}
```

---

## Malformed JSON

Body that cannot be decoded as JSON:

### Response — HTTP 400

```json
{
  "error": "invalid JSON body"
}
```

---

## Notes

- All responses have `Content-Type: application/json`.
- `logId` is a UUID v4 generated server-side; clients do not supply it.
- No authentication — any caller can read/write any `userId` (v1 constraint per roadmap).
- `date` strings in DynamoDB entries may be in either accepted format; stats computation
  handles both transparently.
