package db_test

import (
	"context"
	"fmt"
	"testing"

	"lang-tracker/internal/db"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDynamo implements db.DynamoDBClient for db-package tests.
type mockDynamo struct {
	// pages simulates paginated Query results — each element is one page of items.
	pages    [][]map[string]types.AttributeValue
	pageIdx  int
	putItems []map[string]types.AttributeValue
	putErr   error
	queryErr error
}

func (m *mockDynamo) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if m.putErr != nil {
		return nil, m.putErr
	}
	m.putItems = append(m.putItems, in.Item)
	return &dynamodb.PutItemOutput{}, nil
}

func (m *mockDynamo) Query(_ context.Context, _ *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	if m.pageIdx >= len(m.pages) {
		return &dynamodb.QueryOutput{}, nil
	}
	page := m.pages[m.pageIdx]
	m.pageIdx++

	out := &dynamodb.QueryOutput{Items: page}
	// Simulate pagination: if more pages remain, set a non-nil LastEvaluatedKey.
	if m.pageIdx < len(m.pages) {
		out.LastEvaluatedKey = map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: "next-page-token"},
		}
	}
	return out, nil
}

// --- QueryByUserId tests -----------------------------------------------

func TestQueryByUserId_SinglePage(t *testing.T) {
	items := []map[string]types.AttributeValue{
		{"userId": &types.AttributeValueMemberS{Value: "u1"}},
		{"userId": &types.AttributeValueMemberS{Value: "u1"}},
	}
	mock := &mockDynamo{pages: [][]map[string]types.AttributeValue{items}}

	got, err := db.QueryByUserId(context.Background(), mock, "u1", "test-table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 items, got %d", len(got))
	}
}

func TestQueryByUserId_MultiPage(t *testing.T) {
	page1 := []map[string]types.AttributeValue{
		{"userId": &types.AttributeValueMemberS{Value: "u1"}},
	}
	page2 := []map[string]types.AttributeValue{
		{"userId": &types.AttributeValueMemberS{Value: "u1"}},
		{"userId": &types.AttributeValueMemberS{Value: "u1"}},
	}
	mock := &mockDynamo{pages: [][]map[string]types.AttributeValue{page1, page2}}

	got, err := db.QueryByUserId(context.Background(), mock, "u1", "test-table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 items from 2 pages, got %d", len(got))
	}
}

func TestQueryByUserId_EmptyResult(t *testing.T) {
	mock := &mockDynamo{pages: [][]map[string]types.AttributeValue{{}}}

	got, err := db.QueryByUserId(context.Background(), mock, "u1", "test-table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 items, got %d", len(got))
	}
}

func TestQueryByUserId_ErrorPropagates(t *testing.T) {
	mock := &mockDynamo{queryErr: fmt.Errorf("dynamo down")}

	_, err := db.QueryByUserId(context.Background(), mock, "u1", "test-table")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- CreateItem tests --------------------------------------------------

func TestCreateItem_Success(t *testing.T) {
	mock := &mockDynamo{}
	item := map[string]types.AttributeValue{
		"userId": &types.AttributeValueMemberS{Value: "u1"},
		"logId":  &types.AttributeValueMemberS{Value: "some-uuid"},
	}

	if err := db.CreateItem(context.Background(), mock, "test-table", item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.putItems) != 1 {
		t.Errorf("expected 1 put call, got %d", len(mock.putItems))
	}
}

func TestCreateItem_UsesCorrectTableName(t *testing.T) {
	captured := &struct{ tableName string }{}
	// We verify via the mock: PutItem receives the correct TableName.
	// Since our mock doesn't capture tableName directly, we use a custom mock.
	type capturingMock struct {
		mockDynamo
		tableName string
	}
	cm := &capturingMock{}
	// Override PutItem to capture table name.
	// Because Go doesn't allow method override in struct embedding for interfaces,
	// we use a func-based mock instead.
	funcMock := &funcMockDynamo{
		putFn: func(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			captured.tableName = aws.ToString(in.TableName)
			return &dynamodb.PutItemOutput{}, nil
		},
	}
	_ = cm // unused but shows intent

	item := map[string]types.AttributeValue{
		"userId": &types.AttributeValueMemberS{Value: "u1"},
	}
	if err := db.CreateItem(context.Background(), funcMock, "my-table", item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.tableName != "my-table" {
		t.Errorf("expected table name %q, got %q", "my-table", captured.tableName)
	}
}

func TestCreateItem_ErrorPropagates(t *testing.T) {
	mock := &mockDynamo{putErr: fmt.Errorf("write failed")}
	item := map[string]types.AttributeValue{
		"userId": &types.AttributeValueMemberS{Value: "u1"},
	}
	if err := db.CreateItem(context.Background(), mock, "test-table", item); err == nil {
		t.Error("expected error, got nil")
	}
}

// funcMockDynamo allows injecting arbitrary function implementations.
type funcMockDynamo struct {
	putFn   func(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	queryFn func(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

func (f *funcMockDynamo) PutItem(ctx context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if f.putFn != nil {
		return f.putFn(ctx, in, opts...)
	}
	return &dynamodb.PutItemOutput{}, nil
}

func (f *funcMockDynamo) Query(ctx context.Context, in *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if f.queryFn != nil {
		return f.queryFn(ctx, in, opts...)
	}
	return &dynamodb.QueryOutput{}, nil
}
