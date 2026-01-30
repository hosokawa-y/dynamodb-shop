// product_repo.go
// 商品データのDynamoDB操作を担当するリポジトリ
//
// 【キー設計】
//   PK:     PRODUCT#<商品ID>    - パーティションキー（商品単位）
//   SK:     METADATA            - ソートキー（固定値）
//   GSI1PK: PRODUCT             - 全商品を同じパーティションにまとめる
//   GSI1SK: CATEGORY#<カテゴリ>#<商品ID> - カテゴリ検索用
//
// 【アクセスパターン】
//   1. 商品ID指定で取得     → GetItem(PK, SK)
//   2. 全商品一覧          → Query(GSI1PK = "PRODUCT")
//   3. カテゴリ別商品一覧   → Query(GSI1PK = "PRODUCT" AND begins_with(GSI1SK, "CATEGORY#xxx"))

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

// productRecord はDynamoDBに保存する商品データの構造体
// dynamodbavタグでDynamoDBの属性名を指定
type productRecord struct {
	PK          string `dynamodbav:"PK"`        // パーティションキー: PRODUCT#<id>
	SK          string `dynamodbav:"SK"`        // ソートキー: METADATA
	GSI1PK      string `dynamodbav:"GSI1PK"`    // GSI1パーティションキー: PRODUCT
	GSI1SK      string `dynamodbav:"GSI1SK"`    // GSI1ソートキー: CATEGORY#<category>#<id>
	ID          string `dynamodbav:"id"`
	Name        string `dynamodbav:"name"`
	Description string `dynamodbav:"description"`
	Price       int    `dynamodbav:"price"`
	Category    string `dynamodbav:"category"`
	Stock       int    `dynamodbav:"stock"`
	ImageURL    string `dynamodbav:"imageUrl"`
	CreatedAt   string `dynamodbav:"createdAt"`
	UpdatedAt   string `dynamodbav:"updatedAt"`
}

// ProductRepository は商品のDynamoDB操作を提供する
type ProductRepository struct {
	db *DynamoDBClient
}

// NewProductRepository は ProductRepository のインスタンスを生成する
func NewProductRepository(db *DynamoDBClient) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// Create は新規商品をDynamoDBに保存する
// 【使用API】PutItem - 新規アイテムの作成（または上書き）
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	now := time.Now()
	product.ID = uuid.New().String()
	product.CreatedAt = now
	product.UpdatedAt = now

	// DynamoDB用のレコード構造体を作成
	// GSI1SK の形式: CATEGORY#electronics#uuid
	// → begins_with で "CATEGORY#electronics" を指定するとそのカテゴリの商品だけ取得できる
	record := productRecord{
		PK:          "PRODUCT#" + product.ID,
		SK:          "METADATA",
		GSI1PK:      "PRODUCT",                                          // 全商品で共通
		GSI1SK:      "CATEGORY#" + product.Category + "#" + product.ID,  // カテゴリ検索用
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}

	// Go構造体 → DynamoDB AttributeValue に変換
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}

	// PutItem: アイテムを作成（同じキーが存在する場合は上書き）
	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.db.Table(),
		Item:      item,
	})

	return err
}

// GetByID は商品IDを指定して1件取得する
// 【使用API】GetItem - PK+SKを指定して1件取得（最も高速）
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	// GetItem: パーティションキー + ソートキーを指定して取得
	// 特徴: 1件のみ取得、最も低レイテンシー、読み込みキャパシティ消費が最小
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

	// アイテムが見つからない場合
	if result.Item == nil {
		return nil, ErrProductNotFound
	}

	// DynamoDB AttributeValue → Go構造体 に変換
	var record productRecord
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return nil, err
	}

	return recordToProduct(&record), nil
}

// List は商品一覧を取得する（カテゴリ指定可能）
// 【使用API】Query - GSI1を使用した一覧取得
//
// 【GSI1の構造】
//   GSI1PK: "PRODUCT"（全商品で共通 = 同じパーティションに配置）
//   GSI1SK: "CATEGORY#electronics#001" のような形式
//
// 【begins_with の動作】
//   begins_with(GSI1SK, "CATEGORY#electronics") は以下にマッチ:
//     ✅ CATEGORY#electronics#001
//     ✅ CATEGORY#electronics#002
//     ❌ CATEGORY#clothing#003
func (r *ProductRepository) List(ctx context.Context, category string) ([]*domain.Product, error) {
	var input *dynamodb.QueryInput

	if category != "" {
		// ========================================
		// カテゴリ指定あり: begins_with でフィルタ
		// ========================================
		// KeyConditionExpression で GSI1PK と GSI1SK の両方を条件に含める
		// begins_with は前方一致検索（プレフィックス検索）
		input = &dynamodb.QueryInput{
			TableName:              r.db.Table(),
			IndexName:              aws.String("GSI1"),                                         // GSI1インデックスを使用
			KeyConditionExpression: aws.String("GSI1PK = :pk AND begins_with(GSI1SK, :sk)"),   // キー条件式
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "PRODUCT"},              // 全商品
				":sk": &types.AttributeValueMemberS{Value: "CATEGORY#" + category}, // カテゴリプレフィックス
			},
		}
	} else {
		// ========================================
		// 全商品取得: GSI1PK のみで検索
		// ========================================
		// GSI1PK = "PRODUCT" の全レコードを取得
		input = &dynamodb.QueryInput{
			TableName:              r.db.Table(),
			IndexName:              aws.String("GSI1"),
			KeyConditionExpression: aws.String("GSI1PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "PRODUCT"},
			},
		}
	}

	// Query実行
	// 特徴: パーティション内の複数アイテムを効率的に取得
	// Scanと違い、パーティションキーを指定するので無駄な読み込みが発生しない
	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	// 結果をドメインモデルに変換
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

// Update は既存商品を更新する
// 【使用API】PutItem + ConditionExpression
//
// 【ConditionExpression の役割】
//   attribute_exists(PK) = PKが存在する場合のみ実行
//   → 存在しないアイテムへの誤った更新を防ぐ
//   → 条件を満たさない場合は ConditionalCheckFailedException が発生
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	now := time.Now()

	record := productRecord{
		PK:          "PRODUCT#" + product.ID,
		SK:          "METADATA",
		GSI1PK:      "PRODUCT",
		GSI1SK:      "CATEGORY#" + product.Category + "#" + product.ID,
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}

	// PutItem + ConditionExpression で条件付き更新
	// attribute_exists(PK): 既存アイテムが存在する場合のみ更新を許可
	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           r.db.Table(),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})

	return err
}

// Delete は商品を削除する
// 【使用API】DeleteItem + ConditionExpression
//
// 【注意】DynamoDBのDeleteItemは存在しないキーを指定してもエラーにならない
//   → ConditionExpression で存在チェックを追加することで、
//     存在しない場合にエラーを返すようにしている
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	// DeleteItem: PK+SKを指定して削除
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

// recordToProduct はDynamoDBレコードをドメインモデルに変換する
// PK, SK, GSI1PK, GSI1SK はDynamoDB専用の属性なので、ドメインモデルには含めない
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