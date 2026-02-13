// order_repo.go
// 注文データのDynamoDB操作を担当するリポジトリ
//
// 【トランザクションとは】
//
//	複数の書き込み操作を「全て成功」または「全て失敗」で実行する仕組み
//	→ 注文確定では以下を1つのトランザクションで実行:
//	  1. 注文ヘッダー作成（Put）
//	  2. 注文明細作成（Put × 商品数）
//	  3. 在庫減算（Update × 商品数）条件付き
//	  4. カートクリア（Delete × 商品数）
//
// 【キー設計】
//
//	注文ヘッダー: PK=USER#<userId>, SK=ORDER#<orderId>
//	注文明細:     PK=ORDER#<orderId>, SK=ITEM#<productId>
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
	"github.com/google/uuid"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/timeutil"
)

// トランザクションエラー
var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrTransactionConflict = errors.New("transaction conflict: please retry")
)

type orderRecord struct {
	PK          string `dynamodbav:"PK"`     // USER#<userId>
	SK          string `dynamodbav:"SK"`     // ORDER#<orderId>
	GSI1PK      string `dynamodbav:"GSI1PK"` // ORDERS#<yyyy-mm>（月別検索用）
	GSI1SK      string `dynamodbav:"GSI1SK"` // <timestamp>#<orderId>
	OrderID     string `dynamodbav:"orderId"`
	UserID      string `dynamodbav:"userId"`
	Status      string `dynamodbav:"status"`
	TotalAmount int    `dynamodbav:"totalAmount"`
	ItemCount   int    `dynamodbav:"itemCount"`
	CreatedAt   string `dynamodbav:"createdAt"`
	UpdatedAt   string `dynamodbav:"updatedAt"`
}

type orderItemRecord struct {
	PK          string `dynamodbav:"PK"` // ORDER#<orderId>
	SK          string `dynamodbav:"SK"` // ITEM#<productId>
	OrderID     string `dynamodbav:"orderId"`
	ProductID   string `dynamodbav:"productId"`
	ProductName string `dynamodbav:"productName"`
	Price       int    `dynamodbav:"price"`
	Quantity    int    `dynamodbav:"quantity"`
	Subtotal    int    `dynamodbav:"subtotal"`
}

type OrderRepository struct {
	db *DynamoDBClient
}

