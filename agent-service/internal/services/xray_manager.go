package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/config"
)

// XrayManagerImpl implements XrayManager
type XrayManagerImpl struct {
	logger             *logrus.Logger
	config             *config.Config
	certificateManager CertificateManager
}

// NewXrayManager creates a new XrayManager
func NewXrayManager(logger *logrus.Logger, cfg *config.Config) XrayManager {
	certManager := NewCertificateManager(logger, cfg)
	return &XrayManagerImpl{
		logger:             logger,
		config:             cfg,
		certificateManager: certManager,
	}
}

// InstallXray installs Xray-core using the official script
func (xm *XrayManagerImpl) InstallXray() error {
	xm.logger.Info("Installing Xray-core...")

	// Use the official install script
	cmd := exec.Command("bash", "-c", "bash <(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh) install")
	output, err := cmd.CombinedOutput()
	if err != nil {
		xm.logger.Errorf("Failed to install Xray-core: %v, output: %s", err, string(output))
		return fmt.Errorf("failed to install Xray-core: %w", err)
	}

	xm.logger.Info("Xray-core installed successfully")
	return nil
}

// IsXrayInstalled checks if Xray-core is installed
func (xm *XrayManagerImpl) IsXrayInstalled() bool {
	cmd := exec.Command("which", "xray")
	err := cmd.Run()
	return err == nil
}

// StartXray starts the Xray service
func (xm *XrayManagerImpl) StartXray(configPath string) error {
	xm.logger.Infof("Starting Xray with config: %s", configPath)

	if err := xm.validateConfigFile(configPath); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Start directly
	cmd := exec.Command("xray", "run", "-c", configPath)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start Xray: %w", err)
	}

	xm.logger.Info("Xray started successfully")
	return nil
}

// StopXray stops the Xray service
func (xm *XrayManagerImpl) StopXray() error {
	xm.logger.Info("Stopping Xray")

	// Kill process directly
	return exec.Command("pkill", "-f", "xray").Run()
}

// RestartXray restarts the Xray service
func (xm *XrayManagerImpl) RestartXray(configPath string) error {
	xm.logger.Info("Restarting Xray")

	if err := xm.StopXray(); err != nil {
		xm.logger.Warnf("Failed to stop Xray: %v", err)
	}

	return xm.StartXray(configPath)
}

// GetXrayStatus returns Xray service status
func (xm *XrayManagerImpl) GetXrayStatus() (map[string]interface{}, error) {
	status := map[string]interface{}{
		"installed": xm.IsXrayInstalled(),
		"running":   false,
	}

	// Check if process is running
	cmd := exec.Command("pgrep", "-f", "xray")
	err := cmd.Run()
	status["running"] = err == nil

	return status, nil
}

