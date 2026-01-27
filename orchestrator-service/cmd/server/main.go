package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hysteria2_microservices/orchestrator-service/internal/config"
	"hysteria2_microservices/orchestrator-service/internal/database"
	"hysteria2_microservices/orchestrator-service/internal/models"
	"hysteria2_microservices/orchestrator-service/internal/repositories"
	"hysteria2_microservices/orchestrator-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"hysteria2_microservices/orchestrator-service/pkg/proto"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			logger.Errorf("Failed to close database connection: %v", err)
		} else {
			logger.Info("Database connection closed")
		}
	}()

	// Run migrations
	if err := database.AutoMigrate(db, &models.VPSNode{}, &models.NodeAssignment{}, &models.NodeMetric{}, &models.Deployment{}, &models.User{}); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	repos := setupRepositories(db)

	// Initialize services
	services := setupServices(repos, logger)

	// Setup GRPC server
	grpcServer := setupGRPCServer(services, cfg, logger)
	go startGRPCServer(grpcServer, cfg, logger)

	// Setup REST server
	restServer := setupRESTServer(services, cfg, logger)
	go startRESTServer(restServer, cfg, logger)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Infof("Received shutdown signal: %s", sig.String())

	logger.Info("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown REST server
	if err := restServer.Shutdown(ctx); err != nil {
		logger.Errorf("REST server shutdown error: %v", err)
	} else {
		logger.Info("REST server shutdown gracefully")
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()
	logger.Info("gRPC server shutdown gracefully")
	logger.Info("Servers stopped")
}

func setupLogger(cfg config.LoggingConfig) *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return logger
}

func setupRepositories(db *database.Database) *repositories.Repositories {
	return &repositories.Repositories{
		NodeRepo:       repositories.NewNodeRepository(db.DB),
		AssignmentRepo: repositories.NewNodeAssignmentRepository(db.DB),
		MetricRepo:     repositories.NewNodeMetricRepository(db.DB),
		DeploymentRepo: repositories.NewDeploymentRepository(db.DB),
		UserRepo:       repositories.NewUserRepository(db.DB),
	}
}

func setupServices(repos *repositories.Repositories, logger *logrus.Logger) *services.Services {
	return &services.Services{
		NodeService:       services.NewNodeService(repos.NodeRepo, logger),
		AssignmentService: services.NewAssignmentService(repos.AssignmentRepo, logger),
		MetricsService:    services.NewMetricsService(repos.MetricRepo, logger),
		DeploymentService: services.NewDeploymentService(repos.DeploymentRepo, logger),
		UserService:       services.NewUserService(repos.UserRepo, logger),
	}
}

func setupGRPCServer(services *services.Services, cfg *config.Config, logger *logrus.Logger) *grpc.Server {
	// Setup TLS if configured
	var opts []grpc.ServerOption

	if cfg.GRPC.HostKey != "" && cfg.GRPC.CertKey != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.GRPC.CertKey, cfg.GRPC.HostKey)
		if err != nil {
			logger.Fatalf("Failed to create TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)

	// Register services
	node_management.RegisterMasterServiceServer(s, handlers.NewMasterServiceHandler(services.NodeService, logger))
	node_management.RegisterAdminServiceServer(s, handlers.NewAdminServiceHandler(services.NodeService, services.MetricsService, logger))

	// Enable reflection for development
	reflection.Register(s)

	return s
}

func startGRPCServer(s *grpc.Server, cfg *config.Config, logger *logrus.Logger) {
	addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	logger.Infof("Starting gRPC server on %s", addr)
	if err := s.Serve(lis); err != nil {
		logger.Fatalf("Failed to start gRPC server: %v", err)
	}
}

func setupRESTServer(services *services.Services, cfg *config.Config, logger *logrus.Logger) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())

	// Setup routes
	handlers.SetupRoutes(r, services, logger)

	return r
}

func startRESTServer(r *gin.Engine, cfg *config.Config, logger *logrus.Logger) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Starting REST server on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Failed to start REST server: %v", err)
	}
}
