package services

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/config"
)

// HysteriaManager handles Hysteria2 VPN server management
type HysteriaManager interface {
	InstallHysteria2() error
	IsHysteria2Installed() bool
	GenerateConfig(configTemplate string) (string, error)
	StartHysteria2(configPath string) error
	StopHysteria2() error
	RestartHysteria2(configPath string) error
	GetHysteria2Status() (map[string]interface{}, error)
	EnablePortHopping(startPort, endPort, interval int) error
	DisablePortHopping() error
	EnableSalamander(password string) error
	DisableSalamander() error

	// Advanced obfuscation methods for Russian DPI bypass
	EnableAdvancedObfuscation() error
	DisableAdvancedObfuscation() error
	EnableQUICObfuscation() error
	DisableQUICObfuscation() error
	ConfigureQUICScrambleTransform(enabled bool) error
	SetQUICPacketPadding(padding int) error
	EnableQUICTimingRandomization() error
	DisableQUICTimingRandomization() error
	EnableTLSFingerprintRotation(fingerprints []string) error
	DisableTLSFingerprintRotation() error
	EnableVLESSReality(targets []string) error
	DisableVLESSReality() error
	EnableTrafficShaping() error
	DisableTrafficShaping() error
	GetObfuscationStatus() (map[string]interface{}, error)

	// SNI-related methods
	ConfigureSNI(domains []string, defaultSNI string) error
	GenerateCertificatesForDomains(domains []string) error
	UpdateServerConfigWithSNI() (string, error)
	GetSNIStatus() (map[string]interface{}, error)
	EnableSNI() error
	DisableSNI() error
	AddSNIDomain(domain string) error
	RemoveSNIDomain(domain string) error

	// Let's Encrypt automation
	AutoConfigureSNICertificates(domains []string, email string) error
	SetupInitialCertificates(domains []string, email string) error
	EnableAutoRenewal() error
	DisableAutoRenewal() error
	ValidateAllDomains(domains []string) error
}

func (hm *HysteriaManagerImpl) ConfigureWARP(enabled bool, proxyPort int) error {
	hm.logger.Warnf("DEPRECATED: Use WARPManager.ConfigureWARP() instead")

	hm.config.Hysteria2.WARPEnabled = enabled
	if proxyPort > 0 {
		hm.config.Hysteria2.WARPProxyPort = proxyPort
	}

	// Validate configuration
	if err := hm.ValidateWARPConfiguration(); err != nil {
		return fmt.Errorf("WARP configuration validation failed: %w", err)
	}

	hm.logger.Infof("WARP configuration updated successfully (legacy method)")
	return nil
}

type HysteriaManagerImpl struct {
	logger             *logrus.Logger
	config             *config.Config
	certificateManager CertificateManager
}

// NewHysteriaManager creates a new HysteriaManager
func NewHysteriaManager(logger *logrus.Logger, cfg *config.Config) HysteriaManager {
	certManager := NewCertificateManager(logger, cfg)
	return &HysteriaManagerImpl{
		logger:             logger,
		config:             cfg,
		certificateManager: certManager,
	}
}

// InstallHysteria2 installs Hysteria2 using the official script
func (hm *HysteriaManagerImpl) InstallHysteria2() error {
	hm.logger.Info("Installing Hysteria2...")

	// Use the official install script
	cmd := exec.Command("bash", "-c", "bash <(curl -fsSL https://get.hy2.sh/)")
	output, err := cmd.CombinedOutput()
	if err != nil {
		hm.logger.Errorf("Failed to install Hysteria2: %v, output: %s", err, string(output))
		return fmt.Errorf("failed to install Hysteria2: %w", err)
	}

	hm.logger.Info("Hysteria2 installed successfully")
	return nil
}

// IsHysteria2Installed checks if Hysteria2 is installed
func (hm *HysteriaManagerImpl) IsHysteria2Installed() bool {
	cmd := exec.Command("which", "hysteria")
	err := cmd.Run()
	return err == nil
}

