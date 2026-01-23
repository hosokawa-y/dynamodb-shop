package handler

import (
	"net/http"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/middleware"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/response"
)

type Router struct {
	mux            *http.ServeMux
	jwtAuth        *middleware.JWTAuth
	authHandler    *AuthHandler
	productHandler *ProductHandler
}

func NewRouter(jwtAuth *middleware.JWTAuth, authHandler *AuthHandler, productHandler *ProductHandler) *Router {
	return &Router{
		mux:            http.NewServeMux(),
		jwtAuth:        jwtAuth,
		authHandler:    authHandler,
		productHandler: productHandler,
	}
}

func (r *Router) Setup() http.Handler {
	// Health check
	r.mux.HandleFunc("GET /health", func(w http.ResponseWriter, req *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth routes (public)
	r.mux.HandleFunc("POST /api/v1/auth/register", r.authHandler.Register)
	r.mux.HandleFunc("POST /api/v1/auth/login", r.authHandler.Login)

	// Auth routes (protected)
	r.mux.Handle("GET /api/v1/auth/profile", r.jwtAuth.Middleware(http.HandlerFunc(r.authHandler.GetProfile)))

	// Product routes (public)
	r.mux.HandleFunc("GET /api/v1/products", r.productHandler.List)
	r.mux.HandleFunc("GET /api/v1/products/{id}", r.productHandler.GetByID)

	// Product routes (protected - admin only in real app)
	r.mux.Handle("POST /api/v1/products", r.jwtAuth.Middleware(http.HandlerFunc(r.productHandler.Create)))
	r.mux.Handle("PUT /api/v1/products/{id}", r.jwtAuth.Middleware(http.HandlerFunc(r.productHandler.Update)))
	r.mux.Handle("DELETE /api/v1/products/{id}", r.jwtAuth.Middleware(http.HandlerFunc(r.productHandler.Delete)))

	// Apply middleware
	handler := middleware.Logging(middleware.CORS(r.mux))

	return handler
}
