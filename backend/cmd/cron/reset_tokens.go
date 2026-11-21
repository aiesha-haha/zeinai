package main

import (
	"context"
	"log"
	"time"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/repository"
	"zeinai-backend/internal/service"
	"zeinai-backend/platform/database"
)

func main() {
	log.Println("Starting monthly token reset cron job...")

	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Connect to Database
	dbPool := database.ConnectDB(*cfg)
	defer database.CloseDB()

	// 3. Initialize Repository and Service
	subscriptionRepo := repository.NewSubscriptionRepository(dbPool)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)

	// 4. Reset tokens for all active subscriptions
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := subscriptionService.ResetMonthlyTokens(ctx)
	if err != nil {
		log.Fatalf("Failed to reset monthly tokens: %v", err)
	}

	log.Println("✅ Monthly token reset completed successfully")
}
