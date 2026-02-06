package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/middleware"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/service"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/response"
)

// CartService はカート関連のビジネスロジックを定義するインターフェース
type CartService interface {
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	AddItem(ctx context.Context, userID string, req *domain.AddToCartRequest) (*domain.CartItem, error)
	UpdateQuantity(ctx context.Context, userID, productID string, req *domain.UpdateCartRequest) (*domain.CartItem, error)
	RemoveItem(ctx context.Context, userID, productID string) error
}

type CartHandler struct {
	cartService CartService
}

func NewCartHandler(cartService CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// GetCart はユーザーのカートを取得する
// GET /api/v1/cart
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	cart, err := h.cartService.GetCart(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch cart")
		return
	}

	response.JSON(w, http.StatusOK, cart)
}

// AddItem はカートにアイテムを追加する
// POST /api/v1/cart/items
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req domain.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ProductID == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	if req.Quantity <= 0 {
		response.Error(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}

	item, err := h.cartService.AddItem(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientStock) {
			response.Error(w, http.StatusBadRequest, "Insufficient stock")
			return
		}
		if errors.Is(err, service.ErrInvalidQuantity) {
			response.Error(w, http.StatusBadRequest, "Invalid quantity")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to add item to cart")
		return
	}

	response.JSON(w, http.StatusCreated, item)
}

// UpdateQuantity はカートアイテムの数量を更新する
// PUT /api/v1/cart/items/{productId}
func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	productID := r.PathValue("productId")
	if productID == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var req domain.UpdateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Quantity <= 0 {
		response.Error(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}

	item, err := h.cartService.UpdateQuantity(r.Context(), userID, productID, &req)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientStock) {
			response.Error(w, http.StatusBadRequest, "Insufficient stock")
			return
		}
		if errors.Is(err, service.ErrInvalidQuantity) {
			response.Error(w, http.StatusBadRequest, "Invalid quantity")
			return
		}
		if errors.Is(err, service.ErrOptimisticLockRetry) {
			response.Error(w, http.StatusConflict, "Failed to update due to concurrent modifications, please retry")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to update cart item")
		return
	}

	response.JSON(w, http.StatusOK, item)
}

// RemoveItem はカートからアイテムを削除する
// DELETE /api/v1/cart/items/{productId}
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	productID := r.PathValue("productId")
	if productID == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	if err := h.cartService.RemoveItem(r.Context(), userID, productID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to remove item from cart")
		return
	}

	response.Success(w, http.StatusOK, "Item removed from cart")
}
