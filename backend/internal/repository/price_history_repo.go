// backend/internal/repository/price_history_repo.go
// 価格履歴データのDynamoDB操作を担当するリポジトリ
//
// 【キー設計】
//   PK: PRODUCT#<productId>    - パーティションキー（商品単位）
//   SK: PRICE#<timestamp>      - ソートキー（時系列順）
//
// 【時系列データのポイント】
//   - SK にタイムスタンプを含めることで、時系列順にソートされる
//   - ISO 8601形式（RFC3339）を使用することで、文字列の辞書順=時系列順になる
//   - BETWEEN クエリで範囲取得が可能
//   - ScanIndexForward=false で新しい順に取得

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

type priceHistoryRecord struct {
	PK        string `dynamodbav:"PK"` // PRODUCT#<productId>
	SK        string `dynamodbav:"SK"` // PRICE#<timestamp>
	ProductID string `dynamodbav:"productId"`
	Price     int    `dynamodbav:"price"`
	ChangedBy string `dynamodbav:"changedBy"` // 変更者（ユーザーID）
	ChangedAt string `dynamodbav:"changedAt"` // 変更日時（RFC3339形式）
}

type PriceHistoryRepository struct {
	db *DynamoDBClient
}

func NewPriceHistoryRepository(db *DynamoDBClient) *PriceHistoryRepository {
	return &PriceHistoryRepository{
		db: db,
	}
}

// Create は価格履歴をDynamoDBに保存する
// 【使用API】PutItem
// 【ポイント】タイムスタンプをSKに含めることで、同一商品の価格履歴を時系列で管理
func (r *PriceHistoryRepository) Create(ctx context.Context, history *domain.PriceHistory) error {
	now := time.Now()
	history.Timestamp = now

	// SK の形式: PRICE#2025-01-15T10:30:00Z
	// ISO 8601形式なので、文字列ソートすると時系列順になる
	record := priceHistoryRecord{
		PK:        "PRODUCT#" + history.ProductID,
		SK:        "PRICE#" + now.Format(time.RFC3339),
		ProductID: history.ProductID,
		Price:     history.Price,
		ChangedBy: history.ChangedBy,
		ChangedAt: now.Format(time.RFC3339),
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

// GetByProductID は商品の価格履歴を取得する（新しい順）
// 【使用API】Query + ScanIndexForward=false
//
// 【ScanIndexForward の役割】
//
//	true (デフォルト): ソートキーの昇順（古い→新しい）
//	false: ソートキーの降順（新しい→古い）
//
// 【Limit の役割】
//
//	取得する最大件数を指定
//	LastEvaluatedKey と組み合わせてページネーションに使用
func (r *PriceHistoryRepository) GetByProductID(ctx context.Context, productID string, limit int32) ([]*domain.PriceHistory, error) {
	input := &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"), // PRICE#で始まるSKを全て取得
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "PRODUCT#" + productID},
			":sk": &types.AttributeValueMemberS{Value: "PRICE#"},
		},
		ScanIndexForward: aws.Bool(false),  // 新しい順(降順)に取得
		Limit:            aws.Int32(limit), // 取得件数の上限
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	histories := make([]*domain.PriceHistory, 0, len(result.Items))
	for _, item := range result.Items {
		var record priceHistoryRecord
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		histories = append(histories, recordToPriceHistory(&record))
	}

	return histories, nil
}

// GetByProductIDWithRange は指定期間の価格履歴を取得する
// 【使用API】Query + BETWEEN
//
// 【BETWEEN の使い方】
//
//	SK BETWEEN :start AND :end
//	→ startからendの範囲のアイテムを取得
//	→ 時系列データの範囲クエリに最適
func (r *PriceHistoryRepository) GetByProductIDWithRange(ctx context.Context, productID string, startTime, endTime time.Time) ([]*domain.PriceHistory, error) {
	// SKの形式にあわせて時間をフォーマット
	startSK := "PRICE#" + startTime.Format(time.RFC3339)
	endSK := "PRICE#" + endTime.Format(time.RFC3339)

	input := &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND SK BETWEEN :start AND :end"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":    &types.AttributeValueMemberS{Value: "PRODUCT#" + productID},
			":start": &types.AttributeValueMemberS{Value: startSK},
			":end":   &types.AttributeValueMemberS{Value: endSK},
		},
		ScanIndexForward: aws.Bool(true), // 古い順(昇順)に取得しグラフ描画しやすくする
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	histories := make([]*domain.PriceHistory, 0, len(result.Items))
	for _, item := range result.Items {
		var record priceHistoryRecord
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		histories = append(histories, recordToPriceHistory(&record))
	}

	return histories, nil
}

func recordToPriceHistory(rec *priceHistoryRecord) *domain.PriceHistory {
	return &domain.PriceHistory{
		ProductID: rec.ProductID,
		Price:     rec.Price,
		ChangedBy: rec.ChangedBy,
		Timestamp: timeutil.ParseTime(rec.ChangedAt),
	}
}
