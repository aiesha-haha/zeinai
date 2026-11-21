package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/pkg/fcm"
	"zeinai-backend/internal/pkg/webpush"
	"zeinai-backend/internal/repository"
	"zeinai-backend/internal/service"
	"zeinai-backend/internal/worker"
	"zeinai-backend/platform/database"
	"zeinai-backend/platform/rabbitmq"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Initialize Infrastructure
	// Database
	dbPool := database.ConnectDB(*cfg)
	defer database.CloseDB()

	// RabbitMQ
	if err := rabbitmq.InitRabbitMQ(cfg); err != nil {
		log.Fatalf("Failed to init RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	// AI Generator - Using OpenAI as per user request
	aiGen := ai.NewOpenAIGenerator(cfg)

	// Kimi K2 Generator - for presentation generation
	kimiGen := ai.NewKimiGenerator(cfg)

	// Text generation with retry on rate limits and fallback between providers
	textKimi := ai.KimiFirst(cfg, kimiGen, aiGen)
	textOpenAI := ai.OpenAIFirst(cfg, aiGen, kimiGen)

	// 3. Initialize Repositories
	kmjRepo := repository.NewKMJRepository(dbPool)
	aiRepo := repository.NewAIRepository(dbPool)
	presentationRepo := repository.NewPresentationRepository(dbPool)
	articleRepo := repository.NewArticleRepository(dbPool)
	notificationRepo := repository.NewNotificationRepository(dbPool)
	whiteboardRepo := repository.NewWhiteboardRepository(dbPool)
	classRepo := repository.NewClassRepository(dbPool)
	userRepo := repository.NewUserRepository(dbPool)
	auditRepo := repository.NewAuditRepository(dbPool)
	groupTaskRepo := repository.NewGroupTaskRepository(dbPool)
	tugOfWarRepo := repository.NewTugOfWarRepository(dbPool)
	testRepo := repository.NewTestRepository(dbPool)
	studioRepo := repository.NewStudioRepository(dbPool)
	// 3.5. Push Clients
	var fcmClient *fcm.FCMClient
	var webPushClient *webpush.WebPushClient

	if cfg.FCM.CredentialsPath != "" {
		var err error
		fcmClient, err = fcm.NewFCMClient(cfg.FCM.CredentialsPath)
		if err != nil {
			log.Printf("Warning: Failed to initialize FCM client: %v", err)
		} else {
			log.Println("✅ FCM client initialized")
		}
	}

	if cfg.WebPush.VAPIDPublicKey != "" && cfg.WebPush.VAPIDPrivateKey != "" {
		webPushClient = webpush.NewWebPushClient(
			cfg.WebPush.VAPIDPublicKey,
			cfg.WebPush.VAPIDPrivateKey,
			cfg.WebPush.Subscriber,
		)
		log.Println("✅ Web Push client initialized")
	}

	// 4. Initialize Services
	notificationService := service.NewNotificationService(notificationRepo, fcmClient, webPushClient)
	auditService := service.NewAuditService(auditRepo)
	classService := service.NewClassService(classRepo, userRepo, auditService, notificationService, whiteboardRepo)

	// 4.1. Initialize Workers
	// Generic AI Worker
	aiWorker := worker.NewAIWorker(aiRepo, textOpenAI, cfg)
	aiWorker.Start()

	// KMJ Worker
	chKMJ, err := rabbitmq.NewChannel()
	if err != nil {
		log.Fatalf("Failed to open channel for KMJ: %v", err)
	}
	kmjWorker := worker.NewKMJWorker(kmjRepo, notificationService, textOpenAI, chKMJ)

	// Whiteboard Worker
	chWhiteboard, err := rabbitmq.NewChannel()
	if err != nil {
		log.Fatalf("Failed to open channel for Whiteboard: %v", err)
	}
	whiteboardWorker := worker.NewWhiteboardWorker(whiteboardRepo, notificationService, classService, textKimi, chWhiteboard)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("🚀 AI Workers started. Press Ctrl+C to stop.")

	// Presentation Worker
	chPresentation, err := rabbitmq.NewChannel()
	if err != nil {
		log.Printf("Failed to open channel for Presentation: %v", err)
	} else {
		presentationWorker := worker.NewPresentationWorker(presentationRepo, notificationService, textKimi, aiGen, chPresentation)
		go func() {
			// A consumer that stops (closed channel) must not fail silently
			if err := presentationWorker.Start(ctx); ctx.Err() == nil {
				log.Fatalf("Presentation Worker stopped: %v", err)
			}
		}()
	}

	// Article Worker
	chArticle, err := rabbitmq.NewChannel()
	if err != nil {
		log.Printf("Failed to open channel for Article: %v", err)
	} else {
		articleWorker := worker.NewArticleWorker(articleRepo, notificationService, textKimi, chArticle)
		go func() {
			// A consumer that stops (closed channel) must not fail silently
			if err := articleWorker.Start(ctx); ctx.Err() == nil {
				log.Fatalf("Article Worker stopped: %v", err)
			}
		}()
	}

	// Whiteboard Worker
	go func() {
		// A consumer that stops (closed channel) must not fail silently
		if err := whiteboardWorker.Start(ctx); ctx.Err() == nil {
			log.Fatalf("Whiteboard Worker stopped: %v", err)
		}
	}()

	// KMJ Worker
	go func() {
		// A consumer that stops (closed channel) must not fail silently
		if err := kmjWorker.Start(ctx); ctx.Err() == nil {
			log.Fatalf("KMJ Worker stopped: %v", err)
		}
	}()

	// Group Task Worker
	chGroupTask, err := rabbitmq.NewChannel()
	if err != nil {
		log.Printf("Failed to open channel for Group Task: %v", err)
	} else {
		groupTaskWorker := worker.NewGroupTaskWorker(groupTaskRepo, notificationService, textKimi, chGroupTask)
		go func() {
			// A consumer that stops (closed channel) must not fail silently
			if err := groupTaskWorker.Start(ctx); ctx.Err() == nil {
				log.Fatalf("Group Task Worker stopped: %v", err)
			}
		}()
	}

	// Tug of War Worker
	chTugOfWar, err := rabbitmq.NewChannel()
	if err != nil {
		log.Printf("Failed to open channel for Tug of War: %v", err)
	} else {
		towWorker := worker.NewTugOfWarWorker(tugOfWarRepo, notificationService, textKimi, chTugOfWar)
		go func() {
			// A consumer that stops (closed channel) must not fail silently
			if err := towWorker.Start(ctx); ctx.Err() == nil {
				log.Fatalf("Tug of War Worker stopped: %v", err)
			}
		}()
	}

	// AI Studio Worker (images, videos, feedback templates, lab work, news, certificates)
	chStudio, err := rabbitmq.NewChannel()
	if err != nil {
		log.Printf("Failed to open channel for AI Studio: %v", err)
	} else {
		studioWorker := worker.NewStudioWorker(studioRepo, notificationService, textKimi, ai.NewMediaClient(cfg.OpenAI.APIKey), chStudio)
		go func() {
			// A consumer that stops (closed channel) must not fail silently
			if err := studioWorker.Start(ctx); ctx.Err() == nil {
				log.Fatalf("AI Studio Worker stopped: %v", err)
			}
		}()
	}

	// Test Worker
	chTest, err := rabbitmq.NewChannel()
	if err != nil {
		log.Printf("Failed to open channel for Test: %v", err)
	} else {
		testWorker := worker.NewTestWorker(testRepo, notificationService, textKimi, chTest)
		go func() {
			// A consumer that stops (closed channel) must not fail silently
			if err := testWorker.Start(ctx); ctx.Err() == nil {
				log.Fatalf("Test Worker stopped: %v", err)
			}
		}()
	}

	// If RabbitMQ drops the connection, every consumer stops. Exit so that
	// Docker (restart: unless-stopped) starts a fresh worker instead of a silent one.
	connClosed := rabbitmq.NotifyConnectionClose()
	select {
	case <-ctx.Done():
	case err := <-connClosed:
		log.Fatalf("RabbitMQ connection lost (%v), restarting worker", err)
	}
	log.Println("👋 AI Workers stopping...")
	// Small delay to allow goroutines to finish
	time.Sleep(2 * time.Second)
	log.Println("👋 AI Workers stopped gracefully.")
}
