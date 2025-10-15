package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type FakeDynamo struct {
	Logs []map[string]types.AttributeValue
	mu   sync.Mutex
}

type TableBasics struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

func NewFakeDynamo() *FakeDynamo {
	return &FakeDynamo{
		Logs: make([]map[string]types.AttributeValue, 0),
	}
}

func (f *FakeDynamo) PutItem(ctx context.Context, in *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Store the item
	f.Logs = append(f.Logs, in.Item)

	// Log the stored item
	fmt.Println("📥 PutItem called, stored item:")
	for k, v := range in.Item {
		switch val := v.(type) {
		case *types.AttributeValueMemberS:
			fmt.Printf("  %s: %s\n", k, val.Value)
		case *types.AttributeValueMemberN:
			fmt.Printf("  %s: %s\n", k, val.Value)
		default:
			fmt.Printf("  %s: %+v\n", k, v)
		}
	}

	return &dynamodb.PutItemOutput{}, nil
}

func (f *FakeDynamo) Query(ctx context.Context, in *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Log query call
	fmt.Println("📤 Query called with key condition:", *in.KeyConditionExpression)
	for k, v := range in.ExpressionAttributeValues {
		switch val := v.(type) {
		case *types.AttributeValueMemberS:
			fmt.Printf("  %s = %s\n", k, val.Value)
		case *types.AttributeValueMemberN:
			fmt.Printf("  %s = %s\n", k, val.Value)
		default:
			fmt.Printf("  %s = %+v\n", k, v)
		}
	}

	// Return all items (simple implementation for fake)
	fmt.Printf("  Returning %d items\n", len(f.Logs))

	return &dynamodb.QueryOutput{
		Items: f.Logs,
	}, nil
}
