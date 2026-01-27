package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"hysteria2_microservices/agent-service/internal/config"
	"hysteria2_microservices/agent-service/internal/services"
	pb "hysteria2_microservices/proto"
)

// WARPProxyTester tests complete VPN -> WARP -> Internet functionality
type WARPProxyTester struct {
	logger  *logrus.Logger
	client  pb.NodeManagerClient
	conn    *grpc.Client
	warpMgr services.WARPManager
	monitor services.WARPMonitor
}

// NewWARPProxyTester creates a new tester
func NewWARPProxyTester(logger *logrus.Logger, agentAddr string) (*WARPProxyTester, error) {
	// Connect to agent
	conn, err := grpc.Dial(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}

	client := pb.NewNodeManagerClient(conn)

	// Create config and services
	cfg := &config.Config{
		Hysteria2: config.Hysteria2Config{
			WARPEnabled:      true,
			WARPProxyPort:    1080,
			WARPAutoConnect:  true,
			WARPNotifyOnFail: true,
			WARPClientType:   "local",
		},
	}

	warpMgr := services.NewWARPManager(logger, cfg)
	monitor := services.NewWARPMonitor(logger, cfg, warpMgr)

	return &WARPProxyTester{
		logger:  logger,
		client:  client,
		conn:    conn,
		warpMgr: warpMgr,
		monitor: monitor,
	}, nil
}

// RunComprehensiveTest runs full test suite
func (t *WARPProxyTester) RunComprehensiveTest(ctx context.Context) error {
	t.logger.Info("🚀 Starting comprehensive WARP proxy test suite...")

	// Test 1: Basic WARP installation and setup
	if err := t.testWARPInstallation(ctx); err != nil {
		return fmt.Errorf("WARP installation test failed: %w", err)
	}

	// Test 2: WARP connection
	if err := t.testWARPConnection(ctx); err != nil {
		return fmt.Errorf("WARP connection test failed: %w", err)
	}

	// Test 3: WARP proxy configuration
	if err := t.testWARPProxyConfiguration(ctx); err != nil {
		return fmt.Errorf("WARP proxy configuration test failed: %w", err)
	}

	// Test 4: Traffic routing setup
	if err := t.testTrafficRouting(ctx); err != nil {
		return fmt.Errorf("Traffic routing test failed: %w", err)
	}

	// Test 5: Hysteria2 integration
	if err := t.testHysteria2Integration(ctx); err != nil {
		return fmt.Errorf("Hysteria2 integration test failed: %w", err)
	}

	// Test 6: End-to-end connectivity
	if err := t.testEndToEndConnectivity(ctx); err != nil {
		return fmt.Errorf("End-to-end connectivity test failed: %w", err)
	}

	// Test 7: Monitoring and health checks
	if err := t.testMonitoringAndHealth(ctx); err != nil {
		return fmt.Errorf("Monitoring test failed: %w", err)
	}

	t.logger.Info("✅ All tests completed successfully!")
	return nil
}

