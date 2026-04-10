package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lang-tracker/internal/service"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDynamo implements db.DynamoDBClient for service-level tests.
type mockDynamo struct {
	items    []map[string]types.AttributeValue
	queryErr error
}

func (m *mockDynamo) PutItem(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, nil
}

func (m *mockDynamo) Query(_ context.Context, _ *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return &dynamodb.QueryOutput{Items: m.items}, m.queryErr
}

// marshalItems converts a slice of LogItem-like maps into DynamoDB attribute maps.
func marshalItem(t *testing.T, v any) map[string]types.AttributeValue {
	t.Helper()
	av, err := attributevalue.MarshalMap(v)
	if err != nil {
		t.Fatalf("failed to marshal item: %v", err)
	}
	return av
}

type logItemInput struct {
	UserID       string `dynamodbav:"userId"`
	LogID        string `dynamodbav:"logId"`
	Language     string `dynamodbav:"language"`
	ActivityType string `dynamodbav:"activityType"`
	Minutes      int    `dynamodbav:"minutes"`
	Date         string `dynamodbav:"date"`
}

func statsService(mock *mockDynamo) *service.StatsService {
	return &service.StatsService{DB: mock, TableName: "test"}
}

// --- today fix --------------------------------------------------------

func TestGetStats_TodayOnlyCountsCurrentDate(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	// Same day-of-month as today but in a different month (if possible)
	different := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	if different == today {
		t.Skip("cannot construct a different-month date with the same day from this date")
	}

	items := []map[string]types.AttributeValue{
		marshalItem(t, logItemInput{UserID: "u1", LogID: "1", Language: "French", ActivityType: "Reading", Minutes: 60, Date: today}),
		marshalItem(t, logItemInput{UserID: "u1", LogID: "2", Language: "French", ActivityType: "Reading", Minutes: 30, Date: different}),
	}
	svc := statsService(&mockDynamo{items: items})
	stats, err := svc.GetStats(context.Background(), "u1", "French")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only today's 60 minutes should count
	if stats.Today != 1.0 {
		t.Errorf("Today: expected 1.0 hours, got %f", stats.Today)
	}
}

// --- language filter --------------------------------------------------

func TestGetStats_FiltersOtherLanguages(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	items := []map[string]types.AttributeValue{
		marshalItem(t, logItemInput{UserID: "u1", LogID: "1", Language: "French", ActivityType: "Reading", Minutes: 60, Date: today}),
		marshalItem(t, logItemInput{UserID: "u1", LogID: "2", Language: "Spanish", ActivityType: "Reading", Minutes: 120, Date: today}),
	}
	svc := statsService(&mockDynamo{items: items})
	stats, err := svc.GetStats(context.Background(), "u1", "French")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalHours != 1.0 {
		t.Errorf("TotalHours: expected 1.0, got %f", stats.TotalHours)
	}
}

// --- percentages fix --------------------------------------------------

func TestGetStats_PercentagesAddUpTo100(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	items := []map[string]types.AttributeValue{
		marshalItem(t, logItemInput{UserID: "u1", LogID: "1", Language: "French", ActivityType: "Reading", Minutes: 60, Date: today}),
		marshalItem(t, logItemInput{UserID: "u1", LogID: "2", Language: "French", ActivityType: "Watching", Minutes: 60, Date: today}),
	}
	svc := statsService(&mockDynamo{items: items})
	stats, err := svc.GetStats(context.Background(), "u1", "French")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var total float64
	for _, pct := range stats.Percentages {
		total += pct
	}
	if fmt.Sprintf("%.2f", total) != "100.00" {
		t.Errorf("percentages should sum to 100, got %f", total)
	}
}

func TestGetStats_PercentagesEmptyWhenNoLogs(t *testing.T) {
	svc := statsService(&mockDynamo{items: nil})
	stats, err := svc.GetStats(context.Background(), "u1", "French")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats.Percentages) != 0 {
		t.Errorf("expected empty percentages map, got %v", stats.Percentages)
	}
}

// --- date format support ----------------------------------------------

func TestGetStats_AcceptsSlashDateFormat(t *testing.T) {
	// Date stored as MM/DD/YYYY
	today := time.Now()
	dateStr := fmt.Sprintf("%02d/%02d/%d", int(today.Month()), today.Day(), today.Year())
	items := []map[string]types.AttributeValue{
		marshalItem(t, logItemInput{UserID: "u1", LogID: "1", Language: "French", ActivityType: "Reading", Minutes: 30, Date: dateStr}),
	}
	svc := statsService(&mockDynamo{items: items})
	stats, err := svc.GetStats(context.Background(), "u1", "French")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Today != 0.5 {
		t.Errorf("Today: expected 0.5 hours, got %f", stats.Today)
	}
}

// --- bad date is skipped gracefully -----------------------------------

func TestGetStats_SkipsEntryWithUnparseableDate(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	items := []map[string]types.AttributeValue{
		marshalItem(t, logItemInput{UserID: "u1", LogID: "1", Language: "French", ActivityType: "Reading", Minutes: 60, Date: today}),
		marshalItem(t, logItemInput{UserID: "u1", LogID: "2", Language: "French", ActivityType: "Reading", Minutes: 30, Date: "not-a-date"}),
	}
	svc := statsService(&mockDynamo{items: items})
	stats, err := svc.GetStats(context.Background(), "u1", "French")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the valid entry should count
	if stats.TotalHours != 1.0 {
		t.Errorf("TotalHours: expected 1.0 (bad date skipped), got %f", stats.TotalHours)
	}
}

// --- DB error propagation ---------------------------------------------

func TestGetStats_DBErrorPropagates(t *testing.T) {
	svc := statsService(&mockDynamo{queryErr: fmt.Errorf("dynamo down")})
	_, err := svc.GetStats(context.Background(), "u1", "French")
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

// --- week boundary ----------------------------------------------------

func TestGetStats_ThisWeekDoesNotIncludeLastWeek(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	lastWeek := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	items := []map[string]types.AttributeValue{
		marshalItem(t, logItemInput{UserID: "u1", LogID: "1", Language: "French", ActivityType: "Reading", Minutes: 120, Date: today}),
		marshalItem(t, logItemInput{UserID: "u1", LogID: "2", Language: "French", ActivityType: "Reading", Minutes: 60, Date: lastWeek}),
	}
	svc := statsService(&mockDynamo{items: items})
	stats, err := svc.GetStats(context.Background(), "u1", "French")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// This week should only include the 120-minute entry
	if stats.ThisWeek != 2.0 {
		t.Errorf("ThisWeek: expected 2.0 hours (today only), got %f", stats.ThisWeek)
	}
}