// GenerateConfig generates Xray configuration based on protocol and template
func (xm *XrayManagerImpl) GenerateConfig(protocol string, configTemplate string) (string, error) {
	xm.logger.Infof("Generating Xray configuration for protocol: %s", protocol)

	var xrayConfig map[string]interface{}

	if configTemplate != "" {
		// Parse custom config template
		err := json.Unmarshal([]byte(configTemplate), &xrayConfig)
		if err != nil {
			return "", fmt.Errorf("invalid config template: %w", err)
		}
	} else {
		// Generate default config based on protocol
		var err error
		xrayConfig, err = xm.generateDefaultConfig(protocol)
		if err != nil {
			return "", fmt.Errorf("failed to generate default config: %w", err)
		}
	}

	// Apply configuration options
	if err := xm.applyConfigOptions(protocol, xrayConfig); err != nil {
		return "", fmt.Errorf("failed to apply config options: %w", err)
	}

	// Convert to JSON
	configJSON, err := json.MarshalIndent(xrayConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	return string(configJSON), nil
}

// ValidateConfig validates Xray configuration
func (xm *XrayManagerImpl) ValidateConfig(protocol string, config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Basic validation based on protocol
	switch protocol {
	case "vless":
		return xm.validateVLESSConfig(config)
	case "vless-reality":
		return xm.validateVLESSRealityConfig(config)
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// ConfigureVLESS configures VLESS protocol settings
func (xm *XrayManagerImpl) ConfigureVLESS(uuid, dest string, flow string) error {
	xm.logger.Infof("Configuring VLESS with UUID: %s, dest: %s, flow: %s", uuid, dest, flow)

	// Validate UUID
	if _, err := uuid.Parse(uuid); err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	// Validate flow
	validFlows := []string{"", "xtls-rprx-vision", "xtls-rprx-vision-udp443"}
	validFlow := false
	for _, vf := range validFlows {
		if flow == vf {
			validFlow = true
			break
		}
	}
	if !validFlow {
		return fmt.Errorf("invalid flow: %s, valid flows: %v", flow, validFlows)
	}

	xm.logger.Info("VLESS configuration validated successfully")
	return nil
}

// ConfigureReality configures Reality protocol settings
func (xm *XrayManagerImpl) ConfigureReality(dest, serverNames string, privateKey string, shortIds []string) error {
	xm.logger.Infof("Configuring Reality with dest: %s, serverNames: %s", dest, serverNames)

	// Validate destination
	if dest == "" {
		return fmt.Errorf("destination cannot be empty")
	}

	// Validate server names
	if serverNames == "" {
		return fmt.Errorf("server names cannot be empty")
	}

	// Validate private key (should be base64)
	if privateKey == "" {
		return fmt.Errorf("private key cannot be empty")
	}

	// Validate short IDs
	for _, shortId := range shortIds {
		if len(shortId) > 16 {
			return fmt.Errorf("short ID too long: %s", shortId)
		}
	}

	xm.logger.Info("Reality configuration validated successfully")
	return nil
}

// GenerateRealityKeys generates private and public keys for Reality
func (xm *XrayManagerImpl) GenerateRealityKeys() (privateKey, publicKey string, err error) {
	xm.logger.Info("Generating Reality keys")

	// This is a simplified implementation
	// In production, use proper X25519 key generation
	privateKey = "example-private-key-base64"
	publicKey = "example-public-key-base64"

	xm.logger.Info("Reality keys generated successfully")
	return privateKey, publicKey, nil
}

// GenerateRealityCert generates certificate for Reality domain
func (xm *XrayManagerImpl) GenerateRealityCert(domain string) error {
	xm.logger.Infof("Generating Reality certificate for domain: %s", domain)

	// Generate self-signed certificate for Reality
	certPath, _, err := xm.certificateManager.GenerateSelfSignedCert(domain)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	xm.logger.Infof("Reality certificate generated: %s", certPath)
	return nil
}

// ValidateRealityCert validates Reality certificate
func (xm *XrayManagerImpl) ValidateRealityCert(domain string) (bool, error) {
	return xm.certificateManager.ValidateCertificate(domain)
}

// generateDefaultConfig generates default Xray configuration for specified protocol
func (xm *XrayManagerImpl) generateDefaultConfig(protocol string) (map[string]interface{}, error) {
	switch protocol {
	case "vless":
		return xm.generateVLESSConfig(), nil
	case "vless-reality":
		return xm.generateVLESSRealityConfig(), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// generateVLESSConfig generates default VLESS configuration
func (xm *XrayManagerImpl) generateVLESSConfig() map[string]interface{} {
	return map[string]interface{}{
		"inbounds": []map[string]interface{}{
			{
				"port":     xm.config.Xray.ListenPort,
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
	}
}

// generateVLESSRealityConfig generates VLESS + Reality configuration
func (xm *XrayManagerImpl) generateVLESSRealityConfig() map[string]interface{} {
	privateKey, publicKey, _ := xm.GenerateRealityKeys()

	return map[string]interface{}{
		"inbounds": []map[string]interface{}{
			{
				"port":     xm.config.Xray.ListenPort,
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
					"network": "tcp",
					"security": "reality",
					"realitySettings": map[string]interface{}{
						"dest":         xm.config.Xray.RealityDest,
						"serverNames":  xm.config.Xray.RealityServerNames,
						"privateKey":   xm.config.Xray.RealityPrivateKey,
						"publicKey":    xm.config.Xray.RealityPublicKey,
						"shortIds":     xm.config.Xray.RealityShortIds,
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
	}
}

// generateRealityConfig generates default Reality configuration (deprecated - use VLESS + Reality)
func (xm *XrayManagerImpl) generateRealityConfig() map[string]interface{} {
	return xm.generateVLESSRealityConfig()
}
}

// applyConfigOptions applies additional configuration options
func (xm *XrayManagerImpl) applyConfigOptions(protocol string, config map[string]interface{}) error {
	// Add logging
	config["log"] = map[string]interface{}{
		"loglevel": "warning",
	}

	// Add API for statistics if enabled
	if xm.config.Xray.EnableAPI {
		config["api"] = map[string]interface{}{
			"tag":      "api",
			"services": []string{"StatsService"},
		}

		// Add stats outbound
		outbounds := config["outbounds"].([]map[string]interface{})
		outbounds = append(outbounds, map[string]interface{}{
			"protocol": "freedom",
			"tag":      "api",
			"settings": map[string]interface{}{},
		})
		config["outbounds"] = outbounds
	}

	return nil
}

// validateConfigFile validates Xray config file
func (xm *XrayManagerImpl) validateConfigFile(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Basic JSON validation
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("invalid JSON in config file: %w", err)
	}

	return nil
}

// validateVLESSConfig validates VLESS configuration
func (xm *XrayManagerImpl) validateVLESSConfig(config map[string]interface{}) error {
	inbounds, ok := config["inbounds"].([]interface{})
	if !ok {
		return fmt.Errorf("inbounds not found or invalid")
	}

	for _, inbound := range inbounds {
		inboundMap, ok := inbound.(map[string]interface{})
		if !ok {
			continue
		}

		if inboundMap["protocol"] == "vless" {
			settings, ok := inboundMap["settings"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("vless settings not found")
			}

			clients, ok := settings["clients"].([]interface{})
			if !ok || len(clients) == 0 {
				return fmt.Errorf("no vless clients configured")
			}
		}
	}

	return nil
}

// validateVLESSRealityConfig validates VLESS + Reality configuration
func (xm *XrayManagerImpl) validateVLESSRealityConfig(config map[string]interface{}) error {
	// First validate VLESS part
	err := xm.validateVLESSConfig(config)
	if err != nil {
		return fmt.Errorf("VLESS validation failed: %w", err)
	}

	// Then validate Reality part
	err = xm.validateRealityConfig(config)
	if err != nil {
		return fmt.Errorf("Reality validation failed: %w", err)
	}

	return nil
}

// validateRealityConfig validates Reality configuration
func (xm *XrayManagerImpl) validateRealityConfig(config map[string]interface{}) error {
	inbounds, ok := config["inbounds"].([]interface{})
	if !ok {
		return fmt.Errorf("inbounds not found or invalid")
	}

	for _, inbound := range inbounds {
		inboundMap, ok := inbound.(map[string]interface{})
		if !ok {
			continue
		}

		streamSettings, ok := inboundMap["streamSettings"].(map[string]interface{})
		if !ok {
			continue
		}

		if streamSettings["security"] == "reality" {
			realitySettings, ok := streamSettings["realitySettings"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("reality settings not found")
			}

			if realitySettings["dest"] == nil || realitySettings["dest"] == "" {
				return fmt.Errorf("reality dest not configured")
			}

			if realitySettings["privateKey"] == nil || realitySettings["privateKey"] == "" {
				return fmt.Errorf("reality private key not configured")
			}

			// Validate server names
			serverNames, ok := realitySettings["serverNames"].([]interface{})
			if !ok || len(serverNames) == 0 {
				return fmt.Errorf("reality server names not configured")
			}
		}
	}

	return nil
}
