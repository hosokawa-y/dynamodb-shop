// backend/internal/service/activity_service.go
// ユーザー行動ログのビジネスロジックを担当するサービス
//
// 【ActionType】
//   VIEW      - 商品閲覧
//   CLICK     - 商品クリック
//   ADD_CART  - カートに追加
//   PURCHASE  - 購入完了

package service

import (
	"context"
	"errors"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/repository"
)

// 有効なActionType
const (
	ActionTypeView     = "VIEW"
	ActionTypeClick    = "CLICK"
	ActionTypeAddCart  = "ADD_CART"
	ActionTypePurchase = "PURCHASE"
)

var validActionTypes = map[string]bool{
	ActionTypeView:     true,
	ActionTypeClick:    true,
	ActionTypeAddCart:  true,
	ActionTypePurchase: true,
}

var (
	ErrInvalidActionType = errors.New("invalid action type")
)

type ActivityService struct {
	activityRepo *repository.ActivityRepository
}

func NewActivityService(activityRepo *repository.ActivityRepository) *ActivityService {
	return &ActivityService{
		activityRepo: activityRepo,
	}
}

// LogActivityは行動ログを1件記録する
func (s *ActivityService) LogActivity(ctx context.Context, userID string, req *domain.LogActivityRequest) error {
	// ActionTypeの検証
	if !validActionTypes[req.ActionType] {
		return ErrInvalidActionType
	}

	activity := &domain.UserActivity{
		UserID:     userID,
		ActionType: req.ActionType,
		ProductID:  req.ProductID,
		Metadata:   req.Metadata,
	}

	return s.activityRepo.Create(ctx, activity)
}

// LogActiviesは複数の行動ログを一括記録する（BatchWriteItem使用)
func (s *ActivityService) LogActivities(ctx context.Context, userID string, reqs []*domain.LogActivityRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	activities := make([]*domain.UserActivity, 0, len(reqs))
	for _, req := range reqs {
		// ActionTypeの検証
		if !validActionTypes[req.ActionType] {
			return ErrInvalidActionType
		}

		activities = append(activities, &domain.UserActivity{
			UserID:     userID,
			ActionType: req.ActionType,
			ProductID:  req.ProductID,
			Metadata:   req.Metadata,
		})
	}

	return s.activityRepo.BatchCreate(ctx, activities)
}

// GetUserActivitiesはユーザーの行動ログを取得する
func (s *ActivityService) GetUserActivities(ctx context.Context, userID string, limit int32) ([]*domain.UserActivity, error) {
	if limit <= 0 {
		limit = 50 // デフォルト値
	}
	return s.activityRepo.GetByUserID(ctx, userID, limit)
}

// GetUserActivitiesByActionは特定アクションタイプの行動ログを取得する
func (s *ActivityService) GetUserActivitiesByAction(ctx context.Context, userID string, actionType string, limit int32) ([]*domain.UserActivity, error) {
	if !validActionTypes[actionType] {
		return nil, ErrInvalidActionType
	}

	if limit <= 0 {
		limit = 50
	}

	return s.activityRepo.GetByUserIDAndAction(ctx, userID, actionType, limit)
}
