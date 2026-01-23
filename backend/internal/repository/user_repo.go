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

var ErrUserNotFound = errors.New("user not found")
var ErrEmailAlreadyExists = errors.New("email already exists")

// DynamoDB用の内部構造体
type userRecord struct {
	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
	GSI1PK string `dynamodbav:"GSI1PK"`
	GSI1SK string `dynamodbav:"GSI1SK"`
	ID string `dynamodbav:"id"`
	Email string `dynamodbav:"email"`
	Name string `dynamodbav:"name"`
	PasswordHash string `dynamodbav:"passwordHash"`
	CreatedAt string `dynamodbav:"createdAt"`
	UpdatedAt string `dynamodbav:"updatedAt"`
}

type UserRepository struct {
	db *DynamoDBClient
}

func NewUserRepository(db *DynamoDBClient) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User)error {
	now := time.Now()
	user.ID = uuid.New().String()
	user.CreatedAt = now
	user.UpdatedAt = now

	record := userRecord{
		PK: "USER#" + user.ID, // USER#の#はDynamoDBのSingle Table Designの区切り文字
		SK: "PROFILE",
		GSI1PK: "USER",
		GSI1SK: "EMAIL#" + user.Email, // メールアドレスでの検索用
		ID: user.ID,
		Email: user.Email,
		Name: user.Name,
		PasswordHash: user.PasswordHash,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}

	// ConditionExpression: 条件付き書き込み
	// - ここでは「PKが存在しない場合のみ書き込む」という条件を指定している
	// - 既に同じPKが存在する場合はConditionalCheckFailedExceptionエラー
	// - これにより重複登録を防止
	// ConditionExpressionがないと、PutItemは同じPKのアイテムを無条件で上書きしてしまう
	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.db.Table(),
		Item:      item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	// GetItem: PK+SKを完全一致で指定して単一アイテムを取得
	// - 最速かつ最小コスト（直接アクセス）
	// - 結果は0件または1件のみ
	// - Queryとの違い: Queryは複数件取得可能、GetItemは1件のみ
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: r.db.Table(),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + id},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
	})
	if err != nil {
		return nil, err
	}
	// アイテムが見つからない場合、result.Itemはnilになる
	if result.Item == nil {
		return nil, ErrUserNotFound
	}
	
	var record userRecord
	if err = attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           record.ID,
		Email:        record.Email,
		Name:         record.Name,
		PasswordHash: record.PasswordHash,
		CreatedAt:    timeutil.ParseTime(record.CreatedAt),
		UpdatedAt:    timeutil.ParseTime(record.UpdatedAt),
	}, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error){
	// Query: 条件に一致する複数アイテムを取得
	// - GetItemとの違い: PKだけでなくSKにも条件（範囲・前方一致など）を指定可能
	// - GSI（グローバルセカンダリインデックス）を使用してemail検索を実現
	//   メインテーブルのPKはUSER#<id>なので、IDがわからないとGetItemできない
	//   GSI1を使うことでemailからユーザーを検索可能にしている
	// - KeyConditionExpressionでSKに使える演算子: =, begins_with, BETWEEN, <, <=, >, >=
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName: r.db.Table(),
		IndexName: aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk AND GSI1SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER"},
			":sk": &types.AttributeValueMemberS{Value: "EMAIL#" + email},
		},
	})
	if err != nil {
		return nil, err
	}
	// Queryは複数件返す可能性があるためスライスで返却される
	if len(result.Items) == 0 {
		return nil, ErrUserNotFound
	}

	var record userRecord
	if err = attributevalue.UnmarshalMap(result.Items[0], &record); err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           record.ID,
		Email:        record.Email,
		Name:         record.Name,
		PasswordHash: record.PasswordHash,
		CreatedAt:    timeutil.ParseTime(record.CreatedAt),
		UpdatedAt:    timeutil.ParseTime(record.UpdatedAt),
	}, nil
}