package main

import (
	"context"
	"log"
	"os"

	"quant-agent/internal/api"
	"quant-agent/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env
	_ = godotenv.Load()

	dataDir := getEnv("DATA_DIR", "./data")
	llmURL := getEnv("LLM_URL", "https://api.openai.com/v1")
	llmToken := getEnv("LLM_TOKEN", "")
	connStr := getEnv("DATABASE_URL", "postgres://postgres:postgres@47.99.163.232:5432/quant_agent?sslmode=disable")

	// 连接数据库
	ctx := context.Background()
	s, err := store.NewStore(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer s.Close()

	server := api.NewServer(dataDir, llmURL, llmToken, s)
	server.Run(":" + getEnv("SERVER_ADDR", "8080"))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
