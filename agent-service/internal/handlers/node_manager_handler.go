package handlers

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/services"
	pb "hysteria2_microservices/proto"
)

// NodeManagerHandler implements the NodeManager gRPC service
type NodeManagerHandler struct {
	pb.UnimplementedNodeManagerServer
	localServices *services.LocalServices
	logger        *logrus.Logger
}

// NewNodeManagerHandler creates a new NodeManagerHandler
func NewNodeManagerHandler(localServices *services.LocalServices, logger *logrus.Logger) *NodeManagerHandler {
	return &NodeManagerHandler{
		localServices: localServices,
		logger:        logger,
	}
}

// EnableMasquerading enables IP masquerading on the specified interface
func (h *NodeManagerHandler) EnableMasquerading(ctx context.Context, req *pb.EnableMasqueradingRequest) (*pb.EnableMasqueradingResponse, error) {
	h.logger.Infof("EnableMasquerading called for interface: %s", req.InterfaceName)

	err := h.localServices.NetworkManager.EnableMasquerading(req.InterfaceName)
	if err != nil {
		h.logger.Errorf("Failed to enable masquerading: %v", err)
		return &pb.EnableMasqueradingResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to enable masquerading: %v", err),
		}, nil
	}

	return &pb.EnableMasqueradingResponse{
		Success: true,
		Message: "Masquerading enabled successfully",
	}, nil
}

// DisableMasquerading disables IP masquerading on the specified interface
func (h *NodeManagerHandler) DisableMasquerading(ctx context.Context, req *pb.DisableMasqueradingRequest) (*pb.DisableMasqueradingResponse, error) {
	h.logger.Infof("DisableMasquerading called for interface: %s", req.InterfaceName)

	err := h.localServices.NetworkManager.DisableMasquerading(req.InterfaceName)
	if err != nil {
		h.logger.Errorf("Failed to disable masquerading: %v", err)
		return &pb.DisableMasqueradingResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to disable masquerading: %v", err),
		}, nil
	}

	return &pb.DisableMasqueradingResponse{
		Success: true,
		Message: "Masquerading disabled successfully",
	}, nil
}

// GetNetworkInterfaces returns available network interfaces
func (h *NodeManagerHandler) GetNetworkInterfaces(ctx context.Context, req *pb.GetNetworkInterfacesRequest) (*pb.GetNetworkInterfacesResponse, error) {
	h.logger.Info("GetNetworkInterfaces called")

	interfaces, err := h.localServices.NetworkManager.GetNetworkInterfaces()
	if err != nil {
		h.logger.Errorf("Failed to get network interfaces: %v", err)
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// For now, assume eth0 as default if available
	defaultInterface := "eth0"
	for _, iface := range interfaces {
		if iface == "eth0" {
			defaultInterface = "eth0"
			break
		}
	}

	return &pb.GetNetworkInterfacesResponse{
		Interfaces:       interfaces,
		DefaultInterface: defaultInterface,
	}, nil
}

// IsMasqueradingEnabled checks if masquerading is enabled on the interface
func (h *NodeManagerHandler) IsMasqueradingEnabled(ctx context.Context, req *pb.IsMasqueradingEnabledRequest) (*pb.IsMasqueradingEnabledResponse, error) {
	h.logger.Infof("IsMasqueradingEnabled called for interface: %s", req.InterfaceName)

	enabled, err := h.localServices.NetworkManager.IsMasqueradingEnabled(req.InterfaceName)
	if err != nil {
		h.logger.Errorf("Failed to check masquerading status: %v", err)
		return nil, fmt.Errorf("failed to check masquerading status: %w", err)
	}

	return &pb.IsMasqueradingEnabledResponse{
		Enabled: enabled,
	}, nil
}

// Other methods (placeholders for now)
func (h *NodeManagerHandler) UpdateConfig(ctx context.Context, req *pb.ConfigUpdateRequest) (*pb.ConfigUpdateResponse, error) {
	return &pb.ConfigUpdateResponse{Success: false, Message: "Not implemented"}, nil
}

func (h *NodeManagerHandler) ReloadConfig(ctx context.Context, req *pb.ReloadRequest) (*pb.ReloadResponse, error) {
	return &pb.ReloadResponse{Success: false, Message: "Not implemented"}, nil
}

func (h *NodeManagerHandler) GetStatus(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	return &pb.StatusResponse{}, nil
}

func (h *NodeManagerHandler) AddUser(ctx context.Context, req *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	return &pb.AddUserResponse{Success: false, Message: "Not implemented"}, nil
}

func (h *NodeManagerHandler) RemoveUser(ctx context.Context, req *pb.RemoveUserRequest) (*pb.RemoveUserResponse, error) {
	return &pb.RemoveUserResponse{Success: false, Message: "Not implemented"}, nil
}

func (h *NodeManagerHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return &pb.UpdateUserResponse{Success: false, Message: "Not implemented"}, nil
}

func (h *NodeManagerHandler) GetMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsResponse, error) {
	return &pb.MetricsResponse{}, nil
}

