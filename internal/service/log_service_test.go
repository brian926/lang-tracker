package service_test

import (
	"context"
	"fmt"
	"testing"

	"lang-tracker/internal/models"
	"lang-tracker/internal/service"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// logMockDynamo implements db.DynamoDBClient for LogService tests.
type logMockDynamo struct {
	putErr      error
	capturedItem map[string]types.AttributeValue
}

func (m *logMockDynamo) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if m.putErr != nil {
		return nil, m.putErr
	}
	m.capturedItem = in.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (m *logMockDynamo) Query(_ context.Context, _ *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return &dynamodb.QueryOutput{}, nil
}

func logService(mock *logMockDynamo) *service.LogService {
	return &service.LogService{DB: mock, TableName: "test"}
}

// --- TestLogActivity_Success -------------------------------------------

func TestLogActivity_Success(t *testing.T) {
	mock := &logMockDynamo{}
	svc := logService(mock)

	req := models.Request{
		Action:       "log",
		UserID:       "user-1",
		Language:     "French",
		ActivityType: "Reading",
		Minutes:      30,
		Date:         "2026-05-01",
	}

	if err := svc.LogActivity(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.capturedItem == nil {
		t.Fatal("expected PutItem to be called, but capturedItem is nil")
	}
}

// --- TestLogActivity_AllFieldsPersisted --------------------------------

func TestLogActivity_AllFieldsPersisted(t *testing.T) {
	mock := &logMockDynamo{}
	svc := logService(mock)

	req := models.Request{
		UserID:       "user-2",
		Language:     "Spanish",
		ActivityType: "Watching",
		Minutes:      45,
		Date:         "2026-04-15",
	}

	if err := svc.LogActivity(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Unmarshal the captured item back to a LogItem to verify all fields.
	var item models.LogItem
	if err := attributevalue.UnmarshalMap(mock.capturedItem, &item); err != nil {
		t.Fatalf("failed to unmarshal captured item: %v", err)
	}

	if item.UserID != req.UserID {
		t.Errorf("UserID: expected %q, got %q", req.UserID, item.UserID)
	}
	if item.Language != req.Language {
		t.Errorf("Language: expected %q, got %q", req.Language, item.Language)
	}
	if item.ActivityType != req.ActivityType {
		t.Errorf("ActivityType: expected %q, got %q", req.ActivityType, item.ActivityType)
	}
	if item.Minutes != req.Minutes {
		t.Errorf("Minutes: expected %d, got %d", req.Minutes, item.Minutes)
	}
	if item.Date != req.Date {
		t.Errorf("Date: expected %q, got %q", req.Date, item.Date)
	}
}

// --- TestLogActivity_UUIDGenerated -------------------------------------

func TestLogActivity_UUIDGenerated(t *testing.T) {
	mock := &logMockDynamo{}
	svc := logService(mock)

	req := models.Request{
		UserID:       "user-3",
		Language:     "French",
		ActivityType: "Speaking",
		Minutes:      20,
		Date:         "2026-05-01",
	}

	if err := svc.LogActivity(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var item models.LogItem
	if err := attributevalue.UnmarshalMap(mock.capturedItem, &item); err != nil {
		t.Fatalf("failed to unmarshal captured item: %v", err)
	}

	if item.LogID == "" {
		t.Error("expected a non-empty logId UUID, got empty string")
	}
}

// --- TestLogActivity_UniqueUUIDs ---------------------------------------

func TestLogActivity_UniqueUUIDs(t *testing.T) {
	// Two separate calls must generate different logIds.
	mock1 := &logMockDynamo{}
	mock2 := &logMockDynamo{}
	svc1 := logService(mock1)
	svc2 := logService(mock2)

	req := models.Request{
		UserID:       "user-4",
		Language:     "French",
		ActivityType: "Reading",
		Minutes:      10,
		Date:         "2026-05-01",
	}

	if err := svc1.LogActivity(context.Background(), req); err != nil {
		t.Fatalf("call 1 error: %v", err)
	}
	if err := svc2.LogActivity(context.Background(), req); err != nil {
		t.Fatalf("call 2 error: %v", err)
	}

	var item1, item2 models.LogItem
	_ = attributevalue.UnmarshalMap(mock1.capturedItem, &item1)
	_ = attributevalue.UnmarshalMap(mock2.capturedItem, &item2)

	if item1.LogID == item2.LogID {
		t.Errorf("expected unique logIds, both got %q", item1.LogID)
	}
}

// --- TestLogActivity_PutItemErrorPropagates ----------------------------

func TestLogActivity_PutItemErrorPropagates(t *testing.T) {
	mock := &logMockDynamo{putErr: fmt.Errorf("dynamo unavailable")}
	svc := logService(mock)

	req := models.Request{
		UserID:       "user-5",
		Language:     "French",
		ActivityType: "Reading",
		Minutes:      15,
		Date:         "2026-05-01",
	}

	if err := svc.LogActivity(context.Background(), req); err == nil {
		t.Error("expected error from PutItem, got nil")
	}
}
