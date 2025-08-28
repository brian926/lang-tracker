package db

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var Client DynamoAPI

func Init() {
	ctx := context.TODO()

	// Check if we're running locally
	if os.Getenv("DYNAMO_LOCAL") == "1" {
		fmt.Println("🚀 Using DynamoDB Local")
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(os.Getenv("AWS_REGION")),
			config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					if service == dynamodb.ServiceID {
						return aws.Endpoint{URL: "http://localhost:8000"}, nil
					}
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				}),
			),
		)
		if err != nil {
			panic(err)
		}
		Client = dynamodb.NewFromConfig(cfg)
		return
	}

	// Production AWS
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config, %v", err))
	}
	Client = dynamodb.NewFromConfig(cfg)
}
