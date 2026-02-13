package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/middleware"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/repository"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/response"
)

// OrderServiceInterface は注文関連のビジネスロジックを定義するインターフェース
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, userID string) (*domain.Order, error)
	GetOrders(ctx context.Context, userID string) ([]*domain.Order, error)
	GetOrderByID(ctx context.Context, userID, orderID string) (*domain.Order, error)
}

type OrderHandler struct {
	orderService OrderServiceInterface
}

func NewOrderHandler(orderService OrderServiceInterface) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder は注文を確定する
// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	order, err := h.orderService.CreateOrder(r.Context(), userID)
	if err != nil {
		// カートが空の場合
		if errors.Is(err, repository.ErrCartItemNotFound) {
			response.Error(w, http.StatusBadRequest, "Cart is empty")
			return
		}
		// 在庫不足の場合
		if errors.Is(err, repository.ErrInsufficientStock) {
			response.Error(w, http.StatusConflict, "Insufficient stock for one or more items")
			return
		}
		// トランザクション競合の場合
		if errors.Is(err, repository.ErrTransactionConflict) {
			response.Error(w, http.StatusConflict, "Transaction conflict, please retry")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to create order")
		return
	}

	response.JSON(w, http.StatusCreated, order)
}

// GetOrders はユーザーの注文一覧を取得する
// GET /api/v1/orders
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	orders, err := h.orderService.GetOrders(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	response.JSON(w, http.StatusOK, orders)
}

// GetOrderByID は注文詳細を取得する
// GET /api/v1/orders/{id}
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		response.Error(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	order, err := h.orderService.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			response.Error(w, http.StatusNotFound, "Order not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch order")
		return
	}

	response.JSON(w, http.StatusOK, order)
}