// GenerateConfig generates Hysteria2 configuration based on template
func (hm *HysteriaManagerImpl) GenerateConfig(configTemplate string) (string, error) {
	hm.logger.Info("Generating Hysteria2 configuration")

	var hysteriaConfig map[string]interface{}

	if configTemplate != "" {
		// Parse custom config template
		err := json.Unmarshal([]byte(configTemplate), &hysteriaConfig)
		if err != nil {
			return "", fmt.Errorf("invalid config template: %w", err)
		}
	} else {
		// Generate default config
		hysteriaConfig = hm.generateDefaultConfig()
	}

	// Apply configuration options
	hm.applyConfigOptions(hysteriaConfig)

	// Convert to JSON
	configJSON, err := json.MarshalIndent(hysteriaConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	return string(configJSON), nil
}

func (hm *HysteriaManagerImpl) generateDefaultConfig() map[string]interface{} {
	config := map[string]interface{}{
		"listen": fmt.Sprintf(":%d", hm.config.Hysteria2.DefaultListenPort),
		"tls": map[string]interface{}{
			"cert": "/etc/hysteria/cert.pem",
			"key":  "/etc/hysteria/key.pem",
		},
		"auth": map[string]interface{}{
			"type":     hm.config.Hysteria2.AuthType,
			"password": hm.config.Hysteria2.AuthPassword,
		},
		"bandwidth": map[string]interface{}{
			"up":   fmt.Sprintf("%d mbps", hm.config.Hysteria2.UpMbps),
			"down": fmt.Sprintf("%d mbps", hm.config.Hysteria2.DownMbps),
		},
	}

	// Configure based on WARP settings
	if hm.config.Hysteria2.WARPEnabled {
		// Configure outbound proxy through WARP
		warpPort := hm.config.Hysteria2.WARPProxyPort
		if warpPort == 0 {
			warpPort = 1080 // default
		}

		// Create WARP outbound configuration
		outboundConfig := map[string]interface{}{
			"name": "warp-proxy",
			"type": "socks5",
			"addr": fmt.Sprintf("127.0.0.1:%d", warpPort),
		}

		// Add additional WARP-specific settings
		if hm.config.Hysteria2.WARPOrganization != "" {
			outboundConfig["organization"] = hm.config.Hysteria2.WARPOrganization
		}

		// Configure traffic routing rules for WARP
		config["outbound"] = outboundConfig

		// Configure ACL rules for traffic through WARP
		config["acl"] = map[string]interface{}{
			"file": "/etc/hysteria/acl.yaml",
		}

		// WARP mode uses custom routing instead of traditional masquerade
		// Traffic will be routed at system level through iptables
		hm.logger.Info("WARP outbound proxy configured - all traffic will route through WARP")
	} else if !hm.config.Hysteria2.SalamanderEnabled {
		// Fallback to traditional masquerade
		config["masquerade"] = map[string]interface{}{
			"type": "proxy",
			"proxy": map[string]interface{}{
				"url":         "https://www.google.com",
				"rewriteHost": true,
			},
		}
		hm.logger.Info("Traditional masquerade configured")
	}

	return config
}

func (hm *HysteriaManagerImpl) applyConfigOptions(config map[string]interface{}) {
	// Apply WARP configuration first (highest priority)
	if hm.config.Hysteria2.WARPEnabled {
		// Remove masquerade when WARP is enabled
		delete(config, "masquerade")
		// WARP is mutually exclusive with Salamander obfuscation
		if hm.config.Hysteria2.SalamanderEnabled {
			hm.logger.Warn("WARP and Salamander obfuscation are mutually exclusive. WARP takes priority.")
			hm.config.Hysteria2.SalamanderEnabled = false
			delete(config, "obfs")
		}
		hm.logger.Info("WARP enabled - traditional masquerade and obfuscation disabled")
	} else {
		// Apply Salamander obfuscation (mutually exclusive with masquerade)
		if hm.config.Hysteria2.SalamanderEnabled {
			config["obfs"] = map[string]interface{}{
				"type":     "salamander",
				"password": hm.config.Hysteria2.SalamanderPassword,
			}
			// Remove masquerade if obfs is enabled
			delete(config, "masquerade")
			hm.logger.Info("Obfuscation enabled - masquerade disabled for compatibility")
		} else {
			// Apply default masquerade only if obfs is not enabled and WARP is not enabled
			if _, exists := config["masquerade"]; !exists {
				config["masquerade"] = map[string]interface{}{
					"type": "proxy",
					"proxy": map[string]interface{}{
						"url":         "https://www.google.com",
						"rewriteHost": true,
					},
				}
				hm.logger.Info("Masquerade configured - obfuscation and WARP disabled")
			}
		}
	}

	// Apply Port Hopping
	if hm.config.Hysteria2.PortHopping {
		config["hopping"] = map[string]interface{}{
			"interval": hm.config.Hysteria2.HopInterval,
			"start":    hm.config.Hysteria2.HopStartPort,
			"end":      hm.config.Hysteria2.HopEndPort,
		}
	}
}

// StartHysteria2 starts the Hysteria2 service
func (hm *HysteriaManagerImpl) StartHysteria2(configPath string) error {
	hm.logger.Infof("Starting Hysteria2 with config: %s", configPath)

	if hm.config.Hysteria2.EnableSystemd {
		return hm.startWithSystemd(configPath)
	}

	// Start directly
	cmd := exec.Command("hysteria", "server", "-c", configPath)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start Hysteria2: %w", err)
	}

	hm.logger.Info("Hysteria2 started successfully")
	return nil
}

