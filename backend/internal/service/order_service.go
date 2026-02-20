package service

import (
	"context"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/repository"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// CreateOrder はカートから注文を作成する
// 【処理フロー】
//  1. カートを取得
//  2. カートアイテムを注文明細に変換
//  3. トランザクションで注文確定
//     - 注文ヘッダー作成
//     - 注文明細作成
//     - 在庫減算（条件付き）
//     - カートクリア
func (s *OrderService) CreateOrder(ctx context.Context, userID string) (*domain.Order, error) {
	// 1. カートを取得
	cartItems, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, repository.ErrCartItemNotFound
	}
	// 2. 注文データを構築
	var totalAmount int
	orderItems := make([]domain.OrderItem, 0, len(cartItems))

	for _, cartItem := range cartItems {
		subtotal := cartItem.Price * cartItem.Quantity
		totalAmount += subtotal

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:   cartItem.ProductID,
			ProductName: cartItem.ProductName,
			Price:       cartItem.Price,
			Quantity:    cartItem.Quantity,
			Subtotal:    subtotal,
		})
	}

	order := &domain.Order{
		UserID:      userID,
		Status:      domain.OrderStatusConfirmed,
		TotalAmount: totalAmount,
		ItemCount:   len(orderItems),
	}

	// cartItemsをポインタスライスから値スライスに変換
	cartItemValues := make([]domain.CartItem, len(cartItems))
	for i, item := range cartItems {
		cartItemValues[i] = *item
	}

	// 3. トランザクションで注文確定
	// → 注文作成・在庫減算・カート削除を一括実行
	err = s.orderRepo.CreateOrder(ctx, order, orderItems, cartItemValues)
	if err != nil {
		// エラーの種類に応じたハンドリングはハンドラー層で行う
		return nil, err
	}

	order.Items = orderItems

	return order, nil
}

// GetOrdersはユーザーの注文一覧を取得する
func (s *OrderService) GetOrders(ctx context.Context, userID string) ([]*domain.Order, error) {
	return s.orderRepo.GetByUserID(ctx, userID)
}

// GetOrderByIDは注文詳細を取得する
func (s *OrderService) GetOrderByID(ctx context.Context, userID, orderID string) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, userID, orderID)
}