func NewOrderRepository(db *DynamoDBClient) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder は注文を確定する（トランザクション）
// 【使用API】TransactWriteItems
//
// 【TransactWriteItemsの特徴】
//   - 最大100件の書き込み操作を1つのトランザクションで実行
//   - 全て成功 or 全て失敗（ACID特性）
//   - Put, Update, Delete, ConditionCheck を組み合わせ可能
//   - 各操作に ConditionExpression を設定可能
//
// 【実行する操作】
//  1. Put: 注文ヘッダー
//  2. Put: 注文明細（商品数分）
//  3. Update: 商品の在庫減算（条件: Stock >= 購入数量）
//  4. Delete: カートアイテム（商品数分）
func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order, items []domain.OrderItem, cartItems []domain.CartItem) error {
	now := time.Now()
	orderID := uuid.New().String()
	order.ID = orderID
	order.CreatedAt = now
	order.UpdatedAt = now

	// トランザクションアイテムを構築
	transactionItems := make([]types.TransactWriteItem, 0)

	// 1. 注文ヘッダーのPut
	orderRec := orderRecord{
		PK:          "USER#" + order.UserID,
		SK:          "ORDER#" + order.ID,
		GSI1PK:      "ORDERS#" + now.Format("2006-01"),        // 月別検索用
		GSI1SK:      now.Format(time.RFC3339) + "#" + orderID, // タイムスタンプ順
		OrderID:     orderID,
		UserID:      order.UserID,
		Status:      domain.OrderStatusConfirmed,
		TotalAmount: order.TotalAmount,
		ItemCount:   order.ItemCount,
		CreatedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}
	orderAV, err := attributevalue.MarshalMap(orderRec)
	if err != nil {
		return err
	}
	transactionItems = append(transactionItems, types.TransactWriteItem{
		Put: &types.Put{
			TableName: r.db.Table(),
			Item:      orderAV,
		},
	})

	// 2. 注文明細のPut（商品数分）
	for _, item := range items {
		itemRec := orderItemRecord{
			PK:          "ORDER#" + order.ID,
			SK:          "ITEM#" + item.ProductID,
			OrderID:     orderID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Price * item.Quantity,
		}
		itemAV, err := attributevalue.MarshalMap(itemRec)
		if err != nil {
			return err
		}
		transactionItems = append(transactionItems, types.TransactWriteItem{
			Put: &types.Put{
				TableName: r.db.Table(),
				Item:      itemAV,
			},
		})
	}

	// 3. 在庫減算のUpdate（条件付きUpdate）
	// 【重要】ConditionExpression で在庫チェック
	//   - Stock >= :qty の場合のみ更新を実行
	//   - 在庫不足の場合はトランザクション全体が失敗
	for _, item := range items {
		transactionItems = append(transactionItems, types.TransactWriteItem{
			Update: &types.Update{
				TableName: r.db.Table(),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "PRODUCT#" + item.ProductID},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression: aws.String("SET Stock = Stock - :qty, UpdatedAt = :now"),
				// 【ConditionExpression】在庫が購入数量以上あることを確認
				// この条件を満たさない場合、トランザクション全体がロールバック
				ConditionExpression: aws.String("Stock >= :qty"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":qty": &types.AttributeValueMemberN{Value: strconv.Itoa(item.Quantity)},
					":now": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
				},
			},
		})
	}

	// 4. カートアイテムのDelete（商品数分）
	for _, cartItem := range cartItems {
		transactionItems = append(transactionItems, types.TransactWriteItem{
			Delete: &types.Delete{
				TableName: r.db.Table(),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#" + order.UserID},
					"SK": &types.AttributeValueMemberS{Value: "CART#" + cartItem.ProductID},
				},
			},
		})
	}

	// トランザクション実行
	_, err = r.db.Client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: transactionItems,
	})
	if err != nil {
		// 【エラーハンドリング】
		// TransactionCanceledException: トランザクションがキャンセルされた
		//   - CancellationReasons で各操作の失敗理由を確認可能
		//   - ConditionalCheckFailed: 条件を満たさなかった（在庫不足など）
		//   - TransactionConflict: 別のトランザクションと競合
		var tce *types.TransactionCanceledException
		if errors.As(err, &tce) {
			// 各操作の失敗理由をチェック
			for _, reason := range tce.CancellationReasons {
				if reason.Code != nil {
					switch *reason.Code {
					case "ConditionalCheckFailed":
						return ErrInsufficientStock
					case "TransactionConflict":
						return ErrTransactionConflict
					}
				}
			}
		}
		return err
	}

	return nil
}

// GetByUserIDはユーザーの注文一覧を取得する
func (r *OrderRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk": &types.AttributeValueMemberS{Value: "ORDER#"},
		},
		ScanIndexForward: aws.Bool(false), // 最新注文を先頭に
	})
	if err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, 0, len(result.Items))
	for _, item := range result.Items {
		var rec orderRecord
		if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
			return nil, err
		}
		orders = append(orders, recordToOrder(&rec))
	}

	return orders, nil
}

// GetByIDは注文詳細を取得する
func (r *OrderRepository) GetByID(ctx context.Context, userID, orderID string) (*domain.Order, error) {
	// 注文ヘッダー取得
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: r.db.Table(),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "ORDER#" + orderID},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, ErrOrderNotFound
	}

	var rec orderRecord
	if err := attributevalue.UnmarshalMap(result.Item, &rec); err != nil {
		return nil, err
	}
	order := recordToOrder(&rec)

	// 注文明細取得
	items, err := r.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

// GetOrderItemsは注文明細を取得する
func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              r.db.Table(),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "ORDER#" + orderID},
			":sk": &types.AttributeValueMemberS{Value: "ITEM#"},
		},
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.OrderItem, 0, len(result.Items))
	for _, item := range result.Items {
		var rec orderItemRecord
		if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
			return nil, err
		}
		items = append(items, recordToOrderItem(&rec))
	}

	return items, nil
}

func recordToOrder(r *orderRecord) *domain.Order {
	return &domain.Order{
		ID:          r.OrderID,
		UserID:      r.UserID,
		Status:      r.Status,
		TotalAmount: r.TotalAmount,
		ItemCount:   r.ItemCount,
		CreatedAt:   timeutil.ParseTime(r.CreatedAt),
		UpdatedAt:   timeutil.ParseTime(r.UpdatedAt),
	}
}

func recordToOrderItem(r *orderItemRecord) domain.OrderItem {
	return domain.OrderItem{
		OrderID:     r.OrderID,
		ProductID:   r.ProductID,
		ProductName: r.ProductName,
		Price:       r.Price,
		Quantity:    r.Quantity,
		Subtotal:    r.Subtotal,
	}
}
