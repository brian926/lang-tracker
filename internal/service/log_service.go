package service

import (
	"context"
	"lang-tracker/internal/db"
	"lang-tracker/internal/models"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/google/uuid"
)

const TableName = "lang-tracker"

func LogActivity(ctx context.Context, req models.Request) error {
	logID := uuid.New().String()
	item := models.LogItem{
		UserID:       req.UserID,
		LogID:        logID,
		Language:     req.Language,
		ActivityType: req.ActivityType,
		Minutes:      req.Minutes,
		Date:         req.Date,
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	// _, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
	// 	TableName: aws.String("LangLogs"),
	// 	Item:      av,
	// })
	// return err

	err = db.CreateItem(ctx, TableName, av)
	return err
}