func (h *NodeManagerHandler) StreamMetrics(req *pb.StreamMetricsRequest, stream pb.NodeManager_StreamMetricsServer) error {
	return fmt.Errorf("not implemented")
}

func (h *NodeManagerHandler) RestartServer(ctx context.Context, req *pb.RestartRequest) (*pb.RestartResponse, error) {
	return &pb.RestartResponse{Success: false, Message: "Not implemented"}, nil
}

func (h *NodeManagerHandler) GetLogs(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	return &pb.LogResponse{Success: false, Logs: []string{"Not implemented"}}, nil
}

// Hysteria2 management methods

// InstallHysteria2 installs Hysteria2
func (h *NodeManagerHandler) InstallHysteria2(ctx context.Context, req *pb.InstallHysteria2Request) (*pb.InstallHysteria2Response, error) {
	h.logger.Info("InstallHysteria2 called")

	err := h.localServices.HysteriaManager.InstallHysteria2()
	if err != nil {
		h.logger.Errorf("Failed to install Hysteria2: %v", err)
		return &pb.InstallHysteria2Response{
			Success: false,
			Message: fmt.Sprintf("Failed to install Hysteria2: %v", err),
		}, nil
	}

	return &pb.InstallHysteria2Response{
		Success: true,
		Message: "Hysteria2 installed successfully",
	}, nil
}

// ConfigureHysteria2 configures Hysteria2 with given options
func (h *NodeManagerHandler) ConfigureHysteria2(ctx context.Context, req *pb.ConfigureHysteria2Request) (*pb.ConfigureHysteria2Response, error) {
	h.logger.Info("ConfigureHysteria2 called")

	config, err := h.localServices.HysteriaManager.GenerateConfig(req.ConfigTemplate)
	if err != nil {
		h.logger.Errorf("Failed to generate Hysteria2 config: %v", err)
		return &pb.ConfigureHysteria2Response{
			Success: false,
			Message: fmt.Sprintf("Failed to generate config: %v", err),
		}, nil
	}

	// Save config to file
	configPath := "/etc/hysteria/config.json"
	err = os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		h.logger.Errorf("Failed to save config: %v", err)
		return &pb.ConfigureHysteria2Response{
			Success: false,
			Message: fmt.Sprintf("Failed to save config: %v", err),
		}, nil
	}

	return &pb.ConfigureHysteria2Response{
		Success:         true,
		Message:         "Hysteria2 configured successfully",
		ConfigPath:      configPath,
		GeneratedConfig: config,
	}, nil
}

// StartHysteria2 starts Hysteria2 service
func (h *NodeManagerHandler) StartHysteria2(ctx context.Context, req *pb.StartHysteria2Request) (*pb.StartHysteria2Response, error) {
	h.logger.Info("StartHysteria2 called")

	err := h.localServices.HysteriaManager.StartHysteria2(req.ConfigPath)
	if err != nil {
		h.logger.Errorf("Failed to start Hysteria2: %v", err)
		return &pb.StartHysteria2Response{
			Success: false,
			Message: fmt.Sprintf("Failed to start Hysteria2: %v", err),
		}, nil
	}

	return &pb.StartHysteria2Response{
		Success: true,
		Message: "Hysteria2 started successfully",
	}, nil
}

// StopHysteria2 stops Hysteria2 service
func (h *NodeManagerHandler) StopHysteria2(ctx context.Context, req *pb.StopHysteria2Request) (*pb.StopHysteria2Response, error) {
	h.logger.Info("StopHysteria2 called")

	err := h.localServices.HysteriaManager.StopHysteria2()
	if err != nil {
		h.logger.Errorf("Failed to stop Hysteria2: %v", err)
		return &pb.StopHysteria2Response{
			Success: false,
			Message: fmt.Sprintf("Failed to stop Hysteria2: %v", err),
		}, nil
	}

	return &pb.StopHysteria2Response{
		Success: true,
		Message: "Hysteria2 stopped successfully",
	}, nil
}

