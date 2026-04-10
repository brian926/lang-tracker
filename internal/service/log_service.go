package service

import (
	"context"
	"fmt"
	"lang-tracker/internal/db"
	"lang-tracker/internal/models"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/google/uuid"
)

// LogService handles writing activity log entries.
type LogService struct {
	DB        db.DynamoDBClient
	TableName string
}

// LogActivity writes a new log entry to DynamoDB.
func (s *LogService) LogActivity(ctx context.Context, req models.Request) error {
	item := models.LogItem{
		UserID:       req.UserID,
		LogID:        uuid.New().String(),
		Language:     req.Language,
		ActivityType: req.ActivityType,
		Minutes:      req.Minutes,
		Date:         req.Date,
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal log item: %w", err)
	}
	return db.CreateItem(ctx, s.DB, s.TableName, av)
}
