package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"hysteria2_microservices/agent-service/internal/config"
	"hysteria2_microservices/agent-service/internal/handlers"
	"hysteria2_microservices/agent-service/internal/services"
	pb "hysteria2_microservices/proto"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)

	// Initialize services
	localServices := setupLocalServices(cfg, logger)

	// Setup gRPC client to master server
	masterConn, err := setupMasterClient(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to master server: %v", err)
	}
	var masterClient pb.MasterServiceClient
	if masterConn != nil {
		masterClient = pb.NewMasterServiceClient(masterConn)
	}
	defer func() {
		if masterConn != nil {
			if err := masterConn.Close(); err != nil {
				logger.Errorf("Failed to close master connection: %v", err)
			} else {
				logger.Info("Master connection closed")
			}
		}
	}()

	// Setup gRPC server for master commands
	grpcServer := setupGRPCServer(localServices, masterClient, logger)

	// Start agent
	agent := handlers.NewAgent(localServices, masterClient, cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)

	// Start agent
	g.Go(func() error {
		return agent.Start(gctx)
	})

	// Start gRPC server
	g.Go(func() error {
		return startGRPCServer(grpcServer, cfg, logger)
	})

	// Wait for interrupt signal or error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Infof("Received shutdown signal: %s", sig.String())
	case <-gctx.Done():
		logger.Info("Context cancelled")
	}

	logger.Info("Shutting down agent...")
	cancel()
	grpcServer.GracefulStop()
	logger.Info("Agent stopped gracefully")
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

func setupLocalServices(cfg *config.Config, logger *logrus.Logger) *services.LocalServices {
	return &services.LocalServices{
		ConfigManager:    services.NewConfigManager(logger),
		MetricsCollector: services.NewMetricsCollector(cfg, logger),
		SystemManager:    services.NewSystemManager(logger),
		NetworkManager:   services.NewNetworkManager(logger, cfg),
		HysteriaManager:  services.NewHysteriaManager(logger, cfg),
		WARPManager:      services.NewWARPManager(logger, cfg),
	}
}

func setupMasterClient(cfg *config.Config, logger *logrus.Logger) (*grpc.ClientConn, error) {
	if cfg.MasterServer == "" {
		logger.Warn("No master server configured, running in standalone mode")
		return nil, nil
	}

	conn, err := grpc.Dial(cfg.MasterServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master server: %w", err)
	}

	logger.Infof("Connected to master server: %s", cfg.MasterServer)

	return conn, nil
}

func setupGRPCServer(localServices *services.LocalServices, masterClient pb.MasterServiceClient, logger *logrus.Logger) *grpc.Server {
	s := grpc.NewServer()

	// Register node manager service
	pb.RegisterNodeManagerServer(s, handlers.NewNodeManagerHandler(localServices, logger))

	return s
}

func startGRPCServer(s *grpc.Server, cfg *config.Config, logger *logrus.Logger) error {
	addr := fmt.Sprintf(":%d", cfg.Node.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	logger.Infof("Starting gRPC server on %s", addr)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	return nil
}