// GetHysteria2Status returns Hysteria2 status
func (h *NodeManagerHandler) GetHysteria2Status(ctx context.Context, req *pb.GetHysteria2StatusRequest) (*pb.GetHysteria2StatusResponse, error) {
	h.logger.Info("GetHysteria2Status called")

	status, err := h.localServices.HysteriaManager.GetHysteria2Status()
	if err != nil {
		h.logger.Errorf("Failed to get Hysteria2 status: %v", err)
		return nil, fmt.Errorf("failed to get Hysteria2 status: %w", err)
	}

	// Convert to protobuf types
	var statusMap map[string]string
	for k, v := range status {
		if str, ok := v.(string); ok {
			statusMap[k] = str
		} else if b, ok := v.(bool); ok {
			statusMap[k] = fmt.Sprintf("%t", b)
		}
	}

	return &pb.GetHysteria2StatusResponse{
		Status: statusMap,
	}, nil
}

// EnablePortHopping enables port hopping
func (h *NodeManagerHandler) EnablePortHopping(ctx context.Context, req *pb.EnablePortHoppingRequest) (*pb.EnablePortHoppingResponse, error) {
	h.logger.Infof("EnablePortHopping called: %d-%d every %d", req.StartPort, req.EndPort, req.Interval)

	err := h.localServices.HysteriaManager.EnablePortHopping(int(req.StartPort), int(req.EndPort), int(req.Interval))
	if err != nil {
		h.logger.Errorf("Failed to enable port hopping: %v", err)
		return &pb.EnablePortHoppingResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to enable port hopping: %v", err),
		}, nil
	}

	return &pb.EnablePortHoppingResponse{
		Success: true,
		Message: "Port hopping enabled successfully",
	}, nil
}

// EnableSalamander enables Salamander obfuscation
func (h *NodeManagerHandler) EnableSalamander(ctx context.Context, req *pb.EnableSalamanderRequest) (*pb.EnableSalamanderResponse, error) {
	h.logger.Info("EnableSalamander called")

	err := h.localServices.HysteriaManager.EnableSalamander(req.Password)
	if err != nil {
		h.logger.Errorf("Failed to enable Salamander: %v", err)
		return &pb.EnableSalamanderResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to enable Salamander: %v", err),
		}, nil
	}

	return &pb.EnableSalamanderResponse{
		Success: true,
		Message: "Salamander obfuscation enabled successfully",
	}, nil
}

// WARP management methods

// InstallWARPClient installs Cloudflare WARP client
func (h *NodeManagerHandler) InstallWARPClient(ctx context.Context, req *pb.InstallWARPClientRequest) (*pb.InstallWARPClientResponse, error) {
	h.logger.Info("InstallWARPClient called")

	err := h.localServices.WARPManager.InstallWARPClient()
	if err != nil {
		h.logger.Errorf("Failed to install WARP client: %v", err)
		return &pb.InstallWARPClientResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to install WARP client: %v", err),
		}, nil
	}

	return &pb.InstallWARPClientResponse{
		Success: true,
		Message: "WARP client installed successfully",
	}, nil
}

// ConfigureWARP configures WARP settings
func (h *NodeManagerHandler) ConfigureWARP(ctx context.Context, req *pb.ConfigureWARPRequest) (*pb.ConfigureWARPResponse, error) {
	h.logger.Infof("ConfigureWARP called with enabled: %v, proxy port: %d", req.Enabled, req.ProxyPort)

	// Build WARP config
	config := services.WARPConfig{
		Enabled:      req.Enabled,
		ProxyPort:    int(req.ProxyPort),
		AutoConnect:  req.AutoConnect,
		NotifyOnFail: req.NotifyOnFail,
		ClientType:   req.ClientType,
		LicenseKey:   req.LicenseKey,
		Organization: req.Organization,
		Mode:         req.Mode,
	}

	err := h.localServices.WARPManager.ConfigureWARP(config)
	if err != nil {
		h.logger.Errorf("Failed to configure WARP: %v", err)
		return &pb.ConfigureWARPResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to configure WARP: %v", err),
		}, nil
	}

	return &pb.ConfigureWARPResponse{
		Success: true,
		Message: "WARP configured successfully",
	}, nil
}

// ConnectWARP connects to WARP network
func (h *NodeManagerHandler) ConnectWARP(ctx context.Context, req *pb.ConnectWARPRequest) (*pb.ConnectWARPResponse, error) {
	h.logger.Info("ConnectWARP called")

	err := h.localServices.WARPManager.ConnectWARP()
	if err != nil {
		h.logger.Errorf("Failed to connect to WARP: %v", err)
		return &pb.ConnectWARPResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to connect to WARP: %v", err),
		}, nil
	}

	return &pb.ConnectWARPResponse{
		Success: true,
		Message: "Connected to WARP successfully",
	}, nil
}

