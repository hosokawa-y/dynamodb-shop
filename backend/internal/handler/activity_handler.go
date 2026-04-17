package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/middleware"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/service"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/response"
)

// ActivityService はユーザー行動ログ関連のビジネスロジックを定義するインターフェース
type ActivityService interface {
	LogActivity(ctx context.Context, userID string, req *domain.LogActivityRequest) error
	LogActivities(ctx context.Context, userID string, reqs []*domain.LogActivityRequest) error
	GetUserActivities(ctx context.Context, userID string, limit int32) ([]*domain.UserActivity, error)
	GetUserActivitiesByAction(ctx context.Context, userID string, actionType string, limit int32) ([]*domain.UserActivity, error)
}

type ActivityHandler struct {
	activityService ActivityService
}

func NewActivityHandler(activityService ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

// LogActivity は行動ログを1件記録する
// POST /api/v1/activity
func (h *ActivityHandler) LogActivity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req domain.LogActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ActionType == "" {
		response.Error(w, http.StatusBadRequest, "Action type is required")
		return
	}

	if err := h.activityService.LogActivity(r.Context(), userID, &req); err != nil {
		if errors.Is(err, service.ErrInvalidActionType) {
			response.Error(w, http.StatusBadRequest, "Invalid action type")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to log activity")
		return
	}

	response.Success(w, http.StatusCreated, "Activity logged")
}

// BatchLogActivities は複数の行動ログを一括記録する
// POST /api/v1/activity/batch
func (h *ActivityHandler) BatchLogActivities(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var reqs []*domain.LogActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(reqs) == 0 {
		response.Error(w, http.StatusBadRequest, "At least one activity is required")
		return
	}

	if err := h.activityService.LogActivities(r.Context(), userID, reqs); err != nil {
		if errors.Is(err, service.ErrInvalidActionType) {
			response.Error(w, http.StatusBadRequest, "Invalid action type")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to log activities")
		return
	}

	response.Success(w, http.StatusCreated, "Activities logged")
}

// GetMyActivities は現在のユーザーの行動ログを取得する
// GET /api/v1/activity
func (h *ActivityHandler) GetMyActivities(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// クエリパラメータの解析
	limit := int32(50)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	actionType := r.URL.Query().Get("actionType")

	var activities []*domain.UserActivity
	var err error

	if actionType != "" {
		activities, err = h.activityService.GetUserActivitiesByAction(r.Context(), userID, actionType, limit)
	} else {
		activities, err = h.activityService.GetUserActivities(r.Context(), userID, limit)
	}

	if err != nil {
		if errors.Is(err, service.ErrInvalidActionType) {
			response.Error(w, http.StatusBadRequest, "Invalid action type")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	response.JSON(w, http.StatusOK, activities)
}

// GetUserActivities は管理者が特定ユーザーの行動ログを取得する
// GET /api/v1/admin/users/{userId}/activities
func (h *ActivityHandler) GetUserActivities(w http.ResponseWriter, r *http.Request) {
	// 認証チェック（管理者権限は省略）
	if middleware.GetUserID(r.Context()) == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		response.Error(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// クエリパラメータの解析
	limit := int32(50)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	actionType := r.URL.Query().Get("actionType")

	var activities []*domain.UserActivity
	var err error

	if actionType != "" {
		activities, err = h.activityService.GetUserActivitiesByAction(r.Context(), targetUserID, actionType, limit)
	} else {
		activities, err = h.activityService.GetUserActivities(r.Context(), targetUserID, limit)
	}

	if err != nil {
		if errors.Is(err, service.ErrInvalidActionType) {
			response.Error(w, http.StatusBadRequest, "Invalid action type")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	response.JSON(w, http.StatusOK, activities)
}
