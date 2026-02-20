// backend/internal/repository/inventory_repo.go
// 在庫変動ログのDynamoDB操作を担当するリポジトリ
//
// 【キー設計】
//   PK: PRODUCT#<productId>    - パーティションキー（商品単位）
//   SK: INVLOG#<timestamp>     - ソートキー（時系列順）
//
// 【ChangeType】
//   IN:     入庫（仕入れ）
//   OUT:    出庫（注文による減少）
//   ADJUST: 調整（棚卸し、誤差修正など）

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

type inventoryLogRecord struct {
	PK            string `dynamodbav:"PK"` // PRODUCT#<productId>
	SK            string `dynamodbav:"SK"` // INVLOG#<timestamp>
	ProductID     string `dynamodbav:"productId"`
	ChangeType    string `dynamodbav:"changeType"`        // IN, OUT, ADJUST
	Quantity      int    `dynamodbav:"quantity"`          // 変動数量（正の値）
	PreviousStock int    `dynamodbav:"previousStock"`     // 変更前在庫
	NewStock      int    `dynamodbav:"newStock"`          // 変更後在庫
	Reason        string `dynamodbav:"reason"`            // 変更理由
	OrderID       string `dynamodbav:"orderId,omitempty"` // 注文ID（注文起因の場合）
	CreatedAt     string `dynamodbav:"createdAt"`
}

type InventoryRepository struct {
	db *DynamoDBClient
}

func NewInventoryRepository(db *DynamoDBClient) *InventoryRepository {
	return &InventoryRepository{
		db: db,
	}
}

// Create は在庫変動ログをDynamoDBに保存する
// 【使用API】PutItem
func (r *InventoryRepository) Create(ctx context.Context, log *domain.InventoryLog) error {
	now := time.Now()
	log.Timestamp = now

	record := inventoryLogRecord{
		PK:            "PRODUCT#" + log.ProductID,
		SK:            "INVLOG#" + now.Format(time.RFC3339),
		ProductID:     log.ProductID,
		ChangeType:    log.ChangeType,
		Quantity:      log.Quantity,
		PreviousStock: log.PreviousStock,
		NewStock:      log.NewStock,
		Reason:        log.Reason,
		OrderID:       log.OrderID,
		CreatedAt:     now.Format(time.RFC3339),
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

// GetByProductID は商品の在庫変動履歴を取得する（新しい順）
// 【使用API】Query + ScanIndexForward=false + Limit
func (r *InventoryRepository) GetByProductID(ctx context.Context, productID string, limit int32) ([]*domain.InventoryLog, error) {
	input := &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "PRODUCT#" + productID},
			":sk": &types.AttributeValueMemberS{Value: "INVLOG#"},
		},
		ScanIndexForward: aws.Bool(false), // 新しい順
		Limit:            aws.Int32(limit),
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	logs := make([]*domain.InventoryLog, 0, len(result.Items))
	for _, item := range result.Items {
		var rec inventoryLogRecord
		if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
			return nil, err
		}
		logs = append(logs, recordToInventoryLog(&rec))
	}

	return logs, nil
}

// GetByProductIDWithRange は指定期間の在庫変動履歴を取得する
// 【使用API】Query + BETWEEN

func (r *InventoryRepository) GetByProductIDWithRange(ctx context.Context, productID string, startTime, endTime time.Time) ([]*domain.InventoryLog, error) {
	startSK := "INVLOG#" + startTime.Format(time.RFC3339)
	endSK := "INVLOG#" + endTime.Format(time.RFC3339)

	input := &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND SK BETWEEN :start AND :end"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":    &types.AttributeValueMemberS{Value: "PRODUCT#" + productID},
			":start": &types.AttributeValueMemberS{Value: startSK},
			":end":   &types.AttributeValueMemberS{Value: endSK},
		},
		ScanIndexForward: aws.Bool(false), // 新しい順
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	logs := make([]*domain.InventoryLog, 0, len(result.Items))
	for _, item := range result.Items {
		var rec inventoryLogRecord
		if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
			return nil, err
		}
		logs = append(logs, recordToInventoryLog(&rec))
	}

	return logs, nil
}

func recordToInventoryLog(rec *inventoryLogRecord) *domain.InventoryLog {
	return &domain.InventoryLog{
		ProductID:     rec.ProductID,
		ChangeType:    rec.ChangeType,
		Quantity:      rec.Quantity,
		PreviousStock: rec.PreviousStock,
		NewStock:      rec.NewStock,
		Reason:        rec.Reason,
		OrderID:       rec.OrderID,
		Timestamp:     timeutil.ParseTime(rec.CreatedAt),
	}
}
