package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/middleware"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/response"
)

// PriceHistoryService は価格履歴関連のビジネスロジックを定義するインターフェース
type PriceHistoryService interface {
	UpdatePrice(ctx context.Context, productID string, newPrice int, changedBy string) error
	GetHistory(ctx context.Context, productID string, limit int32) ([]*domain.PriceHistory, error)
	GetHistoryWithRange(ctx context.Context, productID string, startTime, endTime time.Time) ([]*domain.PriceHistory, error)
}

type PriceHistoryHandler struct {
	priceHistoryService PriceHistoryService
}

func NewPriceHistoryHandler(priceHistoryService PriceHistoryService) *PriceHistoryHandler {
	return &PriceHistoryHandler{
		priceHistoryService: priceHistoryService,
	}
}

// UpdatePriceRequest は価格更新リクエストの構造体
type UpdatePriceRequest struct {
	Price int `json:"price"`
}

// UpdatePrice は商品の価格を更新する
// PUT /api/v1/products/{id}/price
func (h *PriceHistoryHandler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if productID == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var req UpdatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Price <= 0 {
		response.Error(w, http.StatusBadRequest, "Price must be positive")
		return
	}

	// JWTからユーザーIDを取得
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		userID = "unknown"
	}

	if err := h.priceHistoryService.UpdatePrice(r.Context(), productID, req.Price, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update price")
		return
	}

	response.Success(w, http.StatusOK, "Price updated successfully")
}

// GetHistory は商品の価格履歴を取得する
// GET /api/v1/products/{id}/price-history?limit=50&start=2025-01-01&end=2025-12-31
func (h *PriceHistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if productID == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	// クエリパラメータからlimitを取得（デフォルト50）
	limitStr := r.URL.Query().Get("limit")
	limit := int32(50)
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	// 期間指定がある場合はGetHistoryWithRangeを使用
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr != "" && endStr != "" {
		startTime, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid start date format (use YYYY-MM-DD)")
			return
		}
		endTime, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid end date format (use YYYY-MM-DD)")
			return
		}
		// 終了日は23:59:59まで含める
		endTime = endTime.Add(24*time.Hour - time.Second)

		histories, err := h.priceHistoryService.GetHistoryWithRange(r.Context(), productID, startTime, endTime)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "Failed to fetch price history")
			return
		}
		response.JSON(w, http.StatusOK, histories)
		return
	}

	// 期間指定がない場合はlimit件数取得
	histories, err := h.priceHistoryService.GetHistory(r.Context(), productID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch price history")
		return
	}

	response.JSON(w, http.StatusOK, histories)
}
