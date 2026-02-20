package service

import (
	"context"
	"time"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/repository"
)

type InventoryService struct {
	inventoryRepo *repository.InventoryRepository
	productRepo   *repository.ProductRepository
}

func NewInventoryService(inventoryRepo *repository.InventoryRepository, productRepo *repository.ProductRepository) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
	}
}

// AdjustStock は在庫を調整し、変動ログを記録する
// changeType: "IN" (入庫), "OUT" (出庫), "ADJUST" (調整)
func (s *InventoryService) AdjustStock(ctx context.Context, productID string, changeType string, quantity int, reason string) error {
	// 現在の商品情報を取得
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	previousStock := product.Stock
	var newStock int

	// 在庫数を計算
	switch changeType {
	case "IN":
		newStock = previousStock + quantity
	case "OUT":
		newStock = previousStock - quantity
		if newStock < 0 {
			newStock = 0 // 在庫は0未満にならないようにする
		}
	case "ADJUST":
		// ADJUSTの場合、quantityは絶対値（新しい在庫数）
		newStock = quantity
	default:
		newStock = previousStock // 変更なし
	}

	// 在庫変動ログを記録
	log := &domain.InventoryLog{
		ProductID:     productID,
		ChangeType:    changeType,
		Quantity:      quantity,
		PreviousStock: previousStock,
		NewStock:      newStock,
		Reason:        reason,
	}

	if err := s.inventoryRepo.Create(ctx, log); err != nil {
		return err
	}

	// 商品の在庫数を更新
	product.Stock = newStock
	return s.productRepo.Update(ctx, product)
}

// GetLogsは在庫変動履歴を取得する
func (s *InventoryService) GetLogs(ctx context.Context, productID string, limit int32) ([]*domain.InventoryLog, error) {
	return s.inventoryRepo.GetByProductID(ctx, productID, limit)
}

// GetLogsWithRangeは指定期間の在庫変動履歴を取得する
func (s *InventoryService) GetLogsWithRange(ctx context.Context, productID string, startTime, endTime time.Time) ([]*domain.InventoryLog, error) {
	return s.inventoryRepo.GetByProductIDWithRange(ctx, productID, startTime, endTime)
}
