package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient is the interface the service layer depends on.
// The real *dynamodb.Client satisfies it; tests can inject a mock.
type DynamoDBClient interface {
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, in *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// CreateItem writes a single item to the given table.
func CreateItem(ctx context.Context, client DynamoDBClient, tableName string, item map[string]types.AttributeValue) error {
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

// QueryByUserId returns all items for a userId, following DynamoDB pagination.
func QueryByUserId(ctx context.Context, client DynamoDBClient, userID, tableName string) ([]map[string]types.AttributeValue, error) {
	var (
		items            []map[string]types.AttributeValue
		lastEvaluatedKey map[string]types.AttributeValue
	)

	for {
		input := &dynamodb.QueryInput{
			TableName:              aws.String(tableName),
			KeyConditionExpression: aws.String("userId = :uid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":uid": &types.AttributeValueMemberS{Value: userID},
			},
		}
		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}

		out, err := client.Query(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}

		items = append(items, out.Items...)

		if out.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = out.LastEvaluatedKey
	}

	return items, nil
}
