// backend/internal/repository/activity_repo.go
// ユーザー行動ログのDynamoDB操作を担当するリポジトリ
//
// 【キー設計】
//   PK: USER#<userId>             - パーティションキー（ユーザー単位）
//   SK: ACTIVITY#<timestamp>      - ソートキー（時系列順）
//
// 【TTL (Time To Live)】
//   DynamoDBのTTL機能を使用して、30日後に自動削除
//   TTL属性にはUnix Epoch秒（int64）を設定
//   DynamoDBが定期的にスキャンし、TTLを過ぎたアイテムを自動削除
//
// 【BatchWriteItem】
//   最大25件まで一度に書き込み可能
//   UnprocessedItems がある場合は再試行が必要

package repository

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/timeutil"
)

const (
	// 行動ログの保持期間（30日）
	TTLDuation = 30 * 24 * time.Hour
	// BatchWriteItemの最大件数
	MaxBatchWriteItems = 25
)

type activityRecord struct {
	PK         string            `dynamodbav:"PK"` // USER#<userId>
	SK         string            `dynamodbav:"SK"` // ACTIVITY#<timestamp>
	UserID     string            `dynamodbav:"UserId"`
	ActionType string            `dynamodbav:"ActionType"` // VIEW, CLICK, ADD_CART, PURCHASE
	ProductID  string            `dynamodbav:"ProductId"`
	Metadata   map[string]string `dynamodbav:"Metadata,omitempty"`
	TTL        int64             `dynamodbav:"TTL"` // Unix Epoch秒（30日後）
	CreatedAt  string            `dynamodbav:"CreatedAt"`
}

type ActivityRepository struct {
	db *DynamoDBClient
}

func NewActivityRepository(db *DynamoDBClient) *ActivityRepository {
	return &ActivityRepository{
		db: db,
	}
}

// Createは行動ログを1件保存する
// 【TTL】30日後のUnix Epoch秒を設定
func (r *ActivityRepository) Create(ctx context.Context, activity *domain.UserActivity) error {
	now := time.Now()
	activity.Timestamp = now
	activity.TTL = now.Add(TTLDuation).Unix() // 30日後のUnix Epoch秒

	record := activityRecord{
		PK:         "USER#" + activity.UserID,
		SK:         "ACTIVITY#" + now.Format(time.RFC3339Nano), // nano秒まで使用して重複を防ぐ
		UserID:     activity.UserID,
		ActionType: activity.ActionType,
		ProductID:  activity.ProductID,
		Metadata:   activity.Metadata,
		TTL:        activity.TTL,
		CreatedAt:  now.Format(time.RFC3339),
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

// BatchCreate は複数の行動ログを一括保存する
// 【使用API】BatchWriteItem
// 【制限】最大25件まで。それ以上の場合は分割して呼び出す
// 【リトライ】UnprocessedItemsがある場合は再試行
func (r *ActivityRepository) BatchCreate(ctx context.Context, activities []*domain.UserActivity) error {
	if len(activities) == 0 {
		return nil
	}

	now := time.Now()

	// 25件ずつに分割して処理
	for i := 0; i < len(activities); i += MaxBatchWriteItems {
		end := i + MaxBatchWriteItems
		if end > len(activities) {
			end = len(activities)
		}
		batch := activities[i:end]
		writeRequests := make([]types.WriteRequest, 0, len(batch))

		for j, activity := range batch {
			// タイムスタンプをずらして重複を防ぐ
			timestamp := now.Add(time.Duration(j) * time.Nanosecond)
			activity.Timestamp = timestamp
			activity.TTL = timestamp.Add(TTLDuation).Unix()

			record := activityRecord{
				PK:         "USER#" + activity.UserID,
				SK:         "ACTIVITY#" + timestamp.Format(time.RFC3339Nano),
				UserID:     activity.UserID,
				ActionType: activity.ActionType,
				ProductID:  activity.ProductID,
				Metadata:   activity.Metadata,
				TTL:        activity.TTL,
				CreatedAt:  timestamp.Format(time.RFC3339),
			}

			item, err := attributevalue.MarshalMap(record)
			if err != nil {
				return err
			}

			writeRequests = append(writeRequests, types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: item,
				},
			})
		}

		// BatchWriteItem実行
		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				*r.db.Table(): writeRequests,
			},
		}

		result, err := r.db.Client.BatchWriteItem(ctx, input)
		if err != nil {
			return err
		}

		// UnprocessedItemsがある場合は再試行(簡易版)
		// 本番環境ではExponential Backoffを実装すべき
		for len(result.UnprocessedItems) > 0 {
			time.Sleep(100 * time.Millisecond)
			input.RequestItems = result.UnprocessedItems
			result, err = r.db.Client.BatchWriteItem(ctx, input)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetByUserID はユーザーの行動ログを取得する（新しい順）
// 【使用API】Query + ScanIndexForward=false
func (r *ActivityRepository) GetByUserID(ctx context.Context, userID string, limit int32) ([]*domain.UserActivity, error) {
	input := &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk": &types.AttributeValueMemberS{Value: "ACTIVITY#"},
		},
		ScanIndexForward: aws.Bool(false), // 新しい順
		Limit:            aws.Int32(limit),
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	activities := make([]*domain.UserActivity, 0, len(result.Items))
	for _, item := range result.Items {
		var rec activityRecord
		if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
			return nil, err
		}
		activities = append(activities, recordToActivity(&rec))
	}

	return activities, nil
}

// GetByUserIDAndAction は特定アクションタイプの行動ログを取得する
// 【使用API】Query + FilterExpression
func (r *ActivityRepository) GetByUserIDAndAction(ctx context.Context, userID string, actionType string, limit int32) ([]*domain.UserActivity, error) {
	input := &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		FilterExpression:       aws.String("ActionType = :actionType"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":         &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk":         &types.AttributeValueMemberS{Value: "ACTIVITY#"},
			":actionType": &types.AttributeValueMemberS{Value: actionType},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(limit),
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	activities := make([]*domain.UserActivity, 0, len(result.Items))
	for _, item := range result.Items {
		var rec activityRecord
		if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
			return nil, err
		}
		activities = append(activities, recordToActivity(&rec))
	}

	return activities, nil
}

func recordToActivity(rec *activityRecord) *domain.UserActivity {
	return &domain.UserActivity{
		UserID:     rec.UserID,
		ActionType: rec.ActionType,
		ProductID:  rec.ProductID,
		Metadata:   rec.Metadata,
		TTL:        rec.TTL,
		Timestamp:  timeutil.ParseTime(rec.CreatedAt),
	}
}
