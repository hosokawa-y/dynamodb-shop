// cart_repo.go
// カートデータのDynamoDB操作を担当するリポジトリ
//
// 【キー設計】
//   PK: USER#<ユーザーID>     - パーティションキー（ユーザー単位）
//   SK: CART#<商品ID>        - ソートキー（商品単位）
//
// 【アクセスパターン】
//   1. ユーザーのカート全件取得  → Query(PK = "USER#xxx" AND begins_with(SK, "CART#"))
//   2. カートアイテム1件取得    → GetItem(PK, SK)
//   3. カートにアイテム追加     → PutItem
//   4. 数量更新（楽観的ロック）  → UpdateItem + ConditionExpression
//   5. カートからアイテム削除   → DeleteItem

package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/timeutil"
)

var ErrCartItemNotFound = errors.New("cart item not found")
var ErrVersionMismatch = errors.New("version mismatch: item was modified by another request")

// cartRecord はDynamoDBに保存するカートデータの構造体
type cartRecord struct {
	PK          string `dynamodbav:"PK"` // USER#<userId>
	SK          string `dynamodbav:"SK"` // CART#<productId>
	UserID      string `dynamodbav:"userId"`
	ProductID   string `dynamodbav:"productId"`
	ProductName string `dynamodbav:"productName"` // 非正規化（商品名をカートに保持）
	Price       int    `dynamodbav:"price"`       // 非正規化（カート追加時の価格）
	Quantity    int    `dynamodbav:"quantity"`
	Version     int    `dynamodbav:"version"` // 楽観的ロック用
	AddedAt     string `dynamodbav:"addedAt"`
	UpdatedAt   string `dynamodbav:"updatedAt"`
}

// CartRepository はカートのDynamoDB操作を提供する
type CartRepository struct {
	db *DynamoDBClient
}

// NewCartRepository は CartRepository のインスタンスを生成する
func NewCartRepository(db *DynamoDBClient) *CartRepository {
	return &CartRepository{
		db: db,
	}
}

// Add はカートにアイテムを追加する
// 【使用API】PutItem
// 【注意】同じ商品が既にある場合は上書き（数量を加算したい場合はUpdateを使う）
func (r *CartRepository) Add(ctx context.Context, item *domain.CartItem) error {
	now := time.Now()
	item.Version = 1 // 新規追加時のバージョンは1
	item.AddedAt = now
	item.UpdatedAt = now

	record := cartRecord{
		PK:          "USER#" + item.UserID,
		SK:          "CART#" + item.ProductID,
		UserID:      item.UserID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		Price:       item.Price,
		Quantity:    item.Quantity,
		Version:     item.Version,
		AddedAt:     item.AddedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}

	av, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.db.Table(),
		Item:      av,
	})

	return err
}

// GetByUserID はユーザーのカートアイテム全件を取得する
// 【使用API】Query - PKで絞り込み、SKのプレフィックスで「CART#」のみ取得
func (r *CartRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.CartItem, error) {
	// Query: PK = USER#xxx AND SK begins_with "CART#"
	// → ユーザーに紐づくカートアイテムだけを取得
	// → 同じユーザーの ORDER# や PROFILE は取得しない
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk": &types.AttributeValueMemberS{Value: "CART#"},
		},
	})
	if err != nil {
		return nil, err
	}

	items := make([]*domain.CartItem, 0, len(result.Items))
	for _, item := range result.Items {
		var record cartRecord
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		items = append(items, recordToCartItem(&record))
	}

	return items, nil
}

// GetItem は特定のカートアイテムを1件取得する
// 【使用API】GetItem - PK+SKで1件取得
func (r *CartRepository) GetItem(ctx context.Context, userID, productID string) (*domain.CartItem, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: r.db.Table(),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "CART#" + productID},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, ErrCartItemNotFound
	}

	var record cartRecord
	if err = attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return nil, err
	}

	return recordToCartItem(&record), nil
}

