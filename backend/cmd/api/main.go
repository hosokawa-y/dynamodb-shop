package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/config"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/handler"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/middleware"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/repository"
	"github.com/hosokawa-y/dynamodb-shop/backend/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルの読み込み（存在する場合）
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// 設定の読み込み
	cfg := config.Load()

	// DynamoDBクライアントの初期化
	ctx := context.Background()
	dbClient, err := repository.NewDynamoDBClient(ctx, cfg.DynamoDBTable)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	// JWT認証の初期化
	jwtExpiry, err := time.ParseDuration(cfg.JWTExpiry)
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}
	jwtAuth := middleware.NewJWTAuth(cfg.JWTSecret, jwtExpiry)

	// Repository の初期化
	userRepo := repository.NewUserRepository(dbClient)
	productRepo := repository.NewProductRepository(dbClient)
	cartRepo := repository.NewCartRepository(dbClient)
	orderRepo := repository.NewOrderRepository(dbClient)
	priceHistoryRepo := repository.NewPriceHistoryRepository(dbClient)
	inventoryRepo := repository.NewInventoryRepository(dbClient)

	// Service の初期化
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo)
	priceHistoryService := service.NewPriceHistoryService(priceHistoryRepo, productRepo)
	inventoryService := service.NewInventoryService(inventoryRepo, productRepo)

	// Handler の初期化
	authHandler := handler.NewAuthHandler(userService, jwtAuth)
	productHandler := handler.NewProductHandler(productService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)
	priceHistoryHandler := handler.NewPriceHistoryHandler(priceHistoryService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	// Router の設定
	router := handler.NewRouter(jwtAuth, authHandler, productHandler, cartHandler, orderHandler, priceHistoryHandler, inventoryHandler)
	httpHandler := router.Setup()

	// サーバーの設定
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// サーバー起動
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
