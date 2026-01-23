package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/response"
)

// ProductService は商品関連のビジネスロジックを定義するインターフェース
type ProductService interface {
	List(ctx context.Context, category string) ([]*domain.Product, error)
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	Create(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error)
	Update(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error)
	Delete(ctx context.Context, id string) error
}

type ProductHandler struct {
	productService ProductService
}

func NewProductHandler(productService ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// List は商品一覧を取得する
// GET /api/v1/products?category=xxx
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	products, err := h.productService.List(r.Context(), category)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	response.JSON(w, http.StatusOK, products)
}

// GetByID は指定IDの商品を取得する
// GET /api/v1/products/{id}
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	product, err := h.productService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Product not found")
		return
	}

	response.JSON(w, http.StatusOK, product)
}

// Create は新規商品を作成する
// POST /api/v1/products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Price <= 0 {
		response.Error(w, http.StatusBadRequest, "Name and positive price are required")
		return
	}

	product, err := h.productService.Create(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	response.JSON(w, http.StatusCreated, product)
}

// Update は商品情報を更新する
// PUT /api/v1/products/{id}
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var req domain.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	product, err := h.productService.Update(r.Context(), id, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, product)
}

// Delete は商品を削除する
// DELETE /api/v1/products/{id}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	if err := h.productService.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	response.Success(w, http.StatusOK, "Product deleted successfully")
}
