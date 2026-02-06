// cart_service.go
// カート機能のビジネスロジックを担当するサービス
//
// 【主な機能】
//   1. GetCart     - カート取得（合計金額計算付き）
//   2. AddItem     - カート追加（在庫チェック付き）
//   3. UpdateQuantity - 数量更新（楽観的ロック + リトライ）
//   4. RemoveItem  - カートからアイテム削除
//
// 【学習ポイント】
//   - 楽観的ロックのリトライロジック
//   - 在庫チェック（条件付き書き込みの前準備）

package service

import (
	"context"
	"errors"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/repository"
)

var (
	ErrInsufficientStock   = errors.New("insufficient stock for the requested item")
	ErrInvalidQuantity     = errors.New("quantity must be greater than 0")
	ErrOptimisticLockRetry = errors.New("failed to update after max retries due to concurrent modifications")
)

const maxRetries = 3

type CartService struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	items, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	cartItems := make([]domain.CartItem, len(items))
	var totalPrice int
	for i, item := range items {
		cartItems[i] = *item
		totalPrice += item.Price * item.Quantity
	}

	return &domain.Cart{
		Items:      cartItems,
		TotalPrice: totalPrice,
		ItemCount:  len(cartItems),
	}, nil
}

// AddItem はカートにアイテムを追加する
// 【在庫チェック】商品の在庫数を確認し、不足している場合はエラー
// 【既存アイテム】既にカートにある場合は数量を加算
func (s *CartService) AddItem(ctx context.Context, userID string, req *domain.AddToCartRequest) (*domain.CartItem, error) {
	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// 商品情報を取得（在庫チェック + 商品名・価格の取得）
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// 既存のカートアイテムを確認
	existingItem, err := s.cartRepo.GetItem(ctx, userID, req.ProductID)
	if err != nil && !errors.Is(err, repository.ErrCartItemNotFound) {
		return nil, err
	}

	// 追加後の合計数量を計算
	totalQuantity := req.Quantity
	if existingItem != nil {
		totalQuantity += existingItem.Quantity
	}

	// 在庫チェック
	// 【学習ポイント】
	// ここでの在庫チェックは「楽観的」なチェック
	// 実際の在庫減算は注文確定時にトランザクション + 条件付き書き込みで行う
	// カート追加時点では在庫を確保しない（ECサイトの一般的なパターン）
	if product.Stock < totalQuantity {
		return nil, ErrInsufficientStock
	}

	if existingItem != nil {
		// 既存アイテムがある場合は数量を更新
		err = s.updateQuantityWithRetry(ctx, userID, req.ProductID, totalQuantity, existingItem.Version)
		if err != nil {
			return nil, err
		}
		// 更新後のアイテムを取得して返却
		return s.cartRepo.GetItem(ctx, userID, req.ProductID)
	}

	// 新規アイテムを追加
	// 商品の価格が変わってもカート内の価格は変わらないようにする
	// 注文確定時に最新価格を使うかどうかはビジネス要件次第
	item := &domain.CartItem{
		UserID:      userID,
		ProductID:   req.ProductID,
		ProductName: product.Name,
		Price:       product.Price, // カート追加時点の価格を保持（非正規化）
		Quantity:    req.Quantity,
	}

	if err := s.cartRepo.Add(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateQuantity はカートアイテムの数量を更新する
// 【楽観的ロック + リトライ】
// 他のリクエストと競合した場合は最新データを取得してリトライ
func (s *CartService) UpdateQuantity(ctx context.Context, userID, productID string, req *domain.UpdateCartRequest) (*domain.CartItem, error) {
	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// 商品の在庫チェック
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product.Stock < req.Quantity {
		return nil, ErrInsufficientStock
	}

	// リトライ付きで更新
	err = s.updateQuantityWithRetry(ctx, userID, productID, req.Quantity, req.Version)
	if err != nil {
		return nil, err
	}

	// 更新後のアイテムを取得して返却
	return s.cartRepo.GetItem(ctx, userID, productID)
}

// updateQuantityWithRetry は楽観的ロックのリトライロジックを実装
// 【リトライの仕組み】
//  1. 指定されたバージョンで更新を試みる
//  2. ErrVersionMismatch（競合）が発生したら最新データを再取得
//  3. 最新バージョンで再度更新を試みる
//  4. maxRetries回まで繰り返す
func (s *CartService) updateQuantityWithRetry(ctx context.Context, userID, productID string, quantity, version int) error {
	currentVersion := version

	for i := 0; i < maxRetries; i++ {
		// クライアントから送られたバージョンで更新を試みる
		err := s.cartRepo.UpdateQuantity(ctx, userID, productID, quantity, currentVersion)
		if err == nil {
			return nil // 更新成功
		}

		// 楽観的ロックによる競合以外のエラー以外はそのまま返す
		if !errors.Is(err, repository.ErrVersionMismatch) {
			return err
		}

		// 競合発生：最新データを取得してリトライ
		// 【学習ポイント】
		// ErrVersionMismatch = 他のリクエストが先に更新した
		// → 最新のバージョンを取得して再試行
		latestItem, err := s.cartRepo.GetItem(ctx, userID, productID)
		if err != nil {
			return err
		}
		currentVersion = latestItem.Version
	}

	return ErrOptimisticLockRetry
}

// RemoveItem はカートからアイテムを削除する
func (s *CartService) RemoveItem(ctx context.Context, userID, productID string) error {
	return s.cartRepo.Delete(ctx, userID, productID)
}

// ClearCart はカート内の全アイテムを削除する
func (s *CartService) ClearCart(ctx context.Context, userID string) error {
	return s.cartRepo.Clear(ctx, userID)
}
