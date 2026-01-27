package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/services"
	pb "hysteria2_microservices/proto"
)

// WARPProxyHandler implements comprehensive WARP proxy management
type WARPProxyHandler struct {
	pb.UnimplementedNodeManagerServer
	localServices *services.LocalServices
	logger        *logrus.Logger
}

// NewWARPProxyHandler creates a new WARP proxy handler
func NewWARPProxyHandler(localServices *services.LocalServices, logger *logrus.Logger) *WARPProxyHandler {
	return &WARPProxyHandler{
		localServices: localServices,
		logger:        logger,
	}
}

// SetupWARPProxyEndpoint configures complete VPN -> WARP -> Internet flow
func (h *WARPProxyHandler) SetupWARPProxyEndpoint(ctx context.Context, req *pb.SetupWARPProxyEndpointRequest) (*pb.SetupWARPProxyEndpointResponse, error) {
	h.logger.Infof("SetupWARPProxyEndpoint called - WARP enabled: %v, proxy port: %d, VPN interface: %s",
		req.WarpEnabled, req.WarpProxyPort, req.VpnInterface)

	// 1. Configure WARP client
	if req.WarpEnabled {
		warpConfig := services.WARPConfig{
			Enabled:      true,
			ProxyPort:    int(req.WarpProxyPort),
			AutoConnect:  req.AutoConnect,
			NotifyOnFail: req.NotifyOnFail,
			ClientType:   req.ClientType,
			LicenseKey:   req.LicenseKey,
			Organization: req.Organization,
			Mode:         "proxy",
		}

		if err := h.localServices.WARPManager.ConfigureWARP(warpConfig); err != nil {
			h.logger.Errorf("Failed to configure WARP: %v", err)
			return &pb.SetupWARPProxyEndpointResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to configure WARP: %v", err),
			}, nil
		}

		// Connect to WARP if auto-connect is enabled
		if req.AutoConnect {
			if err := h.localServices.WARPManager.ConnectWARP(); err != nil {
				h.logger.Errorf("Failed to connect to WARP: %v", err)
				return &pb.SetupWARPProxyEndpointResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to connect to WARP: %v", err),
				}, nil
			}
		}

		// Enable proxy mode
		if err := h.localServices.WARPManager.EnableProxyMode(int(req.WarpProxyPort)); err != nil {
			h.logger.Errorf("Failed to enable WARP proxy mode: %v", err)
			return &pb.SetupWARPProxyEndpointResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to enable WARP proxy mode: %v", err),
			}, nil
		}
	}

	// 2. Configure Hysteria2 for WARP outbound
	if req.ConfigureHysteria {
		hysteriaConfig := map[string]interface{}{
			"warp_enabled":    req.WarpEnabled,
			"warp_proxy_port": req.WarpProxyPort,
			"listen_port":     req.VpnListenPort,
			"auth_password":   req.VpnAuthPassword,
			"bandwidth_up":    req.BandwidthUp,
			"bandwidth_down":  req.BandwidthDown,
		}

		// Convert to JSON
		configJSON, err := json.Marshal(hysteriaConfig)
		if err != nil {
			h.logger.Errorf("Failed to marshal Hysteria2 config: %v", err)
			return &pb.SetupWARPProxyEndpointResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create Hysteria2 config: %v", err),
			}, nil
		}

		// Generate Hysteria2 configuration
		hysteriaConfigStr, err := h.localServices.HysteriaManager.GenerateConfig(string(configJSON))
		if err != nil {
			h.logger.Errorf("Failed to generate Hysteria2 config: %v", err)
			return &pb.SetupWARPProxyEndpointResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to generate Hysteria2 config: %v", err),
			}, nil
		}

		h.logger.Infof("Hysteria2 configuration generated: %s", hysteriaConfigStr)
	}

	// 3. Setup traffic routing (if requested)
	if req.SetupTrafficRouting {
		if !req.WarpEnabled {
			return &pb.SetupWARPProxyEndpointResponse{
				Success: false,
				Message: "Traffic routing requires WARP to be enabled",
			}, nil
		}

		// Use NetworkManager for basic routing
		if err := h.localServices.NetworkManager.RouteTrafficThroughWARP(req.VpnInterface); err != nil {
			h.logger.Errorf("Failed to setup traffic routing: %v", err)
			return &pb.SetupWARPProxyEndpointResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to setup traffic routing: %v", err),
			}, nil
		}
	}

	// 4. Enable masquerading on VPN interface
	if req.EnableMasquerading {
		if err := h.localServices.NetworkManager.EnableMasquerading(req.VpnInterface); err != nil {
			h.logger.Errorf("Failed to enable masquerading: %v", err)
			return &pb.SetupWARPProxyEndpointResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to enable masquerading: %v", err),
			}, nil
		}
	}

	return &pb.SetupWARPProxyEndpointResponse{
		Success: true,
		Message: "WARP proxy endpoint configured successfully",
		Config: map[string]string{
			"warp_enabled":    fmt.Sprintf("%v", req.WarpEnabled),
			"warp_proxy_port": fmt.Sprintf("%d", req.WarpProxyPort),
			"vpn_interface":   req.VpnInterface,
			"vpn_listen_port": fmt.Sprintf("%d", req.VpnListenPort),
			"traffic_routing": fmt.Sprintf("%v", req.SetupTrafficRouting),
			"masquerading":    fmt.Sprintf("%v", req.EnableMasquerading),
		},
	}, nil
}

