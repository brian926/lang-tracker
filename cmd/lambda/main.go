package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"lang-tracker/internal/db"
	"lang-tracker/internal/handler"
	"lang-tracker/internal/service"

	"github.com/aws/aws-lambda-go/lambda"
)

func tableName() string {
	if name := os.Getenv("TABLE_NAME"); name != "" {
		return name
	}
	return "lang-tracker"
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx := context.Background()
	client, err := db.New(ctx)
	if err != nil {
		log.Fatalf("failed to init DynamoDB client: %v", err)
	}

	table := tableName()
	svc := &handler.Services{
		Log:   &service.LogService{DB: client, TableName: table},
		Stats: &service.StatsService{DB: client, TableName: table},
	}

	lambda.Start(svc.Handler)
}
