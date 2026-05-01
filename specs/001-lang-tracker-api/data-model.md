# Data Model: Lang-Tracker API (001-lang-tracker-api)

---

## Entities

### LogItem

Represents a single language-learning activity session logged by a user.

Stored in DynamoDB table `lang-tracker` (overridable via `TABLE_NAME` env var).

| Field | Go Type | DynamoDB Type | Key Role | Validation |
|-------|---------|---------------|----------|------------|
| `userId` | `string` | `S` | Partition Key | MUST be non-empty string |
| `logId` | `string` | `S` | Sort Key | UUID v4, generated at write time (server-side) |
| `language` | `string` | `S` | — | MUST be non-empty string |
| `activityType` | `string` | `S` | — | MUST be non-empty string |
| `minutes` | `int` | `N` | — | MUST be integer 1–1440 inclusive |
| `date` | `string` | `S` | — | MUST match `"2006-01-02"` or `"01/02/2006"` format |

**Go struct** (`internal/models/models.go:LogItem`):
```go
type LogItem struct {
    UserID       string `dynamodbav:"userId"`
    LogID        string `dynamodbav:"logId"`
    Language     string `dynamodbav:"language"`
    ActivityType string `dynamodbav:"activityType"`
    Minutes      int    `dynamodbav:"minutes"`
    Date         string `dynamodbav:"date"`
}
```

---

### Request

Inbound JSON payload for `POST /api`. A single struct serves both `log` and `stats` actions.

| Field | Go Type | JSON Key | Required for `log` | Required for `stats` |
|-------|---------|----------|--------------------|----------------------|
| `Action` | `string` | `action` | ✅ (`"log"`) | ✅ (`"stats"`) |
| `UserID` | `string` | `userId` | ✅ | ✅ |
| `Language` | `string` | `language` | ✅ | ✅ |
| `ActivityType` | `string` | `activityType` | ✅ | — |
| `Minutes` | `int` | `minutes` | ✅ (1–1440) | — |
| `Date` | `string` | `date` | ✅ | — |

**Go struct** (`internal/models/models.go:Request`):
```go
type Request struct {
    Action       string `json:"action"`
    UserID       string `json:"userId"`
    Language     string `json:"language"`
    ActivityType string `json:"activityType"`
    Minutes      int    `json:"minutes"`
    Date         string `json:"date"`
}
```

---

### StatsResponse

Aggregated statistics returned by the `stats` action. All time values in **hours** (float64).

| Field | Go Type | JSON Key | Description |
|-------|---------|----------|-------------|
| `TotalHours` | `float64` | `totalHours` | All-time total hours for user+language |
| `Today` | `float64` | `today` | Hours logged on today's calendar date |
| `ThisWeek` | `float64` | `thisWeek` | Hours in the current ISO week |
| `ThisMonth` | `float64` | `thisMonth` | Hours in the current calendar month |
| `Percentages` | `map[string]float64` | `percentages` | Activity type → % share of total time |

**Go struct** (`internal/models/models.go:StatsResponse`):
```go
type StatsResponse struct {
    TotalHours  float64            `json:"totalHours"`
    Today       float64            `json:"today"`
    ThisWeek    float64            `json:"thisWeek"`
    ThisMonth   float64            `json:"thisMonth"`
    Percentages map[string]float64 `json:"percentages"`
}
```

---

## DynamoDB Table Design

```
Table: lang-tracker
  Partition Key: userId (S)
  Sort Key:      logId  (S)  ← UUID v4

Access patterns supported:
  WRITE: PutItem with userId + generated logId
  READ:  Query by userId (paginated via LastEvaluatedKey)
         → application-side filter by language
         → application-side aggregation for stats
```

**No GSIs defined** in v1. The language filter happens in Go after fetching all items
for a user. A `(userId, language)` GSI is a known future optimization tracked in research.md.

---

## Validation Rules

### `log` action

| Rule | Error message |
|------|---------------|
| `userId` empty | `userId is required` |
| `language` empty | `language is required` |
| `activityType` empty | `activityType is required` |
| `minutes` ≤ 0 | `minutes must be a positive integer` |
| `minutes` > 1440 | `minutes cannot exceed 1440 (24 hours)` |
| `date` empty | `date is required (formats: YYYY-MM-DD or MM/DD/YYYY)` |

Multiple validation failures are concatenated with `"; "` and returned as a single 400 error.

### `stats` action

| Rule | Error message |
|------|---------------|
| `userId` empty | `userId is required` |
| `language` empty | `language is required` |

### Date parsing (stats computation)

- Accepted formats: `"2006-01-02"` (ISO 8601) and `"01/02/2006"` (US slash)
- Entries with unparseable dates are **skipped with a warning log** (not an error)
