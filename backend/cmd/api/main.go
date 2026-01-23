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
	// TODO: Service パッケージを作成後にインポートを追加
	// "github.com/hosokawa-y/dynamodb-shop/backend/internal/service"
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
	dbClient, err := repository.NewDynamoDBClient(ctx, cfg.DynamoDBTable, cfg.DynamoDBEndpoint, cfg.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	// JWT認証の初期化
	jwtExpiry, err := time.ParseDuration(cfg.JWTExpiry)
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}
	jwtAuth := middleware.NewJWTAuth(cfg.JWTSecret, jwtExpiry)

	// TODO: Repository の初期化（ユーザーが実装）
	// userRepo := repository.NewUserRepository(dbClient)
	// productRepo := repository.NewProductRepository(dbClient)

	// TODO: Service の初期化（ユーザーが実装）
	// userService := service.NewUserService(userRepo)
	// productService := service.NewProductService(productRepo)

	// Handler の初期化
	// TODO: 実際の Service を渡すように変更
	// authHandler := handler.NewAuthHandler(userService, jwtAuth)
	// productHandler := handler.NewProductHandler(productService)
	_ = dbClient    // Repository実装後に削除
	_ = jwtAuth     // Handler初期化後に削除

	// TODO: Handler初期化後にコメントを解除
	// router := handler.NewRouter(jwtAuth, authHandler, productHandler)
	// httpHandler := router.Setup()

	// 仮のルーター（Service実装前の動作確認用）
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"Waiting for Service implementation"}`))
	})
	httpHandler := middleware.Logging(middleware.CORS(mux))

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
