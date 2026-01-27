package handlers

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/config"
	"hysteria2_microservices/agent-service/internal/services"
	pb "hysteria2_microservices/proto"
)

// Agent handles the main agent logic
type Agent struct {
	localServices *services.LocalServices
	masterClient  pb.MasterServiceClient
	config        *config.Config
	logger        *logrus.Logger
}

// NewAgent creates a new Agent
func NewAgent(localServices *services.LocalServices, masterClient pb.MasterServiceClient, cfg *config.Config, logger *logrus.Logger) *Agent {
	return &Agent{
		localServices: localServices,
		masterClient:  masterClient,
		config:        cfg,
		logger:        logger,
	}
}

// Start starts the agent
func (a *Agent) Start(ctx context.Context) error {
	a.logger.Info("Starting agent...")

	// Register with master if client available
	if a.masterClient != nil {
		if err := a.registerWithMaster(ctx); err != nil {
			a.logger.Errorf("Failed to register with master: %v", err)
		}
	}

	// Start heartbeat if master client available
	if a.masterClient != nil {
		go a.heartbeatLoop(ctx)
	}

	// Enable masquerading if configured
	if a.config.Network.EnableMasquerading {
		if err := a.localServices.NetworkManager.EnableMasquerading(a.config.Network.DefaultInterface); err != nil {
			a.logger.Errorf("Failed to enable masquerading on startup: %v", err)
		} else {
			a.logger.Infof("Masquerading enabled on interface %s", a.config.Network.DefaultInterface)
		}
	}

	// Check and enable BBR if configured
	if a.config.Hysteria2.EnableBBR {
		if err := a.localServices.SystemManager.CheckAndEnableBBR(); err != nil {
			a.logger.Errorf("Failed to enable BBR: %v", err)
		} else {
			a.logger.Info("BBR checked/enabled successfully")
		}
	}

	// Install Hysteria2 if not installed
	if !a.localServices.HysteriaManager.IsHysteria2Installed() {
		a.logger.Info("Hysteria2 not installed, installing...")
		if err := a.localServices.HysteriaManager.InstallHysteria2(); err != nil {
			a.logger.Errorf("Failed to install Hysteria2: %v", err)
		} else {
			a.logger.Info("Hysteria2 installed successfully")
		}
	}

	a.logger.Info("Agent started")
	return nil
}

func (a *Agent) registerWithMaster(ctx context.Context) error {
	req := &pb.RegisterNodeRequest{
		Name:      a.config.Node.Name,
		Hostname:  a.config.Node.Hostname,
		IpAddress: a.config.Node.IPAddress,
		Location:  a.config.Node.Location,
		Country:   a.config.Node.Country,
		GrpcPort:  int32(a.config.Node.GRPCPort),
		Version:   "1.0.0",
		Capabilities: map[string]string{
			"masquerading":  "true",
			"network":       "true",
			"hysteria2":     "true",
			"xray":          "true",
			"protocols":     "hysteria2,vless,vless-reality",
			"vless":         "true",
			"vless-reality": "true",
		},
		AuthToken: "dummy-token", // TODO: implement proper auth
		Metadata:  a.config.Node.Metadata,
	}

	resp, err := a.masterClient.RegisterNode(ctx, req)
	if err != nil {
		return err
	}

	if resp.Success {
		a.logger.Infof("Registered with master server, node ID: %s", resp.NodeId)
		// Store node ID if needed
	} else {
		a.logger.Errorf("Registration failed: %s", resp.Message)
	}

	return nil
}

func (a *Agent) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // TODO: configurable
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.sendHeartbeat(ctx); err != nil {
				a.logger.Errorf("Failed to send heartbeat: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) sendHeartbeat(ctx context.Context) error {
	// Collect metrics
	metrics, err := a.localServices.MetricsCollector.Collect()
	if err != nil {
		a.logger.Errorf("Failed to collect metrics: %v", err)
	}

	// Convert metrics to map[string]float64
	metricValues := make(map[string]float64)
	for k, v := range metrics {
		if f, ok := v.(float64); ok {
			metricValues[k] = f
		}
	}

	req := &pb.HeartbeatRequest{
		NodeId:    a.config.Node.ID,
		Status:    "online",
		Metrics:   metricValues,
		Timestamp: nil, // Will be set by protobuf
	}

	resp, err := a.masterClient.Heartbeat(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success {
		a.logger.Warnf("Heartbeat failed: %s", resp.Message)
	}

	return nil
}