func (hm *HysteriaManagerImpl) startWithSystemd(configPath string) error {
	// Create systemd service file
	serviceContent := fmt.Sprintf(`[Unit]
Description=Hysteria2 VPN Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/hysteria server -c %s
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, configPath)

	err := hm.runCommand("echo", serviceContent, "|", "tee", "/etc/systemd/system/hysteria2.service")
	if err != nil {
		return fmt.Errorf("failed to create systemd service: %w", err)
	}

	// Reload systemd and start service
	if err := hm.runCommand("systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := hm.runCommand("systemctl", "enable", "hysteria2"); err != nil {
		return fmt.Errorf("failed to enable hysteria2 service: %w", err)
	}

	if err := hm.runCommand("systemctl", "start", "hysteria2"); err != nil {
		return fmt.Errorf("failed to start hysteria2 service: %w", err)
	}

	hm.logger.Info("Hysteria2 started with systemd")
	return nil
}

// StopHysteria2 stops the Hysteria2 service
func (hm *HysteriaManagerImpl) StopHysteria2() error {
	hm.logger.Info("Stopping Hysteria2")

	if hm.config.Hysteria2.EnableSystemd {
		return hm.runCommand("systemctl", "stop", "hysteria2")
	}

	// Kill process directly (not ideal, but for demo)
	return hm.runCommand("pkill", "-f", "hysteria")
}

// RestartHysteria2 restarts the Hysteria2 service
func (hm *HysteriaManagerImpl) RestartHysteria2(configPath string) error {
	hm.logger.Info("Restarting Hysteria2")

	if hm.config.Hysteria2.EnableSystemd {
		return hm.runCommand("systemctl", "restart", "hysteria2")
	}

	if err := hm.StopHysteria2(); err != nil {
		hm.logger.Warnf("Failed to stop Hysteria2: %v", err)
	}

	return hm.StartHysteria2(configPath)
}

// GetHysteria2Status returns Hysteria2 service status
func (hm *HysteriaManagerImpl) GetHysteria2Status() (map[string]interface{}, error) {
	status := map[string]interface{}{
		"installed": hm.IsHysteria2Installed(),
		"running":   false,
		"systemd":   hm.config.Hysteria2.EnableSystemd,
	}

	if hm.config.Hysteria2.EnableSystemd {
		// Check systemd status
		cmd := exec.Command("systemctl", "is-active", "hysteria2")
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "active" {
			status["running"] = true
		}
	} else {
		// Check if process is running
		cmd := exec.Command("pgrep", "-f", "hysteria")
		err := cmd.Run()
		status["running"] = err == nil
	}

	return status, nil
}

// EnablePortHopping enables port hopping
func (hm *HysteriaManagerImpl) EnablePortHopping(startPort, endPort, interval int) error {
	hm.logger.Infof("Enabling port hopping: %d-%d every %d seconds", startPort, endPort, interval)

	// This would require updating the config and restarting
	// For now, just log
	hm.config.Hysteria2.PortHopping = true
	hm.config.Hysteria2.HopStartPort = startPort
	hm.config.Hysteria2.HopEndPort = endPort
	hm.config.Hysteria2.HopInterval = interval

	return nil
}

// DisablePortHopping disables port hopping
func (hm *HysteriaManagerImpl) DisablePortHopping() error {
	hm.logger.Info("Disabling port hopping")

	hm.config.Hysteria2.PortHopping = false
	return nil
}

// EnableSalamander enables Salamander obfuscation and disables masquerade
func (hm *HysteriaManagerImpl) EnableSalamander(password string) error {
	hm.logger.Info("Enabling Salamander obfuscation")

	hm.config.Hysteria2.SalamanderEnabled = true
	hm.config.Hysteria2.SalamanderPassword = password

	// Note: Masquerade will be removed during config generation in applyConfigOptions
	hm.logger.Info("Salamander obfuscation enabled - masquerade will be disabled in configuration")
	return nil
}

// DisableSalamander disables Salamander obfuscation and re-enables masquerade
func (hm *HysteriaManagerImpl) DisableSalamander() error {
	hm.logger.Info("Disabling Salamander obfuscation")

	hm.config.Hysteria2.SalamanderEnabled = false

	// Masquerade will be re-enabled in applyConfigOptions
	hm.logger.Info("Salamander obfuscation disabled - masquerade will be re-enabled in configuration")
	return nil
}

// runCommand executes a system command
func (hm *HysteriaManagerImpl) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	hm.logger.Debugf("Running command: %s %v", name, args)
	return cmd.Run()
}

// SNI-related methods implementation

// ConfigureSNI sets up SNI for multiple domains
func (hm *HysteriaManagerImpl) ConfigureSNI(domains []string, defaultSNI string) error {
	hm.logger.Infof("Configuring SNI for domains: %v", domains)

	// Update configuration
	hm.config.Hysteria2.SNIEnabled = true
	hm.config.Hysteria2.SNIDomains = domains
	if defaultSNI != "" {
		hm.config.Hysteria2.DefaultSNI = defaultSNI
	} else if len(domains) > 0 {
		hm.config.Hysteria2.DefaultSNI = domains[0]
	}

	// Generate certificates for domains if in auto mode
	if hm.config.Hysteria2.SNIAutoMode {
		if err := hm.GenerateCertificatesForDomains(domains); err != nil {
			return fmt.Errorf("failed to generate certificates: %w", err)
		}
	}

	hm.logger.Infof("SNI configured successfully with %d domains", len(domains))
	return nil
}

// GenerateCertificatesForDomains creates SSL certificates for SNI domains
func (hm *HysteriaManagerImpl) GenerateCertificatesForDomains(domains []string) error {
	hm.logger.Infof("Generating certificates for %d domains", len(domains))

	for _, domain := range domains {
		if domain == "" {
			continue
		}

		// Check if certificate already exists and is valid
		if valid, err := hm.certificateManager.ValidateCertificate(domain); valid && err == nil {
			hm.logger.Infof("Certificate already valid for domain: %s", domain)
			continue
		}

		// Generate self-signed certificate
		certPath, _, err := hm.certificateManager.GenerateSelfSignedCert(domain)
		if err != nil {
			hm.logger.Errorf("Failed to generate certificate for %s: %v", domain, err)
			return fmt.Errorf("failed to generate certificate for %s: %w", domain, err)
		}

		hm.logger.Infof("Certificate generated for %s: %s", domain, certPath)
	}

	return nil
}

// UpdateServerConfigWithSNI generates Hysteria2 config with SNI support
func (hm *HysteriaManagerImpl) UpdateServerConfigWithSNI() (string, error) {
	hm.logger.Info("Generating Hysteria2 configuration with SNI support")

	var hysteriaConfig map[string]interface{}

	// Generate base config
	hysteriaConfig = hm.generateDefaultConfig()

	// Apply configuration options
	hm.applyConfigOptions(hysteriaConfig)

	// Apply SNI configuration
	if hm.config.Hysteria2.SNIEnabled {
		sniConfig := hm.buildSNIConfig()
		if len(sniConfig) > 0 {
			hysteriaConfig["sni"] = sniConfig
		}
	}

	// Validate mutual exclusivity
	if err := hm.validateConfig(hysteriaConfig); err != nil {
		return "", fmt.Errorf("configuration validation failed: %w", err)
	}

	// Convert to JSON
	configJSON, err := json.MarshalIndent(hysteriaConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	return string(configJSON), nil
}

// validateConfig checks for configuration conflicts and invalid combinations
func (hm *HysteriaManagerImpl) validateConfig(config map[string]interface{}) error {
	// Check for obfs and masquerade conflict
	_, hasObfs := config["obfs"]
	_, hasMasquerade := config["masquerade"]

	if hasObfs && hasMasquerade {
		return fmt.Errorf("obfuscation and masquerade cannot be enabled simultaneously")
	}

	// Log configuration state
	if hasObfs {
		hm.logger.Info("Configuration validation passed: Obfuscation enabled, masquerade disabled")
	} else if hasMasquerade {
		hm.logger.Info("Configuration validation passed: Masquerade enabled, obfuscation disabled")
	} else {
		hm.logger.Info("Configuration validation passed: Neither obfuscation nor masquerade enabled")
	}

	return nil
}

// buildSNIConfig creates SNI configuration structure
func (hm *HysteriaManagerImpl) buildSNIConfig() map[string]interface{} {
	sniConfig := map[string]interface{}{
		"enabled": true,
		"domains": make([]map[string]interface{}, 0),
	}

	for _, domain := range hm.config.Hysteria2.SNIDomains {
		if domain == "" {
			continue
		}

		domainConfig := map[string]interface{}{
			"domain": domain,
			"cert":   fmt.Sprintf("/etc/hysteria/sni/%s.crt", domain),
			"key":    fmt.Sprintf("/etc/hysteria/sni/%s.key", domain),
		}
		sniConfig["domains"] = append(sniConfig["domains"].([]map[string]interface{}), domainConfig)
	}

	if hm.config.Hysteria2.DefaultSNI != "" {
		sniConfig["default"] = hm.config.Hysteria2.DefaultSNI
	}

	return sniConfig
}

// GetSNIStatus returns current SNI configuration status
func (hm *HysteriaManagerImpl) GetSNIStatus() (map[string]interface{}, error) {
	status := map[string]interface{}{
		"enabled":     hm.config.Hysteria2.SNIEnabled,
		"domains":     hm.config.Hysteria2.SNIDomains,
		"default_sni": hm.config.Hysteria2.DefaultSNI,
		"auto_mode":   hm.config.Hysteria2.SNIAutoMode,
		"auto_renew":  hm.config.Hysteria2.SNIAutoRenew,
		"cert_dir":    hm.config.Hysteria2.SNICertPath,
	}

	// Get certificate information if SNI is enabled
	if hm.config.Hysteria2.SNIEnabled {
		certificates, err := hm.certificateManager.ListCertificates()
		if err != nil {
			hm.logger.Warnf("Failed to list certificates: %v", err)
		} else {
			status["certificates"] = certificates
		}

		// Check for expiring certificates
		expiringSoon := make([]string, 0)
		for _, domain := range hm.config.Hysteria2.SNIDomains {
			if expiring, err := hm.certificateManager.IsCertificateExpiringSoon(domain, 30); err == nil && expiring {
				expiringSoon = append(expiringSoon, domain)
			}
		}
		status["expiring_soon"] = expiringSoon
	}

	return status, nil
}

// EnableSNI enables SNI functionality
func (hm *HysteriaManagerImpl) EnableSNI() error {
	hm.config.Hysteria2.SNIEnabled = true
	hm.logger.Info("SNI functionality enabled")
	return nil
}

// DisableSNI disables SNI functionality
func (hm *HysteriaManagerImpl) DisableSNI() error {
	hm.config.Hysteria2.SNIEnabled = false
	hm.logger.Info("SNI functionality disabled")
	return nil
}

// AddSNIDomain adds a new domain to SNI configuration
func (hm *HysteriaManagerImpl) AddSNIDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Check if domain already exists
	for _, existingDomain := range hm.config.Hysteria2.SNIDomains {
		if existingDomain == domain {
			return fmt.Errorf("domain %s already exists in SNI configuration", domain)
		}
	}

	// Add domain
	hm.config.Hysteria2.SNIDomains = append(hm.config.Hysteria2.SNIDomains, domain)

	// Generate certificate if in auto mode
	if hm.config.Hysteria2.SNIAutoMode {
		if _, _, err := hm.certificateManager.GenerateSelfSignedCert(domain); err != nil {
			hm.logger.Errorf("Failed to generate certificate for new domain %s: %v", domain, err)
			return fmt.Errorf("failed to generate certificate for %s: %w", domain, err)
		}
	}

	hm.logger.Infof("Domain %s added to SNI configuration", domain)
	return nil
}

// RemoveSNIDomain removes a domain from SNI configuration
func (hm *HysteriaManagerImpl) RemoveSNIDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Find and remove domain
	for i, existingDomain := range hm.config.Hysteria2.SNIDomains {
		if existingDomain == domain {
			hm.config.Hysteria2.SNIDomains = append(
				hm.config.Hysteria2.SNIDomains[:i],
				hm.config.Hysteria2.SNIDomains[i+1:]...,
			)

			// Remove certificate files
			if err := hm.certificateManager.DeleteCertificate(domain); err != nil {
				hm.logger.Warnf("Failed to delete certificate for %s: %v", domain, err)
			}

			// Update default SNI if necessary
			if hm.config.Hysteria2.DefaultSNI == domain && len(hm.config.Hysteria2.SNIDomains) > 0 {
				hm.config.Hysteria2.DefaultSNI = hm.config.Hysteria2.SNIDomains[0]
			} else if len(hm.config.Hysteria2.SNIDomains) == 0 {
				hm.config.Hysteria2.DefaultSNI = ""
			}

			hm.logger.Infof("Domain %s removed from SNI configuration", domain)
			return nil
		}
	}

	return fmt.Errorf("domain %s not found in SNI configuration", domain)
}

// ===== LET'S ENCRYPT AUTOMATION METHODS =====

// AutoConfigureSNICertificates automatically configures SNI with Let's Encrypt certificates
func (hm *HysteriaManagerImpl) AutoConfigureSNICertificates(domains []string, email string) error {
	hm.logger.Infof("Auto-configuring SNI with Let's Encrypt certificates for %d domains", len(domains))

	if len(domains) == 0 {
		return fmt.Errorf("at least one domain is required")
	}

	if email == "" {
		return fmt.Errorf("email is required for Let's Encrypt certificates")
	}

	// Check if Let's Encrypt is available
	if !hm.certificateManager.IsLetsEncryptEnabled() {
		hm.logger.Info("Let's Encrypt not available, installing certbot...")
		if err := hm.certificateManager.InstallCertbot(); err != nil {
			hm.logger.Warnf("Failed to install certbot, falling back to self-signed certificates: %v", err)
			return hm.GenerateCertificatesForDomains(domains)
		}
	}

	// Validate all domains first
	if err := hm.ValidateAllDomains(domains); err != nil {
		return fmt.Errorf("domain validation failed: %w", err)
	}

	// Generate Let's Encrypt certificates for all domains
	for _, domain := range domains {
		if domain == "" {
			continue
		}

		hm.logger.Infof("Generating Let's Encrypt certificate for domain: %s", domain)

		certPath, _, err := hm.certificateManager.GenerateLetsEncryptCert(domain, email, "http-01")
		if err != nil {
			hm.logger.Errorf("Failed to generate Let's Encrypt certificate for %s, falling back to self-signed: %v", domain, err)

			// Fallback to self-signed certificate
			certPath, _, err = hm.certificateManager.GenerateSelfSignedCert(domain)
			if err != nil {
				return fmt.Errorf("failed to generate fallback certificate for %s: %w", domain, err)
			}
		}

		hm.logger.Infof("Certificate generated for %s: %s", domain, certPath)
	}

	// Configure SNI with the domains
	defaultSNI := ""
	if len(domains) > 0 {
		defaultSNI = domains[0]
	}
	if err := hm.ConfigureSNI(domains, defaultSNI); err != nil {
		return fmt.Errorf("failed to configure SNI: %w", err)
	}

	// Enable auto-renewal
	if err := hm.EnableAutoRenewal(); err != nil {
		hm.logger.Warnf("Failed to enable auto-renewal: %v", err)
	}

	hm.logger.Infof("SNI auto-configuration completed successfully for %d domains", len(domains))
	return nil
}

// SetupInitialCertificates sets up certificates during initial node installation
func (hm *HysteriaManagerImpl) SetupInitialCertificates(domains []string, email string) error {
	hm.logger.Info("Setting up initial certificates for node")

	// If no domains provided, use server IP for self-signed cert
	if len(domains) == 0 {
		hm.logger.Info("No domains provided, using server IP for self-signed certificate")
		serverIP := hm.getServerIP()
		if serverIP != "" {
			return hm.GenerateCertificatesForDomains([]string{serverIP})
		}
		return fmt.Errorf("no domains provided and could not determine server IP")
	}

	// Try Let's Encrypt first, fallback to self-signed
	if email != "" && hm.certificateManager.IsLetsEncryptEnabled() {
		err := hm.AutoConfigureSNICertificates(domains, email)
		if err == nil {
			return nil
		}
		hm.logger.Warnf("Let's Encrypt configuration failed, using self-signed certificates: %v", err)
	}

	return hm.GenerateCertificatesForDomains(domains)
}

// EnableAutoRenewal sets up automatic certificate renewal
func (hm *HysteriaManagerImpl) EnableAutoRenewal() error {
	hm.logger.Info("Enabling automatic certificate renewal")

	if !hm.certificateManager.IsLetsEncryptEnabled() {
		return fmt.Errorf("Let's Encrypt not available, cannot enable auto-renewal")
	}

	// Create cron job for renewal (simple approach using systemd timer)
	cronJob := `# Auto-renew Let's Encrypt certificates
0 3 * * * root /usr/local/bin/hysteria-agent renew-certificates >> /var/log/hysteria-renewal.log 2>&1`

	cronPath := "/etc/cron.d/hysteria-cert-renewal"
	if err := os.WriteFile(cronPath, []byte(cronJob), 0644); err != nil {
		return fmt.Errorf("failed to create cron job: %w", err)
	}

	// Reload cron service
	if err := hm.runCommand("systemctl", "reload", "cron"); err != nil {
		hm.logger.Warnf("Failed to reload cron service: %v", err)
	}

	hm.logger.Info("Automatic certificate renewal enabled")
	return nil
}

// DisableAutoRenewal disables automatic certificate renewal
func (hm *HysteriaManagerImpl) DisableAutoRenewal() error {
	hm.logger.Info("Disabling automatic certificate renewal")

	cronPath := "/etc/cron.d/hysteria-cert-renewal"
	if err := os.Remove(cronPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cron job: %w", err)
	}

	if err := hm.runCommand("systemctl", "reload", "cron"); err != nil {
		hm.logger.Warnf("Failed to reload cron service: %v", err)
	}

	hm.logger.Info("Automatic certificate renewal disabled")
	return nil
}

// ValidateAllDomains validates multiple domains
func (hm *HysteriaManagerImpl) ValidateAllDomains(domains []string) error {
	hm.logger.Infof("Validating %d domains", len(domains))

	var validationErrors []string

	for _, domain := range domains {
		if domain == "" {
			continue
		}

		// Use certificate manager to validate domain ownership
		if valid, err := hm.certificateManager.ValidateDomainOwnership(domain); !valid {
			errMsg := fmt.Sprintf("Domain %s validation failed: %v", domain, err)
			validationErrors = append(validationErrors, errMsg)
			hm.logger.Warn(errMsg)
		} else {
			hm.logger.Infof("Domain %s validation passed", domain)
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("domain validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// getServerIP gets the server's main IP address
func (hm *HysteriaManagerImpl) getServerIP() string {
	// This is a simplified implementation
	// In a real environment, you'd want to get the actual server IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return ""
}

// ===== WARP-SPECIFIC METHODS =====

// ConfigureWARP enables or disables WARP configuration
func (hm *HysteriaManagerImpl) ConfigureWARP(enabled bool, proxyPort int) error {
	hm.logger.Infof("Configuring WARP - enabled: %v, proxy port: %d", enabled, proxyPort)

	hm.config.Hysteria2.WARPEnabled = enabled
	if proxyPort > 0 {
		hm.config.Hysteria2.WARPProxyPort = proxyPort
	}

	// Validate configuration
	if err := hm.ValidateWARPConfiguration(); err != nil {
		return fmt.Errorf("WARP configuration validation failed: %w", err)
	}

	hm.logger.Infof("WARP configuration updated successfully")
	return nil
}

// IsWARPEnabled returns whether WARP is enabled in configuration
func (hm *HysteriaManagerImpl) IsWARPEnabled() bool {
	return hm.config.Hysteria2.WARPEnabled
}

// GetWARPConfiguration returns current WARP configuration
func (hm *HysteriaManagerImpl) GetWARPConfiguration() (map[string]interface{}, error) {
	config := map[string]interface{}{
		"enabled":      hm.config.Hysteria2.WARPEnabled,
		"proxy_port":   hm.config.Hysteria2.WARPProxyPort,
		"auto_connect": hm.config.Hysteria2.WARPAutoConnect,
		"notify_fail":  hm.config.Hysteria2.WARPNotifyOnFail,
		"client_type":  hm.config.Hysteria2.WARPClientType,
		"organization": hm.config.Hysteria2.WARPOrganization,
	}

	// Add license key info only if present
	if hm.config.Hysteria2.WARPLicenseKey != "" {
		config["has_license"] = true
	} else {
		config["has_license"] = false
	}

	return config, nil
}

// ===== ADVANCED OBFUSCATION METHODS FOR RUSSIAN DPI BYPASS =====

// EnableAdvancedObfuscation enables all advanced obfuscation features
func (hm *HysteriaManagerImpl) EnableAdvancedObfuscation() error {
	hm.logger.Info("Enabling advanced obfuscation for Russian DPI bypass")

	hm.config.Hysteria2.AdvancedObfuscationEnabled = true

	// Enable core obfuscation features
	if err := hm.EnableQUICObfuscation(); err != nil {
		hm.logger.Warnf("Failed to enable QUIC obfuscation: %v", err)
	}
	if err := hm.EnableTLSFingerprintRotation([]string{"chrome", "firefox", "safari"}); err != nil {
		hm.logger.Warnf("Failed to enable TLS fingerprint rotation: %v", err)
	}
	if err := hm.EnableTrafficShaping(); err != nil {
		hm.logger.Warnf("Failed to enable traffic shaping: %v", err)
	}

	hm.logger.Info("Advanced obfuscation enabled successfully")
	return nil
}

// DisableAdvancedObfuscation disables all advanced obfuscation features
func (hm *HysteriaManagerImpl) DisableAdvancedObfuscation() error {
	hm.logger.Info("Disabling advanced obfuscation")

	hm.config.Hysteria2.AdvancedObfuscationEnabled = false

	// Disable all obfuscation features
	hm.DisableQUICObfuscation()
	hm.DisableTLSFingerprintRotation()
	hm.DisableTrafficShaping()
	hm.DisableVLESSReality()

	hm.logger.Info("Advanced obfuscation disabled successfully")
	return nil
}

// EnableQUICObfuscation enables QUIC-level obfuscation
func (hm *HysteriaManagerImpl) EnableQUICObfuscation() error {
	hm.logger.Info("Enabling QUIC obfuscation")

	hm.config.Hysteria2.QUICObfuscationEnabled = true
	hm.config.Hysteria2.QUICScrambleTransform = true
	hm.config.Hysteria2.QUICPacketPadding = 1300
	hm.config.Hysteria2.QUICTimingRandomization = true

	hm.logger.Info("QUIC obfuscation enabled")
	return nil
}

// DisableQUICObfuscation disables QUIC-level obfuscation
func (hm *HysteriaManagerImpl) DisableQUICObfuscation() error {
	hm.logger.Info("Disabling QUIC obfuscation")

	hm.config.Hysteria2.QUICObfuscationEnabled = false
	hm.config.Hysteria2.QUICScrambleTransform = false
	hm.config.Hysteria2.QUICTimingRandomization = false

	hm.logger.Info("QUIC obfuscation disabled")
	return nil
}

// ConfigureQUICScrambleTransform configures QUIC scramble transform
func (hm *HysteriaManagerImpl) ConfigureQUICScrambleTransform(enabled bool) error {
	hm.logger.Infof("Configuring QUIC scramble transform: %v", enabled)
	hm.config.Hysteria2.QUICScrambleTransform = enabled
	return nil
}

// SetQUICPacketPadding sets QUIC packet padding size
func (hm *HysteriaManagerImpl) SetQUICPacketPadding(padding int) error {
	if padding < 1200 || padding > 1500 {
		return fmt.Errorf("packet padding must be between 1200-1500 bytes, got %d", padding)
	}
	hm.logger.Infof("Setting QUIC packet padding to %d bytes", padding)
	hm.config.Hysteria2.QUICPacketPadding = padding
	return nil
}

// EnableQUICTimingRandomization enables QUIC timing randomization
func (hm *HysteriaManagerImpl) EnableQUICTimingRandomization() error {
	hm.logger.Info("Enabling QUIC timing randomization")
	hm.config.Hysteria2.QUICTimingRandomization = true
	return nil
}

// DisableQUICTimingRandomization disables QUIC timing randomization
func (hm *HysteriaManagerImpl) DisableQUICTimingRandomization() error {
	hm.logger.Info("Disabling QUIC timing randomization")
	hm.config.Hysteria2.QUICTimingRandomization = false
	return nil
}

// EnableTLSFingerprintRotation enables TLS fingerprint rotation
func (hm *HysteriaManagerImpl) EnableTLSFingerprintRotation(fingerprints []string) error {
	if len(fingerprints) == 0 {
		return fmt.Errorf("at least one fingerprint must be provided")
	}
	hm.logger.Infof("Enabling TLS fingerprint rotation with fingerprints: %v", fingerprints)
	hm.config.Hysteria2.TLSFingerprintRotation = true
	hm.config.Hysteria2.TLSFingerprints = fingerprints
	return nil
}

// DisableTLSFingerprintRotation disables TLS fingerprint rotation
func (hm *HysteriaManagerImpl) DisableTLSFingerprintRotation() error {
	hm.logger.Info("Disabling TLS fingerprint rotation")
	hm.config.Hysteria2.TLSFingerprintRotation = false
	return nil
}

// EnableVLESSReality enables VLESS Reality protocol
func (hm *HysteriaManagerImpl) EnableVLESSReality(targets []string) error {
	if len(targets) == 0 {
		return fmt.Errorf("at least one target domain must be provided")
	}
	hm.logger.Infof("Enabling VLESS Reality with targets: %v", targets)
	hm.config.Hysteria2.VLESSRealityEnabled = true
	hm.config.Hysteria2.VLESSRealityTargets = targets
	return nil
}

// DisableVLESSReality disables VLESS Reality protocol
func (hm *HysteriaManagerImpl) DisableVLESSReality() error {
	hm.logger.Info("Disabling VLESS Reality")
	hm.config.Hysteria2.VLESSRealityEnabled = false
	return nil
}

// EnableTrafficShaping enables traffic shaping and behavioral randomization
func (hm *HysteriaManagerImpl) EnableTrafficShaping() error {
	hm.logger.Info("Enabling traffic shaping and behavioral randomization")
	hm.config.Hysteria2.TrafficShapingEnabled = true
	hm.config.Hysteria2.BehavioralRandomization = true
	return nil
}

// DisableTrafficShaping disables traffic shaping and behavioral randomization
func (hm *HysteriaManagerImpl) DisableTrafficShaping() error {
	hm.logger.Info("Disabling traffic shaping and behavioral randomization")
	hm.config.Hysteria2.TrafficShapingEnabled = false
	hm.config.Hysteria2.BehavioralRandomization = false
	return nil
}

// GetObfuscationStatus returns current obfuscation configuration status
func (hm *HysteriaManagerImpl) GetObfuscationStatus() (map[string]interface{}, error) {
	status := map[string]interface{}{
		"advanced_obfuscation_enabled": hm.config.Hysteria2.AdvancedObfuscationEnabled,
		"quic_obfuscation": map[string]interface{}{
			"enabled":              hm.config.Hysteria2.QUICObfuscationEnabled,
			"scramble_transform":   hm.config.Hysteria2.QUICScrambleTransform,
			"packet_padding":       hm.config.Hysteria2.QUICPacketPadding,
			"timing_randomization": hm.config.Hysteria2.QUICTimingRandomization,
		},
		"tls_fingerprint": map[string]interface{}{
			"rotation_enabled": hm.config.Hysteria2.TLSFingerprintRotation,
			"fingerprints":     hm.config.Hysteria2.TLSFingerprints,
		},
		"vless_reality": map[string]interface{}{
			"enabled": hm.config.Hysteria2.VLESSRealityEnabled,
			"targets": hm.config.Hysteria2.VLESSRealityTargets,
		},
		"traffic_shaping": map[string]interface{}{
			"enabled":                  hm.config.Hysteria2.TrafficShapingEnabled,
			"behavioral_randomization": hm.config.Hysteria2.BehavioralRandomization,
		},
		"multi_hop_enabled": hm.config.Hysteria2.MultiHopEnabled,
	}
	return status, nil
}

// ValidateWARPConfiguration validates WARP configuration settings
func (hm *HysteriaManagerImpl) ValidateWARPConfiguration() error {
	if !hm.config.Hysteria2.WARPEnabled {
		return nil // No validation needed if disabled
	}

	// Check for conflicts with Salamander
	if hm.config.Hysteria2.SalamanderEnabled {
		return fmt.Errorf("WARP cannot be enabled simultaneously with Salamander obfuscation")
	}

	// Validate proxy port
	if hm.config.Hysteria2.WARPProxyPort <= 0 || hm.config.Hysteria2.WARPProxyPort > 65535 {
		return fmt.Errorf("invalid WARP proxy port: %d (must be 1-65535)", hm.config.Hysteria2.WARPProxyPort)
	}

	// Validate client type
	validClientTypes := []string{"local", "docker"}
	isValidType := false
	for _, validType := range validClientTypes {
		if hm.config.Hysteria2.WARPClientType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid WARP client type: %s (valid types: %v)", hm.config.Hysteria2.WARPClientType, validClientTypes)
	}

	hm.logger.Info("WARP configuration validation passed")
	return nil
}
