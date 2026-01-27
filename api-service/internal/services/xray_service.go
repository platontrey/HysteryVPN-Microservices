package services

import (
	"context"
	"fmt"

	"hysteria2_microservices/api-service/internal/models"
	repoInterfaces "hysteria2_microservices/api-service/internal/repositories/interfaces"
	serviceInterfaces "hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/google/uuid"
)

type XrayServiceImpl struct {
	xrayRepo   repoInterfaces.XrayConfigRepository
	userRepo   repoInterfaces.UserRepository
	deviceRepo repoInterfaces.DeviceRepository
	logger     *logger.Logger
}

func NewXrayService(
	xrayRepo repoInterfaces.XrayConfigRepository,
	userRepo repoInterfaces.UserRepository,
	deviceRepo repoInterfaces.DeviceRepository,
	logger *logger.Logger,
) serviceInterfaces.XrayService {
	return &XrayServiceImpl{
		xrayRepo:   xrayRepo,
		userRepo:   userRepo,
		deviceRepo: deviceRepo,
		logger:     logger,
	}
}

func (s *XrayServiceImpl) GenerateUserConfig(ctx context.Context, userID, deviceID string, protocol string) (*models.XrayConfig, error) {
	s.logger.Infof("Generating Xray config for user %s, device %s, protocol %s", userID, deviceID, protocol)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var deviceUUID *uuid.UUID
	if deviceID != "" {
		parsed, err := uuid.Parse(deviceID)
		if err != nil {
			return nil, fmt.Errorf("invalid device ID: %w", err)
		}
		deviceUUID = &parsed
	}

	// Validate protocol
	if !s.isValidProtocol(protocol) {
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	// Generate configuration based on protocol
	configData, err := s.generateProtocolConfig(protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to generate protocol config: %w", err)
	}

	config := &models.XrayConfig{
		UserID:     userUUID,
		DeviceID:   deviceUUID,
		ConfigName: fmt.Sprintf("%s-%s", protocol, userID[:8]),
		Protocol:   protocol,
		ConfigData: configData,
		IsActive:   true,
	}

	// Save to database
	if err := s.xrayRepo.Create(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Infof("Xray config generated successfully: %s", config.ID)
	return config, nil
}

func (s *XrayServiceImpl) UpdateUserConfig(ctx context.Context, userID, deviceID string, config *models.XrayConfig) error {
	s.logger.Infof("Updating Xray config for user %s, device %s", userID, deviceID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	var deviceUUID *uuid.UUID
	if deviceID != "" {
		parsed, err := uuid.Parse(deviceID)
		if err != nil {
			return fmt.Errorf("invalid device ID: %w", err)
		}
		deviceUUID = &parsed
	}

	// Validate ownership
	if config.UserID != userUUID || (deviceUUID != nil && config.DeviceID != nil && *config.DeviceID != *deviceUUID) {
		return fmt.Errorf("config does not belong to user/device")
	}

	// Validate protocol
	if !s.isValidProtocol(config.Protocol) {
		return fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}

	// Update config
	if err := s.xrayRepo.Update(ctx, config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	s.logger.Infof("Xray config updated successfully: %s", config.ID)
	return nil
}

func (s *XrayServiceImpl) GetSupportedProtocols(ctx context.Context) ([]string, error) {
	// Return supported protocols - in production this could be configurable
	return []string{"vless", "vless-reality", "vmess", "trojan", "shadowsocks"}, nil
}

func (s *XrayServiceImpl) ReloadConfiguration(ctx context.Context) error {
	s.logger.Info("Reloading Xray configuration")
	// Implementation would communicate with nodes to reload configs
	return fmt.Errorf("not implemented")
}

func (s *XrayServiceImpl) GetActiveConnections(ctx context.Context) ([]models.Connection, error) {
	s.logger.Info("Getting active Xray connections")
	// Implementation would query nodes for active connections
	return nil, fmt.Errorf("not implemented")
}

func (s *XrayServiceImpl) DisconnectUser(ctx context.Context, userID, deviceID string) error {
	s.logger.Infof("Disconnecting Xray user %s, device %s", userID, deviceID)
	// Implementation would send disconnect command to nodes
	return fmt.Errorf("not implemented")
}

func (s *XrayServiceImpl) GetServiceStatus(ctx context.Context) (*models.ServiceStatus, error) {
	s.logger.Info("Getting Xray service status")
	// Implementation would query nodes for service status
	return &models.ServiceStatus{
		IsRunning:         false,
		Version:           "unknown",
		Uptime:            0,
		ActiveConnections: 0,
		TotalTraffic:      0,
		MemoryUsage:       0,
		CPUUsage:          0,
	}, nil
}

func (s *XrayServiceImpl) isValidProtocol(protocol string) bool {
	validProtocols := []string{"vless", "vless-reality", "vmess", "trojan", "shadowsocks"}
	for _, p := range validProtocols {
		if p == protocol {
			return true
		}
	}
	return false
}

func (s *XrayServiceImpl) generateProtocolConfig(protocol string) (map[string]interface{}, error) {
	switch protocol {
	case "vless":
		return s.generateVLESSConfig()
	case "vless-reality":
		return s.generateVLESSRealityConfig()
	case "vmess":
		return s.generateVMessConfig()
	case "trojan":
		return s.generateTrojanConfig()
	case "shadowsocks":
		return s.generateShadowsocksConfig()
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func (s *XrayServiceImpl) generateVLESSConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"inbounds": []map[string]interface{}{
			{
				"port":     443,
				"protocol": "vless",
				"settings": map[string]interface{}{
					"clients": []map[string]interface{}{
						{
							"id":   uuid.New().String(),
							"flow": "xtls-rprx-vision",
						},
					},
					"decryption": "none",
				},
				"streamSettings": map[string]interface{}{
					"network":  "tcp",
					"security": "none",
				},
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"settings": map[string]interface{}{},
			},
		},
	}, nil
}

func (s *XrayServiceImpl) generateVLESSRealityConfig() (map[string]interface{}, error) {
	// Generate Reality keys (simplified - in production use proper crypto)
	privateKey := "example-private-key-base64"
	publicKey := "example-public-key-base64"

	return map[string]interface{}{
		"inbounds": []map[string]interface{}{
			{
				"port":     443,
				"protocol": "vless",
				"settings": map[string]interface{}{
					"clients": []map[string]interface{}{
						{
							"id":   uuid.New().String(),
							"flow": "xtls-rprx-vision",
						},
					},
					"decryption": "none",
				},
				"streamSettings": map[string]interface{}{
					"network":  "tcp",
					"security": "reality",
					"realitySettings": map[string]interface{}{
						"dest":        "www.example.com:443",
						"serverNames": []string{"www.example.com"},
						"privateKey":  privateKey,
						"publicKey":   publicKey,
						"shortIds":    []string{"abc123"},
					},
				},
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"settings": map[string]interface{}{},
			},
		},
	}, nil
}

func (s *XrayServiceImpl) generateRealityConfig() (map[string]interface{}, error) {
	// Deprecated - use VLESS + Reality instead
	return s.generateVLESSRealityConfig()
}

func (s *XrayServiceImpl) generateVMessConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"inbounds": []map[string]interface{}{
			{
				"port":     443,
				"protocol": "vmess",
				"settings": map[string]interface{}{
					"clients": []map[string]interface{}{
						{
							"id":      uuid.New().String(),
							"alterId": 0,
						},
					},
				},
				"streamSettings": map[string]interface{}{
					"network":  "tcp",
					"security": "none",
				},
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"settings": map[string]interface{}{},
			},
		},
	}, nil
}

func (s *XrayServiceImpl) generateTrojanConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"inbounds": []map[string]interface{}{
			{
				"port":     443,
				"protocol": "trojan",
				"settings": map[string]interface{}{
					"clients": []map[string]interface{}{
						{
							"password": "example-password",
						},
					},
				},
				"streamSettings": map[string]interface{}{
					"network":  "tcp",
					"security": "tls",
					"tlsSettings": map[string]interface{}{
						"certificates": []map[string]interface{}{
							{
								"certificateFile": "/etc/xray/cert.pem",
								"keyFile":         "/etc/xray/key.pem",
							},
						},
					},
				},
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"settings": map[string]interface{}{},
			},
		},
	}, nil
}

func (s *XrayServiceImpl) generateShadowsocksConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"inbounds": []map[string]interface{}{
			{
				"port":     443,
				"protocol": "shadowsocks",
				"settings": map[string]interface{}{
					"method":   "aes-256-gcm",
					"password": "example-password",
					"network":  "tcp,udp",
				},
				"streamSettings": map[string]interface{}{
					"network": "tcp",
				},
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"settings": map[string]interface{}{},
			},
		},
	}, nil
}
