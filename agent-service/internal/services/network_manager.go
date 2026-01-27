package services

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/config"
)

// NetworkManagerImpl implements NetworkManager interface
type NetworkManagerImpl struct {
	logger *logrus.Logger
	config *config.Config
}

// NewNetworkManager creates a new NetworkManager
func NewNetworkManager(logger *logrus.Logger, cfg *config.Config) NetworkManager {
	return &NetworkManagerImpl{
		logger: logger,
		config: cfg,
	}
}

// EnableMasquerading enables IP masquerading on the specified interface
func (nm *NetworkManagerImpl) EnableMasquerading(interfaceName string) error {
	nm.logger.Infof("Enabling masquerading on interface: %s", interfaceName)

	// Enable IP forwarding
	if err := nm.runCommand("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		nm.logger.Errorf("Failed to enable IP forwarding: %v", err)
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Add iptables rule for masquerading
	cmd := fmt.Sprintf("-t nat -A POSTROUTING -o %s -j MASQUERADE", interfaceName)
	if err := nm.runCommand("iptables", strings.Fields(cmd)...); err != nil {
		nm.logger.Errorf("Failed to add masquerading rule: %v", err)
		return fmt.Errorf("failed to add masquerading rule: %w", err)
	}

	nm.logger.Infof("Masquerading enabled successfully on interface: %s", interfaceName)
	return nil
}

// DisableMasquerading disables IP masquerading on the specified interface
func (nm *NetworkManagerImpl) DisableMasquerading(interfaceName string) error {
	nm.logger.Infof("Disabling masquerading on interface: %s", interfaceName)

	// Remove iptables rule for masquerading
	cmd := fmt.Sprintf("-t nat -D POSTROUTING -o %s -j MASQUERADE", interfaceName)
	if err := nm.runCommand("iptables", strings.Fields(cmd)...); err != nil {
		nm.logger.Errorf("Failed to remove masquerading rule: %v", err)
		return fmt.Errorf("failed to remove masquerading rule: %w", err)
	}

	nm.logger.Infof("Masquerading disabled successfully on interface: %s", interfaceName)
	return nil
}

// IsMasqueradingEnabled checks if masquerading is enabled on the specified interface
func (nm *NetworkManagerImpl) IsMasqueradingEnabled(interfaceName string) (bool, error) {
	cmd := fmt.Sprintf("-t nat -C POSTROUTING -o %s -j MASQUERADE", interfaceName)
	err := nm.runCommand("iptables", strings.Fields(cmd)...)
	if err != nil {
		// If the rule doesn't exist, iptables -C returns exit code 1
		return false, nil
	}
	return true, nil
}

// GetNetworkInterfaces returns a list of available network interfaces
func (nm *NetworkManagerImpl) GetNetworkInterfaces() ([]string, error) {
	output, err := nm.runCommandWithOutput("ip", "link", "show")
	if err != nil {
		nm.logger.Errorf("Failed to get network interfaces: %v", err)
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var interfaces []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ": ") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				ifaceName := strings.TrimSpace(parts[1])
				if ifaceName != "" && ifaceName != "lo" {
					interfaces = append(interfaces, ifaceName)
				}
			}
		}
	}

	return interfaces, nil
}

// runCommand executes a system command and returns error if any
func (nm *NetworkManagerImpl) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	nm.logger.Debugf("Running command: %s %v", name, args)
	return cmd.Run()
}

// runCommandWithOutput executes a system command and returns its output
func (nm *NetworkManagerImpl) runCommandWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	nm.logger.Debugf("Running command with output: %s %v", name, args)
	output, err := cmd.Output()
	return string(output), err
}

// ===== WARP CLIENT MANAGEMENT =====

// InstallWARPClient installs Cloudflare WARP client
func (nm *NetworkManagerImpl) InstallWARPClient() error {
	nm.logger.Info("Installing Cloudflare WARP client...")

	// Check if WARP is already installed
	if nm.isWARPInstalled() {
		nm.logger.Info("WARP client is already installed")
		return nil
	}

	// Add Cloudflare repository and install
	commands := [][]string{
		{"curl", "-fsSL", "https://pkg.cloudflareclient.com/pubkey.gpg", "|", "gpg", "--yes", "--dearmor", "--output", "/usr/share/keyrings/cloudflare-warp-archive-keyring.gpg"},
		{"sh", "-c", "echo \"deb [arch=amd64 signed-by=/usr/share/keyrings/cloudflare-warp-archive-keyring.gpg] https://pkg.cloudflareclient.com/ $(lsb_release -cs) main\" | tee /etc/apt/sources.list.d/cloudflare-client.list"},
		{"apt", "update"},
		{"apt", "install", "-y", "cloudflare-warp"},
	}

	for _, cmd := range commands {
		if err := nm.runCommand("sh", "-c", strings.Join(cmd, " ")); err != nil {
			return fmt.Errorf("failed to execute command %v: %w", cmd, err)
		}
	}

	nm.logger.Info("WARP client installed successfully")
	return nil
}

// isWARPInstalled checks if WARP client is installed
func (nm *NetworkManagerImpl) isWARPInstalled() bool {
	cmd := exec.Command("which", "warp-cli")
	err := cmd.Run()
	return err == nil
}