// DisconnectWARP disconnects from WARP network
func (h *NodeManagerHandler) DisconnectWARP(ctx context.Context, req *pb.DisconnectWARPRequest) (*pb.DisconnectWARPResponse, error) {
	h.logger.Info("DisconnectWARP called")

	err := h.localServices.WARPManager.DisconnectWARP()
	if err != nil {
		h.logger.Errorf("Failed to disconnect from WARP: %v", err)
		return &pb.DisconnectWARPResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to disconnect from WARP: %v", err),
		}, nil
	}

	return &pb.DisconnectWARPResponse{
		Success: true,
		Message: "Disconnected from WARP successfully",
	}, nil
}

// GetWARPStatus returns WARP connection status
func (h *NodeManagerHandler) GetWARPStatus(ctx context.Context, req *pb.GetWARPStatusRequest) (*pb.GetWARPStatusResponse, error) {
	h.logger.Info("GetWARPStatus called")

	status, err := h.localServices.WARPManager.GetWARPStatus()
	if err != nil {
		h.logger.Errorf("Failed to get WARP status: %v", err)
		return nil, fmt.Errorf("failed to get WARP status: %w", err)
	}

	// Convert to protobuf format
	pbStatus := &pb.WARPStatus{
		Installed:      status.Installed,
		Connected:      status.Connected,
		Mode:           status.Mode,
		ProxyPort:      int32(status.ProxyPort),
		AccountType:    status.AccountType,
		Organization:   status.Organization,
		IpAddress:      status.IPAddress,
		Location:       status.Location,
		ServerLocation: status.ServerLocation,
		LastConnected:  status.LastConnected.Unix(),
		Uptime:         int64(status.Uptime.Seconds()),
		BytesSent:      status.BytesSent,
		BytesReceived:  status.BytesReceived,
		Health:         status.Health,
		Error:          status.Error,
	}

	return &pb.GetWARPStatusResponse{
		Status: pbStatus,
	}, nil
}

// EnableWARPProxy enables WARP proxy mode
func (h *NodeManagerHandler) EnableWARPProxy(ctx context.Context, req *pb.EnableWARPProxyRequest) (*pb.EnableWARPProxyResponse, error) {
	h.logger.Infof("EnableWARPProxy called on port: %d", req.Port)

	err := h.localServices.WARPManager.EnableProxyMode(int(req.Port))
	if err != nil {
		h.logger.Errorf("Failed to enable WARP proxy: %v", err)
		return &pb.EnableWARPProxyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to enable WARP proxy: %v", err),
		}, nil
	}

	return &pb.EnableWARPProxyResponse{
		Success: true,
		Message: "WARP proxy mode enabled successfully",
	}, nil
}

// DisableWARPProxy disables WARP proxy mode
func (h *NodeManagerHandler) DisableWARPProxy(ctx context.Context, req *pb.DisableWARPProxyRequest) (*pb.DisableWARPProxyResponse, error) {
	h.logger.Info("DisableWARPProxy called")

	err := h.localServices.WARPManager.DisableProxyMode()
	if err != nil {
		h.logger.Errorf("Failed to disable WARP proxy: %v", err)
		return &pb.DisableWARPProxyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to disable WARP proxy: %v", err),
		}, nil
	}

	return &pb.DisableWARPProxyResponse{
		Success: true,
		Message: "WARP proxy mode disabled successfully",
	}, nil
}

// EnableWARPTrafficRouting enables traffic routing through WARP
func (h *NodeManagerHandler) EnableWARPTrafficRouting(ctx context.Context, req *pb.EnableWARPTrafficRoutingRequest) (*pb.EnableWARPTrafficRoutingResponse, error) {
	h.logger.Infof("EnableWARPTrafficRouting called for interface: %s", req.InterfaceName)

	err := h.localServices.WARPManager.EnableTrafficRouting(req.InterfaceName)
	if err != nil {
		h.logger.Errorf("Failed to enable WARP traffic routing: %v", err)
		return &pb.EnableWARPTrafficRoutingResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to enable WARP traffic routing: %v", err),
		}, nil
	}

	return &pb.EnableWARPTrafficRoutingResponse{
		Success: true,
		Message: "WARP traffic routing enabled successfully",
	}, nil
}

// DisableWARPTrafficRouting disables WARP traffic routing
func (h *NodeManagerHandler) DisableWARPTrafficRouting(ctx context.Context, req *pb.DisableWARPTrafficRoutingRequest) (*pb.DisableWARPTrafficRoutingResponse, error) {
	h.logger.Info("DisableWARPTrafficRouting called")

	err := h.localServices.WARPManager.DisableTrafficRouting()
	if err != nil {
		h.logger.Errorf("Failed to disable WARP traffic routing: %v", err)
		return &pb.DisableWARPTrafficRoutingResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to disable WARP traffic routing: %v", err),
		}, nil
	}

	return &pb.DisableWARPTrafficRoutingResponse{
		Success: true,
		Message: "WARP traffic routing disabled successfully",
	}, nil
}
