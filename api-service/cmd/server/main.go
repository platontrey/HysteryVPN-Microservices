package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"golang.org/x/sync/errgroup"

	"hysteria2_microservices/api-service/internal/config"
	"hysteria2_microservices/api-service/internal/database"
	"hysteria2_microservices/api-service/internal/handlers"
	"hysteria2_microservices/api-service/internal/middleware"
	"hysteria2_microservices/api-service/internal/repositories"
	"hysteria2_microservices/api-service/internal/services"
	"hysteria2_microservices/api-service/pkg/cache"
	"hysteria2_microservices/api-service/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger := logger.NewLogger(cfg.LogLevel)

	// Initialize database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			appLogger.Error("Failed to get database instance", "error", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("Failed to close database connection", "error", err)
		} else {
			appLogger.Info("Database connection closed")
		}
	}()

	// Initialize Redis cache
	redisClient := cache.NewRedisClient(cfg.RedisURL)
	defer func() {
		if err := redisClient.Close(); err != nil {
			appLogger.Error("Failed to close Redis connection", "error", err)
		} else {
			appLogger.Info("Redis connection closed")
		}
	}()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	deviceRepo := repositories.NewDeviceRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	trafficRepo := repositories.NewTrafficRepository(db)
	nodeRepo := repositories.NewNodeRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, redisClient, cfg.JWTSecret, time.Hour*time.Duration(cfg.JWTExpiryHour))
	userService := services.NewUserService(userRepo, deviceRepo, redisClient)
	nodeService := services.NewNodeService(nodeRepo, appLogger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, appLogger)
	userHandler := handlers.NewUserHandler(userService, appLogger)
	nodeHandler := handlers.NewNodeHandler(nodeService, appLogger)

	// Initialize WebSocket handler first (no dependency on trafficService yet)
	wsHandler := handlers.NewWebSocketHandler(nil, appLogger) // Will set trafficService later

	// Initialize traffic service with wsHandler
	trafficService := services.NewTrafficService(trafficRepo, redisClient, wsHandler)

	// Set trafficService in wsHandler
	wsHandler.SetTrafficService(trafficService)

	// Initialize remaining handlers
	trafficHandler := handlers.NewTrafficHandler(trafficService, appLogger)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(middleware.Logging(appLogger))
	app.Use(middleware.Metrics())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Metrics endpoint - TODO: Implement proper Fiber-compatible prometheus handler
	// app.Get("/metrics", func(c *fiber.Ctx) error {
	// 	promhttp.Handler().ServeHTTP(c.Response(), c.Request())
	// 	return nil
	// })

	// API routes
	api := app.Group("/api/v1")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	// Protected routes
	protected := api.Group("", middleware.JWTAuth(authService))

	// User routes
	users := protected.Group("/users")
	users.Get("", userHandler.GetUsers)
	users.Post("", userHandler.CreateUser)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	// Device routes
	users.Group("/:userId/devices").Get("", userHandler.GetUserDevices)

	// Node routes
	nodes := protected.Group("/nodes")
	nodes.Get("", nodeHandler.GetNodes)
	nodes.Post("", nodeHandler.CreateNode)
	nodes.Get("/:id", nodeHandler.GetNode)
	nodes.Put("/:id", nodeHandler.UpdateNode)
	nodes.Delete("/:id", nodeHandler.DeleteNode)
	nodes.Get("/:id/metrics", nodeHandler.GetNodeMetrics)
	nodes.Post("/:id/restart", nodeHandler.RestartNode)
	nodes.Get("/:id/logs", nodeHandler.GetNodeLogs)

	// Traffic routes
	traffic := protected.Group("/traffic")
	traffic.Get("/users/:userId", trafficHandler.GetUserTraffic)
	traffic.Get("/summary", trafficHandler.GetTrafficSummary)

	// WebSocket routes
	app.Get("/ws", middleware.JWTAuth(authService), wsHandler.WebSocketUpgrade())

	appLogger.Info("Server started", "port", cfg.Port)

	g, gctx := errgroup.WithContext(context.Background())

	// Start server
	g.Go(func() error {
		return app.Listen(":" + cfg.Port)
	})

	// Wait for interrupt signal or error
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-c:
		appLogger.Info("Received shutdown signal", "signal", sig.String())
	case <-gctx.Done():
		appLogger.Info("Context cancelled")
	}

	appLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
	} else {
		appLogger.Info("Server shutdown gracefully")
	}

	appLogger.Info("Server exited")
}
