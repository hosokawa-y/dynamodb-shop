package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/middleware"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/response"
)

// UserService はユーザー関連のビジネスロジックを定義するインターフェース
type UserService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
}

type AuthHandler struct {
	userService UserService
	jwtAuth     *middleware.JWTAuth
}

func NewAuthHandler(userService UserService, jwtAuth *middleware.JWTAuth) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtAuth:     jwtAuth,
	}
}

// Register は新規ユーザー登録を処理する
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "Email, password, and name are required")
		return
	}

	user, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	token, err := h.jwtAuth.GenerateToken(user.ID, user.Email)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response.JSON(w, http.StatusCreated, domain.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login はユーザーログインを処理する
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := h.userService.Login(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := h.jwtAuth.GenerateToken(user.ID, user.Email)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response.JSON(w, http.StatusOK, domain.AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile は現在ログイン中のユーザー情報を取得する
// GET /api/v1/auth/profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	response.JSON(w, http.StatusOK, user)
}
