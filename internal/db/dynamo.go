package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Dynamo struct {
	Logs []map[string]types.AttributeValue
	mu   sync.Mutex
}

func NewDynamo() *Dynamo {
	return &Dynamo{
		Logs: make([]map[string]types.AttributeValue, 0),
	}
}

func CreateItem(ctx context.Context, tableName string, item map[string]types.AttributeValue) error {
	// Log the stored item
	fmt.Println("📥 PutItem called, stored item:")
	for k, v := range item {
		switch val := v.(type) {
		case *types.AttributeValueMemberS:
			fmt.Printf("  %s: %s\n", k, val.Value)
		case *types.AttributeValueMemberN:
			fmt.Printf("  %s: %s\n", k, val.Value)
		default:
			fmt.Printf("  %s: %+v\n", k, v)
		}
	}

	_, err := Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      item,
	})
	return err
}

func QueryByUserId(ctx context.Context, userID, tableName string) (*dynamodb.QueryOutput, error) {

	fmt.Println("📤 Querying DynamoDB for userId:", userID)

	out, err := Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &tableName,
		KeyConditionExpression: aws.String("userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	fmt.Printf("  ✅ Retrieved %d items\n", len(out.Items))
	return out, nil
}
