package main

import (
	"context"
	"log"
	"os"

	_ "zeinai-backend/docs" // Import generated docs
	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/handler"
	"zeinai-backend/internal/middleware"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/pkg/fcm"
	"zeinai-backend/internal/pkg/i18n"
	"zeinai-backend/internal/pkg/mailer"
	"zeinai-backend/internal/pkg/mobizon"
	"zeinai-backend/internal/pkg/webpush"
	"zeinai-backend/internal/repository"
	"zeinai-backend/internal/service"
	"zeinai-backend/internal/worker"
	"zeinai-backend/platform/database"
	"zeinai-backend/platform/rabbitmq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title Zeinai EdTech API
// @version 1.0
// @description Backend API for Zeinai EdTech Platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@zeinai.kz

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Initialize Infrastructure
	// I18n
	if err := i18n.Init("locales", "en"); err != nil {
		log.Printf("Warning: failed to load locales: %v", err)
	}

	// Database
	dbPool := database.ConnectDB(*cfg)
	defer database.CloseDB()

	// RabbitMQ
	if err := rabbitmq.InitRabbitMQ(cfg); err != nil {
		log.Fatalf("Failed to init RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	// AI Generator - Using OpenAI as primary, Kimi K2 for presentation generation
	aiGen := ai.NewOpenAIGenerator(cfg)
	kimiGen := ai.NewKimiGenerator(cfg)
	textKimi := ai.KimiFirst(cfg, kimiGen, aiGen)
	textOpenAI := ai.OpenAIFirst(cfg, aiGen, kimiGen)

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(dbPool)
	aiRepo := repository.NewAIRepository(dbPool)
	schoolRepo := repository.NewSchoolRepository(dbPool)
	subscriptionRepo := repository.NewSubscriptionRepository(dbPool)
	notificationRepo := repository.NewNotificationRepository(dbPool)
	verificationRepo := repository.NewVerificationRepository(dbPool)
	classRepo := repository.NewClassRepository(dbPool)
	auditRepo := repository.NewAuditRepository(dbPool)
	kmjRepo := repository.NewKMJRepository(dbPool)
	nationalValueRepo := repository.NewNationalValueRepository(dbPool)
	methodRepo := repository.NewTeachingMethodRepository(dbPool)
	presentationRepo := repository.NewPresentationRepository(dbPool)
	articleRepo := repository.NewArticleRepository(dbPool)
	whiteboardRepo := repository.NewWhiteboardRepository(dbPool)
	groupTaskRepo := repository.NewGroupTaskRepository(dbPool)
	tugOfWarRepo := repository.NewTugOfWarRepository(dbPool)
	testRepo := repository.NewTestRepository(dbPool)
	studioRepo := repository.NewStudioRepository(dbPool)

	// 3.5. Initialize Push Notification Clients
	// ... (no changes here usually)
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

	// 3.6. Initialize SMS and Mailer Clients
	var mobizonClient *mobizon.MobizonClient
	if cfg.Mobizon.APIKey != "" {
		mobizonClient = mobizon.NewMobizonClient(cfg.Mobizon.APIKey)
		log.Println("✅ Mobizon SMS client initialized")
	}

	mailClient := mailer.NewMailer(cfg.SMTP)
	log.Println("✅ Mailer client initialized")

	// 4. Initialize Services (Audit first)
	auditService := service.NewAuditService(auditRepo)
	notificationService := service.NewNotificationService(notificationRepo, fcmClient, webPushClient)
	classService := service.NewClassService(classRepo, userRepo, auditService, notificationService, whiteboardRepo)
	authService := service.NewAuthService(userRepo, cfg, auditService, classService)
	aiService := service.NewAIService(aiRepo)
	userService := service.NewUserService(userRepo, auditService)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	verificationService := service.NewVerificationService(verificationRepo, userRepo, mobizonClient, mailClient, auditService)
	adminService := service.NewAdminService(userRepo, auditService, notificationService)
	kmjService := service.NewKMJService(kmjRepo, rabbitmq.GetChannel(), subscriptionRepo, nationalValueRepo, methodRepo)
	presentationService := service.NewPresentationService(presentationRepo, subscriptionRepo, rabbitmq.GetChannel())
	articleService := service.NewArticleService(articleRepo, subscriptionRepo, rabbitmq.GetChannel())
	whiteboardService := service.NewWhiteboardService(whiteboardRepo, userRepo, notificationService, classService, rabbitmq.GetChannel())
	groupTaskService := service.NewGroupTaskService(groupTaskRepo, classRepo, subscriptionRepo, rabbitmq.GetChannel())
	tugOfWarService := service.NewTugOfWarService(tugOfWarRepo, subscriptionRepo, rabbitmq.GetChannel())
	testService := service.NewTestService(testRepo, subscriptionRepo, classRepo, userRepo, notificationService, rabbitmq.GetChannel())
	studioService := service.NewStudioService(studioRepo, subscriptionRepo, rabbitmq.GetChannel(), cfg)

	// 5. Initialize Workers (Optional in API container)
	if aiGen != nil && os.Getenv("START_WORKERS") == "true" {
		aiWorker := worker.NewAIWorker(aiRepo, textOpenAI, cfg)
		aiWorker.Start()

		// KMJ Worker
		kmjWorker := worker.NewKMJWorker(kmjRepo, notificationService, textOpenAI, rabbitmq.GetChannel())
		go func() {
			if err := kmjWorker.Start(context.Background()); err != nil {
				log.Printf("KMJ Worker error: %v", err)
			}
		}()

		// Presentation Worker (Kimi K2 for text, OpenAI for images)
		presentationWorker := worker.NewPresentationWorker(presentationRepo, notificationService, textKimi, aiGen, rabbitmq.GetChannel())
		go func() {
			if err := presentationWorker.Start(context.Background()); err != nil {
				log.Printf("Presentation Worker error: %v", err)
			}
		}()

		// Article Worker (Kimi K2)
		articleWorker := worker.NewArticleWorker(articleRepo, notificationService, textKimi, rabbitmq.GetChannel())
		go func() {
			if err := articleWorker.Start(context.Background()); err != nil {
				log.Printf("Article Worker error: %v", err)
			}
		}()

		// Whiteboard Worker (Kimi K2)
		whiteboardWorker := worker.NewWhiteboardWorker(whiteboardRepo, notificationService, classService, textKimi, rabbitmq.GetChannel())
		go func() {
			if err := whiteboardWorker.Start(context.Background()); err != nil {
				log.Printf("Whiteboard Worker error: %v", err)
			}
		}()

		// Group Task Worker
		groupTaskWorker := worker.NewGroupTaskWorker(groupTaskRepo, notificationService, textKimi, rabbitmq.GetChannel())
		go func() {
			if err := groupTaskWorker.Start(context.Background()); err != nil {
				log.Printf("GroupTask Worker error: %v", err)
			}
		}()

		// Tug of War Worker
		tugOfWarWorker := worker.NewTugOfWarWorker(tugOfWarRepo, notificationService, textKimi, rabbitmq.GetChannel())
		go func() {
			if err := tugOfWarWorker.Start(context.Background()); err != nil {
				log.Printf("TugOfWar Worker error: %v", err)
			}
		}()

		// AI Studio Worker (images, videos, feedback templates, lab work, news, certificates)
		if studioCh, err := rabbitmq.NewChannel(); err != nil {
			log.Printf("Failed to open channel for AI Studio: %v", err)
		} else {
			studioWorker := worker.NewStudioWorker(studioRepo, notificationService, textKimi, ai.NewMediaClient(cfg.OpenAI.APIKey), studioCh)
			go func() {
				if err := studioWorker.Start(context.Background()); err != nil {
					log.Printf("AI Studio Worker error: %v", err)
				}
			}()
		}

		// Test Worker
		testWorker := worker.NewTestWorker(testRepo, notificationService, textKimi, rabbitmq.GetChannel())
		go func() {
			if err := testWorker.Start(context.Background()); err != nil {
				log.Printf("Test Worker error: %v", err)
			}
		}()
	}

	// 6. Initialize Handlers
	authHandler := handler.NewAuthHandler(authService)
	aiHandler := handler.NewAIHandler(aiService)
	schoolHandler := handler.NewSchoolHandler(schoolRepo)
	userHandler := handler.NewUserHandler(userService, verificationService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, userService)
	notificationHandler := handler.NewNotificationHandler(notificationService, notificationRepo)
	verificationHandler := handler.NewVerificationHandler(verificationService)
	classHandler := handler.NewClassHandler(classService)
	auditHandler := handler.NewAuditHandler(auditService)
	adminHandler := handler.NewAdminHandler(adminService)
	educationHandler := handler.NewEducationHandler()
	kmjHandler := handler.NewKMJHandler(kmjService)
	presentationHandler := handler.NewPresentationHandler(presentationService)
	articleHandler := handler.NewArticleHandler(articleService)
	whiteboardHandler := handler.NewWhiteboardHandler(whiteboardService)
	groupTaskHandler := handler.NewGroupTaskHandler(groupTaskService)
	tugOfWarHandler := handler.NewTugOfWarHandler(tugOfWarService)
	testHandler := handler.NewTestHandler(testService)
	studioHandler := handler.NewStudioHandler(studioService)

	// 7. Initialize Fiber
	app := fiber.New(fiber.Config{
		AppName: "Zeinai EdTech API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": err.Error(),
				},
			})
		},
	})

	// 8. Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(func(c *fiber.Ctx) error {
		lang := c.Get("Accept-Language")
		if lang == "" {
			lang = "en"
		}
		if len(lang) > 2 {
			lang = lang[:2]
		}
		c.Locals("lang", lang)
		return c.Next()
	})

	// 9. Routes
	api := app.Group("/api/v1")
	app.Static("/api/v1/uploads", "./uploads")

	// Swagger
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Auth
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/login/student", authHandler.LoginStudent)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Get("/check-iin", authHandler.CheckIIN) // Public
	auth.Post("/logout", middleware.Protected(cfg), authHandler.Logout)
	// Student reset password flow
	auth.Post("/reset-password/student/init", authHandler.InitStudentReset)
	auth.Post("/reset-password/student/confirm", authHandler.ConfirmStudentReset)
	// Google OAuth
	auth.Get("/google/login", authHandler.GoogleLogin)
	auth.Get("/google/callback", authHandler.GoogleCallback)
	auth.Post("/google/register/complete", authHandler.GoogleRegisterComplete)

	// Schools (Public for registration)
	api.Get("/schools", schoolHandler.List)

	// Users (Protected)
	users := api.Group("/users", middleware.Protected(cfg))
	users.Get("/me", userHandler.GetProfile)
	users.Put("/me", userHandler.UpdateProfile)
	users.Put("/me/password", userHandler.ChangePassword)
	users.Post("/me/email/request", userHandler.RequestEmailChange)
	users.Post("/me/email/verify", userHandler.VerifyEmailChange)
	users.Post("/me/phone/request", userHandler.RequestPhoneChange)
	users.Post("/me/phone/verify", userHandler.VerifyPhoneChange)
	users.Post("/me/avatar", userHandler.UploadAvatar)
	users.Delete("/me/avatar", userHandler.DeleteAvatar)

	// Subscriptions (Protected)
	subscriptions := api.Group("/subscriptions", middleware.Protected(cfg))
	subscriptions.Get("/current", subscriptionHandler.GetSubscription)
	subscriptions.Post("/upgrade", subscriptionHandler.UpgradeSubscription)
	subscriptions.Post("/renew", subscriptionHandler.RenewSubscription)

	// Notifications (Protected)
	notifications := api.Group("/notifications", middleware.Protected(cfg))
	notifications.Get("", notificationHandler.GetNotifications)
	notifications.Put("/:id/read", notificationHandler.MarkAsRead)
	notifications.Put("/read-all", notificationHandler.MarkAllAsRead)
	notifications.Post("/register-token", notificationHandler.RegisterPushToken)
	notifications.Get("/tokens", notificationHandler.GetPushTokens)
	notifications.Delete("/tokens/:id", notificationHandler.DeactivatePushToken)
	notifications.Delete("/:id", notificationHandler.DeleteNotification)

	// Verification (Public - for registration and password reset)
	verification := api.Group("/verification")
	verification.Post("/send-code", verificationHandler.SendCode)
	verification.Post("/verify-code", verificationHandler.VerifyCode)
	verification.Post("/reset-password", verificationHandler.ResetPassword)
	verification.Get("/status", verificationHandler.GetStatus)

	// AI (Protected + Token Limit Check)
	ai := api.Group("/ai", middleware.Protected(cfg))
	ai.Post("/generate", middleware.CheckTokenLimit(userRepo), aiHandler.GenerateText)
	ai.Post("/suggest-article-instructions", aiHandler.SuggestArticleInstructions)
	ai.Post("/suggest-presentation-instructions", aiHandler.SuggestPresentationInstructions)
	ai.Get("/status/:id", aiHandler.GetStatus)

	// Classes (Teacher)
	classes := api.Group("/classes", middleware.Protected(cfg))
	classes.Post("", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.CreateClass)
	classes.Get("", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.ListClasses)
	classes.Get("/:id", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.GetClass)
	classes.Put("/:id", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.UpdateClass)
	classes.Delete("/:id", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.DeleteClass)
	classes.Post("/:id/students", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.AddStudent)
	classes.Get("/:id/students", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.ListStudents)
	classes.Delete("/:id/students/:student_id", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.RemoveStudent)
	classes.Post("/:id/invites", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.CreateInvite)
	classes.Get("/:id/invites/active", middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin), classHandler.GetActiveInvite)

	// Teacher specific (not directly class-related)
	teacher := api.Group("/teacher", middleware.Protected(cfg), middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin))
	teacher.Put("/students/:id/reset-password", userHandler.ResetStudentPassword)
	teacher.Put("/students/:id", userHandler.UpdateStudent)
	teacher.Post("/notify-student", classHandler.NotifyStudent)
	teacher.Post("/notify-class", classHandler.NotifyClass)
	teacher.Post("/notify-all-students", classHandler.NotifyAllStudents)

	invites := api.Group("/invites")
	invites.Get("/:code", classHandler.GetInviteInfo) // Public
	invites.Post("/:code/accept", middleware.Protected(cfg), classHandler.AcceptInvite)

	admin := api.Group("/admin", middleware.Protected(cfg), middleware.RBAC(domain.RoleAdmin))
	admin.Get("/logs", auditHandler.GetLogs)
	// Verification Management
	admin.Get("/verification/blocks", verificationHandler.GetAllBlocks)
	admin.Delete("/verification/unblock", verificationHandler.Unblock)

	// Teacher Management
	admin.Post("/teachers", adminHandler.CreateTeacher)
	admin.Get("/teachers", adminHandler.ListTeachers)
	admin.Get("/teachers/:id", adminHandler.GetTeacher)
	admin.Put("/teachers/:id", adminHandler.UpdateTeacher)
	admin.Delete("/teachers/:id", adminHandler.DeleteTeacher)
	admin.Put("/teachers/:id/subscription", adminHandler.UpdateSubscription)
	admin.Post("/teachers/:id/set-password", adminHandler.SetTeacherPassword)
	admin.Post("/notify", adminHandler.NotifyTeacher)
	admin.Post("/notify-all", adminHandler.NotifyAllTeachers)

	// Student (Personal Cabinet)
	student := api.Group("/student", middleware.Protected(cfg), middleware.RBAC(domain.RoleStudent))
	student.Get("/dashboard", classHandler.GetStudentDashboard)
	testHandler.RegisterStudentRoutes(student)

	// KMJ (Lesson Plans) - Protected, Teacher only
	kmjHandler.RegisterRoutes(api, cfg)

	// Presentations (AI Slides via Kimi K2) - Protected, Teacher only
	presentationHandler.RegisterRoutes(api, cfg)

	// Articles (AI Articles via Kimi K2) - Protected, Teacher only
	articleHandler.RegisterRoutes(api, cfg)

	// Whiteboard (Real-time & AI)
	whiteboardHandler.RegisterRoutes(api, cfg)

	// Group Tasks (AI & Gamified)
	groupTaskHandler.RegisterRoutes(api, cfg)

	// Tug of War (AI & Gamified)
	tugOfWarHandler.RegisterRoutes(api, cfg)

	// Tests (AI Test Generation)
	testHandler.RegisterRoutes(api, cfg)
	studioHandler.RegisterRoutes(api, cfg)

	// Education Materials (Protected)
	education := api.Group("/education", middleware.Protected(cfg), middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin))
	education.Post("/short-term-plan", educationHandler.Placeholder)
	// /slides is now handled by /api/v1/presentations
	education.Post("/video", educationHandler.Placeholder)
	education.Post("/image", educationHandler.Placeholder)
	education.Post("/feedback", educationHandler.Placeholder)
	education.Post("/article", educationHandler.Placeholder)
	// /group-tasks is now handled via groupTaskHandler.RegisterRoutes
	// /tug-of-war is now handled via tugOfWarHandler.RegisterRoutes
	education.Post("/tests", educationHandler.Placeholder)
	education.Post("/pisa", educationHandler.Placeholder)
	education.Post("/unt", educationHandler.Placeholder)
	education.Post("/lab", educationHandler.Placeholder)
	education.Post("/olympiad", educationHandler.Placeholder)
	education.Post("/news", educationHandler.Placeholder)
	education.Post("/achievements", educationHandler.Placeholder)
	education.Post("/certificates", educationHandler.Placeholder)

	// 10. Start server
	log.Printf("Starting server on port %s in %s mode", cfg.App.Port, cfg.App.Env)
	log.Fatal(app.Listen(":" + cfg.App.Port))
}
