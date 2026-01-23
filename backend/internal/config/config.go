package config

import (
	"os"
)

type Config struct {
	AWSRegion       string
	DynamoDBTable   string
	DynamoDBEndpoint string // ローカル開発用
	JWTSecret       string
	JWTExpiry       string
	ServerPort      string
}

func Load() *Config {
	return &Config{
		AWSRegion:        getEnv("AWS_REGION", "ap-northeast-1"),
		DynamoDBTable:    getEnv("DYNAMODB_TABLE", "DynamoDBShop"),
		DynamoDBEndpoint: getEnv("DYNAMODB_ENDPOINT", ""), // 空の場合はAWS実環境
		JWTSecret:        getEnv("JWT_SECRET", "default-secret-change-me"),
		JWTExpiry:        getEnv("JWT_EXPIRY", "24h"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
