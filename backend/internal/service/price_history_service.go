package service

import (
	"context"
	"time"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/repository"
)

type PriceHistoryService struct {
	priceHistoryRepo *repository.PriceHistoryRepository
	productRepo      *repository.ProductRepository
}

func NewPriceHistoryService(priceHistoryRepo *repository.PriceHistoryRepository, productRepo *repository.ProductRepository) *PriceHistoryService {
	return &PriceHistoryService{
		priceHistoryRepo: priceHistoryRepo,
		productRepo:      productRepo,
	}
}

// UpdatePriceは商品価格を更新し、価格履歴を記録する
func (s *PriceHistoryService) UpdatePrice(ctx context.Context, productID string, newPrice int, changedBy string) error {
	// 商品の現在の価格を取得
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// 価格が変わらない場合は何もしない
	if product.Price == newPrice {
		return nil
	}

	// 価格履歴を記録
	history := &domain.PriceHistory{
		ProductID: productID,
		Price:     newPrice,
		ChangedBy: changedBy,
	}

	if err := s.priceHistoryRepo.Create(ctx, history); err != nil {
		return err
	}

	// 商品価格を更新
	product.Price = newPrice
	return s.productRepo.Update(ctx, product)
}

// GetHistoryは価格履歴を取得する
func (s *PriceHistoryService) GetHistory(ctx context.Context, productID string, limit int32) ([]*domain.PriceHistory, error) {
	return s.priceHistoryRepo.GetByProductID(ctx, productID, limit)
}

// GetHistoryWithRangeは指定期間の価格履歴を取得する
func (s *PriceHistoryService) GetHistoryWithRange(ctx context.Context, productID string, startTime, endTime time.Time) ([]*domain.PriceHistory, error) {
	return s.priceHistoryRepo.GetByProductIDWithRange(ctx, productID, startTime, endTime)
}
