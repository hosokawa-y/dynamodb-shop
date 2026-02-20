package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/response"
)

// InventoryService は在庫管理関連のビジネスロジックを定義するインターフェース
type InventoryService interface {
	AdjustStock(ctx context.Context, productID string, changeType string, quantity int, reason string) error
	GetLogs(ctx context.Context, productID string, limit int32) ([]*domain.InventoryLog, error)
	GetLogsWithRange(ctx context.Context, productID string, startTime, endTime time.Time) ([]*domain.InventoryLog, error)
}

type InventoryHandler struct {
	inventoryService InventoryService
}

func NewInventoryHandler(inventoryService InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// AdjustStockRequest は在庫調整リクエストの構造体
type AdjustStockRequest struct {
	ChangeType string `json:"changeType"` // IN, OUT, ADJUST
	Quantity   int    `json:"quantity"`
	Reason     string `json:"reason"`
}

// AdjustStock は商品の在庫を調整する
// PUT /api/v1/products/{id}/stock
func (h *InventoryHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if productID == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var req AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// バリデーション
	if req.ChangeType != "IN" && req.ChangeType != "OUT" && req.ChangeType != "ADJUST" {
		response.Error(w, http.StatusBadRequest, "ChangeType must be IN, OUT, or ADJUST")
		return
	}

	if req.Quantity < 0 {
		response.Error(w, http.StatusBadRequest, "Quantity must be non-negative")
		return
	}

	if req.Reason == "" {
		response.Error(w, http.StatusBadRequest, "Reason is required")
		return
	}

	if err := h.inventoryService.AdjustStock(r.Context(), productID, req.ChangeType, req.Quantity, req.Reason); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to adjust stock")
		return
	}

	response.Success(w, http.StatusOK, "Stock adjusted successfully")
}

// GetLogs は商品の在庫変動履歴を取得する
// GET /api/v1/products/{id}/inventory-logs?limit=50&start=2025-01-01&end=2025-12-31
func (h *InventoryHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
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

	// 期間指定がある場合はGetLogsWithRangeを使用
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

		logs, err := h.inventoryService.GetLogsWithRange(r.Context(), productID, startTime, endTime)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "Failed to fetch inventory logs")
			return
		}
		response.JSON(w, http.StatusOK, logs)
		return
	}

	// 期間指定がない場合はlimit件数取得
	logs, err := h.inventoryService.GetLogs(r.Context(), productID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch inventory logs")
		return
	}

	response.JSON(w, http.StatusOK, logs)
}

// GetAllLogs は全商品の在庫変動履歴を取得する（管理者用）
// GET /api/v1/admin/inventory-logs?productId=xxx&limit=50
func (h *InventoryHandler) GetAllLogs(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("productId")
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

	logs, err := h.inventoryService.GetLogs(r.Context(), productID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch inventory logs")
		return
	}

	response.JSON(w, http.StatusOK, logs)
}
