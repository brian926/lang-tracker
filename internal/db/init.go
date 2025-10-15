package db

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var Client DynamoAPI

func Init() {
	ctx := context.TODO()

	// Production AWS
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config, %v", err))
	}
	Client = dynamodb.NewFromConfig(cfg)
	log.Println("✅ DynamoDB client initialized")
}