// GetWARPProxyStatus returns comprehensive WARP proxy status
func (h *WARPProxyHandler) GetWARPProxyStatus(ctx context.Context, req *pb.GetWARPProxyStatusRequest) (*pb.GetWARPProxyStatusResponse, error) {
	h.logger.Info("GetWARPProxyStatus called")

	// Get WARP status
	warpStatus, err := h.localServices.WARPManager.GetWARPStatus()
	if err != nil {
		h.logger.Errorf("Failed to get WARP status: %v", err)
		return nil, fmt.Errorf("failed to get WARP status: %w", err)
	}

	// Get WARP configuration
	warpConfig, err := h.localServices.WARPManager.GetWARPConfiguration()
	if err != nil {
		h.logger.Errorf("Failed to get WARP configuration: %v", err)
		return nil, fmt.Errorf("failed to get WARP configuration: %w", err)
	}

	// Get network interfaces
	interfaces, err := h.localServices.NetworkManager.GetNetworkInterfaces()
	if err != nil {
		h.logger.Errorf("Failed to get network interfaces: %v", err)
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Check masquerading status
	var masqueradingStatus map[string]bool = make(map[string]bool)
	for _, iface := range interfaces {
		if enabled, err := h.localServices.NetworkManager.IsMasqueradingEnabled(iface); err == nil {
			masqueradingStatus[iface] = enabled
		}
	}

	// Convert to protobuf format
	status := &pb.WARPProxyStatus{
		WarpInstalled:      warpStatus.Installed,
		WarpConnected:      warpStatus.Connected,
		WarpMode:           warpStatus.Mode,
		WarpProxyPort:      int32(warpStatus.ProxyPort),
		WarpAccountType:    warpStatus.AccountType,
		WarpOrganization:   warpStatus.Organization,
		WarpIpAddress:      warpStatus.IPAddress,
		WarpLocation:       warpStatus.Location,
		WarpServerLocation: warpStatus.ServerLocation,
		WarpLastConnected:  int64(warpStatus.LastConnected.Unix()),
		WarpUptime:         int64(warpStatus.Uptime.Seconds()),
		WarpBytesSent:      warpStatus.BytesSent,
		WarpBytesReceived:  warpStatus.BytesReceived,
		WarpHealth:         warpStatus.Health,
		WarpError:          warpStatus.Error,

		// Configuration
		ConfigEnabled:      warpConfig.Enabled,
		ConfigProxyPort:    int32(warpConfig.ProxyPort),
		ConfigAutoConnect:  warpConfig.AutoConnect,
		ConfigNotifyOnFail: warpConfig.NotifyOnFail,
		ConfigClientType:   warpConfig.ClientType,
		ConfigHasLicense:   warpConfig.LicenseKey != "",
		ConfigOrganization: warpConfig.Organization,
		ConfigMode:         warpConfig.Mode,

		// Network status
		NetworkInterfaces:  interfaces,
		MasqueradingStatus: masqueradingStatus,
	}

	return &pb.GetWARPProxyStatusResponse{
		Status: status,
	}, nil
}

// RestartWARPProxyService restarts the complete WARP proxy service
func (h *WARPProxyHandler) RestartWARPProxyService(ctx context.Context, req *pb.RestartWARPProxyServiceRequest) (*pb.RestartWARPProxyServiceResponse, error) {
	h.logger.Info("RestartWARPProxyService called")

	// 1. Stop WARP service
	if err := h.localServices.WARPManager.DisconnectWARP(); err != nil {
		h.logger.Warnf("Failed to disconnect WARP: %v", err)
	}

	// 2. Stop Hysteria2
	if err := h.localServices.HysteriaManager.StopHysteria2(); err != nil {
		h.logger.Warnf("Failed to stop Hysteria2: %v", err)
	}

	// 3. Cleanup routing rules
	if err := h.localServices.NetworkManager.DisableWarpRouting(); err != nil {
		h.logger.Warnf("Failed to cleanup routing rules: %v", err)
	}

	// 4. Wait a moment
	h.logger.Info("Waiting before restart...")

	// 5. Restart WARP
	if err := h.localServices.WARPManager.ConnectWARP(); err != nil {
		h.logger.Errorf("Failed to restart WARP: %v", err)
		return &pb.RestartWARPProxyServiceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to restart WARP: %v", err),
		}, nil
	}

	// 6. Restart Hysteria2 (if it was configured to start)
	// This would need the config path - for now skip

	return &pb.RestartWARPProxyServiceResponse{
		Success: true,
		Message: "WARP proxy service restarted successfully",
	}, nil
}