// testWARPInstallation tests WARP client installation
func (t *WARPProxyTester) testWARPInstallation(ctx context.Context) error {
	t.logger.Info("📦 Testing WARP installation...")

	// Check if WARP is already installed
	if t.warpMgr.IsWARPInstalled() {
		t.logger.Info("✅ WARP client is already installed")
		return nil
	}

	// Install WARP
	resp, err := t.client.InstallWARPClient(ctx, &pb.InstallWARPClientRequest{
		NodeId: "test-node-1",
	})
	if err != nil {
		return fmt.Errorf("failed to install WARP: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("WARP installation failed: %s", resp.Message)
	}

	t.logger.Infof("✅ WARP installation successful: %s", resp.Message)
	return nil
}

// testWARPConnection tests WARP connection
func (t *WARPProxyTester) testWARPConnection(ctx context.Context) error {
	t.logger.Info("🔗 Testing WARP connection...")

	// Configure WARP
	configResp, err := t.client.ConfigureWARP(ctx, &pb.ConfigureWARPRequest{
		NodeId:       "test-node-1",
		Enabled:      true,
		ProxyPort:    1080,
		AutoConnect:  true,
		NotifyOnFail: true,
		ClientType:   "local",
		Mode:         "proxy",
	})
	if err != nil {
		return fmt.Errorf("failed to configure WARP: %w", err)
	}

	if !configResp.Success {
		return fmt.Errorf("WARP configuration failed: %s", configResp.Message)
	}

	// Connect to WARP
	connectResp, err := t.client.ConnectWARP(ctx, &pb.ConnectWARPRequest{
		NodeId: "test-node-1",
	})
	if err != nil {
		return fmt.Errorf("failed to connect WARP: %w", err)
	}

	if !connectResp.Success {
		return fmt.Errorf("WARP connection failed: %s", connectResp.Message)
	}

	// Wait for connection to establish
	time.Sleep(5 * time.Second)

	// Check status
	statusResp, err := t.client.GetWARPStatus(ctx, &pb.GetWARPStatusRequest{
		NodeId: "test-node-1",
	})
	if err != nil {
		return fmt.Errorf("failed to get WARP status: %w", err)
	}

	status := statusResp.Status
	if !status.Installed {
		return fmt.Errorf("WARP is not installed")
	}

	if !status.Connected {
		return fmt.Errorf("WARP is not connected")
	}

	t.logger.Infof("✅ WARP connection successful - IP: %s, Location: %s", status.IpAddress, status.Location)
	return nil
}

// testWARPProxyConfiguration tests WARP proxy configuration
func (t *WARPProxyTester) testWARPProxyConfiguration(ctx context.Context) error {
	t.logger.Info("⚙️ Testing WARP proxy configuration...")

	// Enable proxy mode
	proxyResp, err := t.client.EnableWARPProxy(ctx, &pb.EnableWARPProxyRequest{
		NodeId: "test-node-1",
		Port:   1080,
	})
	if err != nil {
		return fmt.Errorf("failed to enable WARP proxy: %w", err)
	}

	if !proxyResp.Success {
		return fmt.Errorf("WARP proxy enablement failed: %s", proxyResp.Message)
	}

	// Wait for proxy to start
	time.Sleep(3 * time.Second)

	t.logger.Info("✅ WARP proxy configuration successful")
	return nil
}

// testTrafficRouting tests traffic routing setup
func (t *WARPProxyTester) testTrafficRouting(ctx context.Context) error {
	t.logger.Info("🌐 Testing traffic routing setup...")

	// Enable traffic routing
	routingResp, err := t.client.EnableWARPTrafficRouting(ctx, &pb.EnableWARPTrafficRoutingRequest{
		NodeId:        "test-node-1",
		InterfaceName: "eth0", // Default interface
	})
	if err != nil {
		return fmt.Errorf("failed to enable WARP traffic routing: %w", err)
	}

	if !routingResp.Success {
		return fmt.Errorf("WARP traffic routing failed: %s", routingResp.Message)
	}

	t.logger.Info("✅ Traffic routing setup successful")
	return nil
}

// testHysteria2Integration tests Hysteria2 integration with WARP
func (t *WARPProxyTester) testHysteria2Integration(ctx context.Context) error {
	t.logger.Info("🔒 Testing Hysteria2 integration with WARP...")

	// Setup comprehensive WARP proxy endpoint
	setupResp, err := t.client.SetupWARPProxyEndpoint(ctx, &pb.SetupWARPProxyEndpointRequest{
		NodeId:              "test-node-1",
		WarpEnabled:         true,
		WarpProxyPort:       1080,
		AutoConnect:         true,
		NotifyOnFail:        true,
		ClientType:          "local",
		VpnInterface:        "eth0",
		VpnListenPort:       8080,
		VpnAuthPassword:     "test-password",
		BandwidthUp:         100,
		BandwidthDown:       100,
		ConfigureHysteria:   true,
		SetupTrafficRouting: true,
		EnableMasquerading:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to setup WARP proxy endpoint: %w", err)
	}

	if !setupResp.Success {
		return fmt.Errorf("WARP proxy endpoint setup failed: %s", setupResp.Message)
	}

	t.logger.Infof("✅ Hysteria2 integration successful - Config: %+v", setupResp.Config)
	return nil
}

// testEndToEndConnectivity tests complete end-to-end connectivity
func (t *WARPProxyTester) testEndToEndConnectivity(ctx context.Context) error {
	t.logger.Info("🔍 Testing end-to-end connectivity...")

	// Run connectivity test
	testResp, err := t.client.TestWARPProxyConnectivity(ctx, &pb.TestWARPProxyConnectivityRequest{
		NodeId: "test-node-1",
	})
	if err != nil {
		return fmt.Errorf("failed to run connectivity test: %w", err)
	}

	t.logger.Infof("🔍 Connectivity test results:")
	for test, passed := range testResp.Results {
		status := "✅"
		if !passed {
			status = "❌"
		}
		t.logger.Infof("  %s %s: %v", status, test, passed)
	}

	if !testResp.Success {
		return fmt.Errorf("End-to-end connectivity test failed: %s", testResp.Message)
	}

	t.logger.Info("✅ End-to-end connectivity test passed")
	return nil
}

// testMonitoringAndHealth tests monitoring and health checks
func (t *WARPProxyTester) testMonitoringAndHealth(ctx context.Context) error {
	t.logger.Info("📊 Testing monitoring and health checks...")

	// Start monitoring
	if err := t.monitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}

	// Wait for initial monitoring cycle
	time.Sleep(5 * time.Second)

	// Run health check
	healthResult := t.monitor.RunHealthCheck()

	t.logger.Infof("🏥 Health check results:")
	t.logger.Infof("  Overall Healthy: %v", healthResult.OverallHealthy)
	t.logger.Infof("  Score: %d/100", healthResult.Score)
	t.logger.Infof("  WARP Running: %v", healthResult.WARPRunning)
	t.logger.Infof("  WARP Connected: %v", healthResult.WARPConnected)
	t.logger.Infof("  Proxy Working: %v", healthResult.ProxyWorking)
	t.logger.Infof("  Internet Reachable: %v", healthResult.InternetReachable)
	t.logger.Infof("  DNS Working: %v", healthResult.DNSWorking)

	for checkName, result := range healthResult.Checks {
		status := "✅"
		if !result.Passed {
			status = "❌"
		}
		t.logger.Infof("  %s %s: %s", status, checkName, result.Message)
	}

	if !healthResult.OverallHealthy {
		return fmt.Errorf("Health check failed: %s", healthResult.Message)
	}

	// Get current status
	currentStatus := t.monitor.GetCurrentStatus()
	t.logger.Infof("📈 Current monitoring status:")
	t.logger.Infof("  WARP Connected: %v", currentStatus.WARPConnected)
	t.logger.Infof("  Health Score: %.1f/100", currentStatus.HealthScore)
	if len(currentStatus.HealthIssues) > 0 {
		t.logger.Infof("  Health Issues: %v", currentStatus.HealthIssues)
	}

	// Stop monitoring
	if err := t.monitor.Stop(); err != nil {
		t.logger.Warnf("Failed to stop monitoring: %v", err)
	}

	t.logger.Info("✅ Monitoring and health checks completed successfully")
	return nil
}

// Close cleans up resources
func (t *WARPProxyTester) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Get agent address from environment or use default
	agentAddr := os.Getenv("AGENT_ADDRESS")
	if agentAddr == "" {
		agentAddr = "localhost:50051"
	}

	// Create tester
	tester, err := NewWARPProxyTester(logger, agentAddr)
	if err != nil {
		log.Fatalf("Failed to create tester: %v", err)
	}
	defer tester.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run comprehensive test
	if err := tester.RunComprehensiveTest(ctx); err != nil {
		logger.Fatalf("Test suite failed: %v", err)
	}

	logger.Info("🎉 All WARP proxy tests completed successfully!")
	fmt.Println("\n" + "="*60)
	fmt.Println("🎯 WARP Proxy Implementation Summary")
	fmt.Println("=" * 60)
	fmt.Println("✅ WARP Client Integration")
	fmt.Println("✅ Proxy Configuration")
	fmt.Println("✅ Traffic Routing")
	fmt.Println("✅ Hysteria2 Integration")
	fmt.Println("✅ End-to-End Connectivity")
	fmt.Println("✅ Health Monitoring")
	fmt.Println("✅ Comprehensive Testing")
	fmt.Println("\n🚀 VPN -> WARP -> Internet proxy is ready!")
}
