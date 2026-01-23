package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	Client    *dynamodb.Client
	TableName string
}

func NewDynamoDBClient(ctx context.Context, tableName, endpoint, region string) (*DynamoDBClient, error) {
	var cfg aws.Config
	var err error

	if endpoint != "" {
		// ローカル開発用（DynamoDB Local）
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(
					func(service, region string, options ...interface{}) (aws.Endpoint, error) {
						return aws.Endpoint{URL: endpoint}, nil
					},
				),
			),
		)
	} else {
		// AWS実環境
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDBClient{
		Client:    client,
		TableName: tableName,
	}, nil
}