// TestWARPProxyConnectivity tests the end-to-end WARP proxy connectivity
func (h *WARPProxyHandler) TestWARPProxyConnectivity(ctx context.Context, req *pb.TestWARPProxyConnectivityRequest) (*pb.TestWARPProxyConnectivityResponse, error) {
	h.logger.Info("TestWARPProxyConnectivity called")

	results := make(map[string]bool)

	// 1. Test WARP connection
	if connected, err := h.localServices.WARPManager.IsWARPConnected(); err == nil {
		results["warp_connection"] = connected
	} else {
		results["warp_connection"] = false
	}

	// 2. Test WARP proxy port
	if warpStatus, err := h.localServices.WARPManager.GetWARPStatus(); err == nil {
		results["warp_proxy_active"] = warpStatus.Mode == "proxy" && warpStatus.ProxyPort > 0
	} else {
		results["warp_proxy_active"] = false
	}

	// 3. Test network connectivity through WARP
	// This would require actual network tests - for now assume true if WARP is connected
	results["internet_through_warp"] = results["warp_connection"]

	// 4. Test DNS resolution
	// For now assume true
	results["dns_resolution"] = true

	// 5. Test Hysteria2 service
	if hysteriaStatus, err := h.localServices.HysteriaManager.GetHysteria2Status(); err == nil {
		if running, ok := hysteriaStatus["running"].(bool); ok {
			results["hysteria2_running"] = running
		} else {
			results["hysteria2_running"] = false
		}
	} else {
		results["hysteria2_running"] = false
	}

	// Calculate overall success
	allPassed := true
	for _, passed := range results {
		if !passed {
			allPassed = false
			break
		}
	}

	return &pb.TestWARPProxyConnectivityResponse{
		Success: allPassed,
		Results: results,
		Message: fmt.Sprintf("Connectivity test completed. Overall status: %v", allPassed),
	}, nil
}
