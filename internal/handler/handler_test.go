package handler_test

import (
	"context"
	"encoding/json"
	"testing"

	"lang-tracker/internal/handler"
	"lang-tracker/internal/models"
	"lang-tracker/internal/service"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDB is a no-op DynamoDB client for handler-level tests.
type mockDB struct {
	putErr   error
	queryOut *dynamodb.QueryOutput
	queryErr error
}

func (m *mockDB) PutItem(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, m.putErr
}

func (m *mockDB) Query(_ context.Context, _ *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if m.queryOut != nil {
		return m.queryOut, m.queryErr
	}
	return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, m.queryErr
}

func newServices(m *mockDB) *handler.Services {
	return &handler.Services{
		Log:   &service.LogService{DB: m, TableName: "test"},
		Stats: &service.StatsService{DB: m, TableName: "test"},
	}
}

func makeRequest(body any) events.APIGatewayProxyRequest {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyRequest{Body: string(b)}
}

// --- log action -------------------------------------------------------

func TestHandler_Log_Success(t *testing.T) {
	svc := newServices(&mockDB{})
	resp, err := svc.Handler(context.Background(), makeRequest(map[string]any{
		"action":       "log",
		"userId":       "user1",
		"language":     "French",
		"activityType": "Reading",
		"minutes":      30,
		"date":         "2026-04-10",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestHandler_Log_MissingFields(t *testing.T) {
	cases := []struct {
		name string
		body map[string]any
	}{
		{"missing userId", map[string]any{"action": "log", "language": "French", "activityType": "Reading", "minutes": 30, "date": "2026-04-10"}},
		{"missing language", map[string]any{"action": "log", "userId": "u1", "activityType": "Reading", "minutes": 30, "date": "2026-04-10"}},
		{"missing activityType", map[string]any{"action": "log", "userId": "u1", "language": "French", "minutes": 30, "date": "2026-04-10"}},
		{"zero minutes", map[string]any{"action": "log", "userId": "u1", "language": "French", "activityType": "Reading", "minutes": 0, "date": "2026-04-10"}},
		{"negative minutes", map[string]any{"action": "log", "userId": "u1", "language": "French", "activityType": "Reading", "minutes": -5, "date": "2026-04-10"}},
		{"minutes over 1440", map[string]any{"action": "log", "userId": "u1", "language": "French", "activityType": "Reading", "minutes": 1441, "date": "2026-04-10"}},
		{"missing date", map[string]any{"action": "log", "userId": "u1", "language": "French", "activityType": "Reading", "minutes": 30}},
	}
	svc := newServices(&mockDB{})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.Handler(context.Background(), makeRequest(tc.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != 400 {
				t.Errorf("expected 400, got %d: %s", resp.StatusCode, resp.Body)
			}
		})
	}
}

// --- stats action -----------------------------------------------------

func TestHandler_Stats_Success(t *testing.T) {
	svc := newServices(&mockDB{})
	resp, err := svc.Handler(context.Background(), makeRequest(map[string]any{
		"action":   "stats",
		"userId":   "user1",
		"language": "French",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	var body models.StatsResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid StatsResponse JSON: %v", err)
	}
}

func TestHandler_Stats_MissingFields(t *testing.T) {
	cases := []struct {
		name string
		body map[string]any
	}{
		{"missing userId", map[string]any{"action": "stats", "language": "French"}},
		{"missing language", map[string]any{"action": "stats", "userId": "u1"}},
	}
	svc := newServices(&mockDB{})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.Handler(context.Background(), makeRequest(tc.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != 400 {
				t.Errorf("expected 400, got %d: %s", resp.StatusCode, resp.Body)
			}
		})
	}
}

// --- edge cases -------------------------------------------------------

func TestHandler_InvalidJSON(t *testing.T) {
	svc := newServices(&mockDB{})
	resp, err := svc.Handler(context.Background(), events.APIGatewayProxyRequest{Body: "not json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandler_UnknownAction(t *testing.T) {
	svc := newServices(&mockDB{})
	resp, err := svc.Handler(context.Background(), makeRequest(map[string]any{"action": "delete"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandler_ResponseHasContentType(t *testing.T) {
	svc := newServices(&mockDB{})
	resp, _ := svc.Handler(context.Background(), makeRequest(map[string]any{
		"action":   "stats",
		"userId":   "u1",
		"language": "French",
	}))
	if ct := resp.Headers["Content-Type"]; ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
