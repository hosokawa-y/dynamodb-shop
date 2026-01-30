package repository

import (
	"context"
	"errors"
	"time"

  	"github.com/aws/aws-sdk-go-v2/aws"
  	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
  	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
  	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
  	"github.com/google/uuid"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/timeutil"
)

var ErrProductNotFound = errors.New("product not found")

type productRecord struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	GSI1PK    string `dynamodbav:"GSI1PK"`
	GSI1SK    string `dynamodbav:"GSI1SK"`
	ID        string `dynamodbav:"id"`
	Name      string `dynamodbav:"name"`
	Description string `dynamodbav:"description"`
	Price     int    `dynamodbav:"price"`
	Category  string `dynamodbav:"category"`
	Stock     int    `dynamodbav:"stock"`
	ImageURL  string `dynamodbav:"imageUrl"`
	CreatedAt string `dynamodbav:"createdAt"`
	UpdatedAt string `dynamodbav:"updatedAt"`
}

type ProductRepository struct {
	db *DynamoDBClient
}

func NewProductRepository(db *DynamoDBClient) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	now := time.Now()
	product.ID = uuid.New().String()
	product.CreatedAt = now
	product.UpdatedAt = now

	record := productRecord{
		PK: "PRODUCT#" + product.ID,
		SK: "METADATA",
		GSI1PK: "PRODUCT",
		GSI1SK: "CATEGORY#" + product.Category + "#" + product.ID,
		ID: product.ID,
		Name: product.Name,
		Description: product.Description,
		Price: product.Price,
		Category: product.Category,
		Stock: product.Stock,
		ImageURL: product.ImageURL,
		CreatedAt: product.CreatedAt.Format(time.RFC3339),
		UpdatedAt: product.UpdatedAt.Format(time.RFC3339),
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.db.Table(),
		Item:      item,
	})

	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: r.db.Table(),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "PRODUCT#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, ErrProductNotFound
	}

	var record productRecord
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return nil, err
	}

	return recordToProduct(&record), nil
}

func (r *ProductRepository) List(ctx context.Context, category string)([]* domain.Product, error) {
	var input *dynamodb.QueryInput

	if category != "" {
		// カテゴリ指定あり：GSI1でフィルタ
		input = &dynamodb.QueryInput{
			TableName: r.db.Table(),
			IndexName: aws.String("GSI1"),
			KeyConditionExpression: aws.String("GSI1PK = :pk AND begins_with(GSI1SK, :sk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "PRODUCT"},
				":sk": &types.AttributeValueMemberS{Value: "CATEGORY#" + category },
			},
		}
	} else {
		// 全商品取得
		input = &dynamodb.QueryInput{
			TableName: r.db.Table(),
			IndexName: aws.String("GSI1"),
			KeyConditionExpression: aws.String("GSI1PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "PRODUCT"},
			},
		}
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	products := make([]*domain.Product, 0, len(result.Items))
	for _, item := range result.Items {
		var record productRecord
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		products = append(products, recordToProduct(&record))
	}
	
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	now := time.Now()
	
	record := productRecord{
		PK: "PRODUCT#" + product.ID,
		SK: "METADATA",
		GSI1PK: "PRODUCT",
		GSI1SK: "CATEGORY#" + product.Category + "#" + product.ID,
		ID: product.ID,
		Name: product.Name,
		Description: product.Description,
		Price: product.Price,
		Category: product.Category,
		Stock: product.Stock,
		ImageURL: product.ImageURL,
		CreatedAt: product.CreatedAt.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}
	
	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.db.Table(),
		Item:      item,
		ConditionExpression: aws.String("attribute_exists(PK)"), // 存在する場合のみ更新
	})

	return err
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: r.db.Table(),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "PRODUCT#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"), // 存在する場合のみ削除
	})
	
	return err
}

func recordToProduct(r *productRecord) *domain.Product {
	return &domain.Product{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Price:       r.Price,
		Category:    r.Category,
		Stock:       r.Stock,
		ImageURL:    r.ImageURL,
		CreatedAt:   timeutil.ParseTime(r.CreatedAt),
		UpdatedAt:   timeutil.ParseTime(r.UpdatedAt),
	}
}