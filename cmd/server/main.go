package main

import (
	"os"

	"quant-agent/internal/api"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env
	_ = godotenv.Load()

	addr := getEnv("SERVER_ADDR", ":8080")
	dataDir := getEnv("DATA_DIR", "./data")
	llmURL := getEnv("LLM_URL", "https://api.openai.com/v1")
	llmToken := getEnv("LLM_TOKEN", "")

	server := api.NewServer(dataDir, llmURL, llmToken)
	server.Run(addr)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