// UpdateQuantity はカートアイテムの数量を更新する（楽観的ロック付き）
// 【使用API】UpdateItem + ConditionExpression
//
// 【楽観的ロックの仕組み】
//  1. クライアントは現在のVersionを送信
//  2. ConditionExpression で「DBのVersion = 送信されたVersion」をチェック
//  3. 条件を満たす場合のみ更新を実行し、Versionを+1
//  4. 条件を満たさない場合は ConditionalCheckFailedException
//     → 他のリクエストが先に更新したことを意味する
func (r *CartRepository) UpdateQuantity(ctx context.Context, userID, productID string, quantity, currentVersion int) error {
	now := time.Now()
	newVesrion := currentVersion + 1

	// UpdateItem: 指定した属性のみを更新（PutItemと違い全属性を指定する必要がない）
	// SET: 属性の値を設定
	// ConditionExpression: version = :currentVer の場合のみ更新を実行
	_, err := r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: r.db.Table(),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "CART#" + productID},
		},
		UpdateExpression: aws.String("SET quantity = :qty, version = :newVer, updatedAt = :now"),
		// ConditionExpression: 楽観的ロックの条件
		// DBに保存されているversionと、リクエストで送られたversionが一致する場合のみ更新
		ConditionExpression: aws.String("version = :currentVer"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":qty":        &types.AttributeValueMemberN{Value: strconv.Itoa(quantity)},
			":currentVer": &types.AttributeValueMemberN{Value: strconv.Itoa(currentVersion)},
			":newVer":     &types.AttributeValueMemberN{Value: strconv.Itoa(newVesrion)},
			":now":        &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		},
	})
	if err != nil {
		// ConditionalCheckFailedException を判定
		var cfe *types.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			return ErrVersionMismatch
		}
		return err
	}

	return nil
}

// Delete はカートからアイテムを削除する
// 【使用API】DeleteItem
func (r *CartRepository) Delete(ctx context.Context, userID, productID string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: r.db.Table(),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "CART#" + productID},
		},
		// 存在しない場合もエラーにしたい場合は以下を追加
		// ConditionExpression: aws.String("attribute_exists(PK)"),
	})
	return err
}

// Clear はユーザーのカートを全て削除する
// 【使用API】Query + BatchWriteItem
// 【注意】BatchWriteItemは最大25件まで。カートが25件を超える場合は分割が必要
func (r *CartRepository) Clear(ctx context.Context, userID string) error {
	// まずカートアイテムを全件取得
	items, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil // 削除するものがない
	}

	// BatchWriteItem用のリクエストを作成
	// 【BatchWriteItemの特徴】
	//   - 最大25件のPut/Deleteを1回のAPIコールで実行
	//   - 個別にDeleteItemを呼ぶより効率的（API呼び出し回数削減）
	//   - 全件成功 or 全件失敗ではない（部分的な失敗あり）
	//   - 失敗したアイテムはUnprocessedItemsで返却される
	writeRequests := make([]types.WriteRequest, 0, len(items))
	for _, item := range items {
		writeRequests = append(writeRequests, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#" + item.UserID},
					"SK": &types.AttributeValueMemberS{Value: "CART#" + item.ProductID},
				},
			},
		})
	}

	// BatchWriteItemを実行
	_, err = r.db.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			*r.db.Table(): writeRequests,
		},
	})

	return err
}

// recordToCartItem はDynamoDBレコードをドメインモデルに変換する
func recordToCartItem(r *cartRecord) *domain.CartItem {
	return &domain.CartItem{
		UserID:      r.UserID,
		ProductID:   r.ProductID,
		ProductName: r.ProductName,
		Price:       r.Price,
		Quantity:    r.Quantity,
		Version:     r.Version,
		AddedAt:     timeutil.ParseTime(r.AddedAt),
		UpdatedAt:   timeutil.ParseTime(r.UpdatedAt),
	}
}