// ConfigureWARPProxy configures WARP in SOCKS5 proxy mode
func (nm *NetworkManagerImpl) ConfigureWARPProxy(port int) error {
	nm.logger.Infof("Configuring WARP proxy mode on port %d", port)

	if !nm.isWARPInstalled() {
		return fmt.Errorf("WARP client is not installed")
	}

	// Register WARP client (if not already registered)
	if err := nm.runCommand("warp-cli", "registration", "new"); err != nil {
		nm.logger.Warnf("WARP registration failed (may already be registered): %v", err)
	}

	// Set proxy mode
	if err := nm.runCommand("warp-cli", "mode", "proxy"); err != nil {
		return fmt.Errorf("failed to set proxy mode: %w", err)
	}

	// Set proxy port
	if err := nm.runCommand("warp-cli", "proxy", "port", fmt.Sprintf("%d", port)); err != nil {
		return fmt.Errorf("failed to set proxy port: %w", err)
	}

	nm.logger.Infof("WARP proxy configured on port %d", port)
	return nil
}

// StartWARPService starts the WARP service
func (nm *NetworkManagerImpl) StartWARPService() error {
	nm.logger.Info("Starting WARP service")

	if !nm.isWARPInstalled() {
		return fmt.Errorf("WARP client is not installed")
	}

	// Connect to WARP
	if err := nm.runCommand("warp-cli", "connect"); err != nil {
		return fmt.Errorf("failed to connect to WARP: %w", err)
	}

	nm.logger.Info("WARP service started successfully")
	return nil
}

// StopWARPService stops the WARP service
func (nm *NetworkManagerImpl) StopWARPService() error {
	nm.logger.Info("Stopping WARP service")

	if !nm.isWARPInstalled() {
		return fmt.Errorf("WARP client is not installed")
	}

	// Disconnect from WARP
	if err := nm.runCommand("warp-cli", "disconnect"); err != nil {
		return fmt.Errorf("failed to disconnect from WARP: %w", err)
	}

	nm.logger.Info("WARP service stopped successfully")
	return nil
}

// RestartWARPService restarts the WARP service
func (nm *NetworkManagerImpl) RestartWARPService() error {
	nm.logger.Info("Restarting WARP service")

	if err := nm.StopWARPService(); err != nil {
		nm.logger.Warnf("Failed to stop WARP service: %v", err)
	}

	// Wait a moment before starting
	time.Sleep(2 * time.Second)

	return nm.StartWARPService()
}

// IsWARPConnected checks if WARP is connected
func (nm *NetworkManagerImpl) IsWARPConnected() (bool, error) {
	if !nm.isWARPInstalled() {
		return false, fmt.Errorf("WARP client is not installed")
	}

	output, err := nm.runCommandWithOutput("warp-cli", "status")
	if err != nil {
		return false, fmt.Errorf("failed to get WARP status: %w", err)
	}

	// Check if status contains "Connected"
	return strings.Contains(strings.ToLower(output), "connected"), nil
}

// GetWARPStatus returns detailed WARP status
func (nm *NetworkManagerImpl) GetWARPStatus() (map[string]interface{}, error) {
	status := map[string]interface{}{
		"installed":  false,
		"connected":  false,
		"mode":       "unknown",
		"proxy_port": 0,
	}

	if !nm.isWARPInstalled() {
		return status, nil
	}

	status["installed"] = true

	// Get detailed status
	output, err := nm.runCommandWithOutput("warp-cli", "status")
	if err != nil {
		return status, fmt.Errorf("failed to get WARP status: %w", err)
	}

	// Parse status output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Status") {
			if strings.Contains(line, "Connected") {
				status["connected"] = true
			}
		} else if strings.Contains(line, "Mode") {
			if strings.Contains(line, "Proxy") {
				status["mode"] = "proxy"
			} else if strings.Contains(line, "Warp") {
				status["mode"] = "warp"
			}
		}
	}

	// Get proxy port if in proxy mode
	if mode, ok := status["mode"].(string); ok && mode == "proxy" {
		proxyOutput, err := nm.runCommandWithOutput("warp-cli", "proxy", "port")
		if err == nil {
			if port := strings.TrimSpace(proxyOutput); port != "" {
				status["proxy_port"] = port
			}
		}
	}

	return status, nil
}

// RouteTrafficThroughWARP configures routing through WARP proxy
func (nm *NetworkManagerImpl) RouteTrafficThroughWARP(interfaceName string) error {
	nm.logger.Infof("Configuring traffic routing through WARP on interface: %s", interfaceName)

	// Enable IP forwarding
	if err := nm.runCommand("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Clear existing NAT rules for this interface
	nm.runCommand("iptables", "-t", "nat", "-F", "OUTPUT")
	nm.runCommand("iptables", "-t", "nat", "-F", "POSTROUTING")

	// Configure routing through WARP using iptables
	// Redirect HTTP/HTTPS traffic to WARP proxy
	warpPort := nm.config.Hysteria2.WARPProxyPort
	if warpPort == 0 {
		warpPort = 1080 // default
	}

	rules := []string{
		// Redirect HTTP/HTTPS traffic to SOCKS5 proxy
		fmt.Sprintf("-t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports %d", warpPort),
		fmt.Sprintf("-t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports %d", warpPort),
		// Masquerade traffic going out through the interface
		fmt.Sprintf("-t nat -A POSTROUTING -o %s -j MASQUERADE", interfaceName),
	}

	for _, rule := range rules {
		if err := nm.runCommand("iptables", strings.Fields(rule)...); err != nil {
			return fmt.Errorf("failed to apply routing rule %s: %w", rule, err)
		}
	}

	nm.logger.Info("Traffic routing configured through WARP successfully")
	return nil
}

// DisableWarpRouting disables WARP-specific routing rules
func (nm *NetworkManagerImpl) DisableWarpRouting() error {
	nm.logger.Info("Disabling WARP routing rules")

	// Clear NAT rules that redirect to WARP
	nm.runCommand("iptables", "-t", "nat", "-F", "OUTPUT")
	nm.runCommand("iptables", "-t", "nat", "-F", "POSTROUTING")

	nm.logger.Info("WARP routing rules disabled")
	return nil
}
