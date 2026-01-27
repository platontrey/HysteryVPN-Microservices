package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/config"
)

// TrafficRouter handles advanced traffic routing through WARP
type TrafficRouter interface {
	// Configure VPN -> WARP -> Internet routing
	ConfigureVPNTroughWARP(warpPort int, vpnInterface string) error

	// Create ACL rules for Hysteria2
	CreateHysteriaACL(warpEnabled bool) error

	// Setup iptables rules for traffic flow
	SetupIPTablesRules(warpPort int, vpnInterface string) error

	// Cleanup routing rules
	CleanupRoutingRules() error

	// Get current routing status
	GetRoutingStatus() (map[string]interface{}, error)
}

type TrafficRouterImpl struct {
	logger *logrus.Logger
	config *config.Config
}

// NewTrafficRouter creates a new TrafficRouter
func NewTrafficRouter(logger *logrus.Logger, cfg *config.Config) TrafficRouter {
	return &TrafficRouterImpl{
		logger: logger,
		config: cfg,
	}
}

// ConfigureVPNTroughWARP sets up complete VPN -> WARP -> Internet routing
func (tr *TrafficRouterImpl) ConfigureVPNTroughWARP(warpPort int, vpnInterface string) error {
	tr.logger.Infof("Configuring VPN -> WARP -> Internet routing (WARP port: %d, VPN interface: %s)", warpPort, vpnInterface)

	// 1. Enable IP forwarding
	if err := tr.enableIPForwarding(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// 2. Setup Hysteria2 ACL rules
	if err := tr.CreateHysteriaACL(true); err != nil {
		return fmt.Errorf("failed to create Hysteria2 ACL: %w", err)
	}

	// 3. Setup iptables rules for traffic routing
	if err := tr.SetupIPTablesRules(warpPort, vpnInterface); err != nil {
		return fmt.Errorf("failed to setup iptables rules: %w", err)
	}

	tr.logger.Info("VPN -> WARP -> Internet routing configured successfully")
	return nil
}

// CreateHysteriaACL creates ACL configuration for Hysteria2
func (tr *TrafficRouterImpl) CreateHysteriaACL(warpEnabled bool) error {
	tr.logger.Infof("Creating Hysteria2 ACL configuration (WARP enabled: %v)", warpEnabled)

	aclContent := `# Hysteria2 ACL Configuration for WARP Integration
# This ACL controls how traffic is routed through the VPN

# Local network bypass - don't route LAN traffic through VPN
- src, 127.0.0.1/8, dst, 127.0.0.1/8
- src, 10.0.0.0/8, dst, 10.0.0.0/8
- src, 172.16.0.0/12, dst, 172.16.0.0/12
- src, 192.168.0.0/16, dst, 192.168.0.0/16

# Private network ranges
- src, 100.64.0.0/10, dst, 100.64.0.0/10

# When WARP is enabled, route all other traffic through WARP
`

	if warpEnabled {
		aclContent += `# Route all internet traffic through WARP
# This is handled by the outbound proxy configuration
# All non-local traffic will be sent to WARP SOCKS5 proxy
`
	} else {
		aclContent += `# Allow all other traffic (traditional routing)
+
`
	}

	// Ensure ACL directory exists
	aclDir := "/etc/hysteria"
	if err := os.MkdirAll(aclDir, 0755); err != nil {
		return fmt.Errorf("failed to create ACL directory: %w", err)
	}

	// Write ACL file
	aclPath := "/etc/hysteria/acl.yaml"
	if err := os.WriteFile(aclPath, []byte(aclContent), 0644); err != nil {
		return fmt.Errorf("failed to write ACL file: %w", err)
	}

	tr.logger.Info("Hysteria2 ACL configuration created successfully")
	return nil
}

// SetupIPTablesRules creates iptables rules for traffic routing
func (tr *TrafficRouterImpl) SetupIPTablesRules(warpPort int, vpnInterface string) error {
	tr.logger.Infof("Setting up iptables rules (WARP port: %d, VPN interface: %s)", warpPort, vpnInterface)

	// Clear existing rules first
	if err := tr.cleanupExistingRules(); err != nil {
		tr.logger.Warnf("Failed to cleanup existing rules: %v", err)
	}

	rules := []string{
		// NAT table rules
		"-t nat -N HYSTERIA2-WARP",

		// Skip local traffic
		"-t nat -A HYSTERIA2-WARP -d 127.0.0.0/8 -j RETURN",
		"-t nat -A HYSTERIA2-WARP -d 10.0.0.0/8 -j RETURN",
		"-t nat -A HYSTERIA2-WARP -d 172.16.0.0/12 -j RETURN",
		"-t nat -A HYSTERIA2-WARP -d 192.168.0.0/16 -j RETURN",
		"-t nat -A HYSTERIA2-WARP -d 100.64.0.0/10 -j RETURN",

		// Redirect HTTP/HTTPS to WARP proxy
		fmt.Sprintf("-t nat -A HYSTERIA2-WARP -p tcp --dport 80 -j REDIRECT --to-ports %d", warpPort),
		fmt.Sprintf("-t nat -A HYSTERIA2-WARP -p tcp --dport 443 -j REDIRECT --to-ports %d", warpPort),

		// Redirect DNS to prevent leaks
		"-t nat -A HYSTERIA2-WARP -p udp --dport 53 -j REDIRECT --to-ports 53",
		"-t nat -A HYSTERIA2-WARP -p tcp --dport 53 -j REDIRECT --to-ports 53",

		// Apply the chain to OUTPUT
		"-t nat -A OUTPUT -j HYSTERIA2-WARP",

		// Filter table rules
		"-t filter -N HYSTERIA2-FORWARD",

		// Allow forwarding for VPN interface
		fmt.Sprintf("-t filter -A FORWARD -i %s -j ACCEPT", vpnInterface),
		fmt.Sprintf("-t filter -A FORWARD -o %s -j ACCEPT", vpnInterface),

		// Apply forward rules
		"-t filter -A FORWARD -j HYSTERIA2-FORWARD",

		// Mangle table for QoS if needed
		"-t mangle -N HYSTERIA2-QOS",
		"-t mangle -A OUTPUT -j HYSTERIA2-QOS",
	}

	// Apply all rules
	for _, rule := range rules {
		if err := tr.runCommand("iptables", strings.Fields(rule)...); err != nil {
			return fmt.Errorf("failed to apply rule '%s': %w", rule, err)
		}
	}

	tr.logger.Info("iptables rules configured successfully")
	return nil
}

// CleanupRoutingRules removes all routing rules
func (tr *TrafficRouterImpl) CleanupRoutingRules() error {
	tr.logger.Info("Cleaning up routing rules")

	return tr.cleanupExistingRules()
}

// GetRoutingStatus returns current routing configuration status
func (tr *TrafficRouterImpl) GetRoutingStatus() (map[string]interface{}, error) {
	status := map[string]interface{}{
		"ip_forwarding": false,
		"nat_rules":     0,
		"filter_rules":  0,
		"warp_rules":    false,
	}

	// Check IP forwarding
	if output, err := tr.runCommandWithOutput("sysctl", "net.ipv4.ip_forward"); err == nil {
		if strings.Contains(output, "1") {
			status["ip_forwarding"] = true
		}
	}

	// Count NAT rules
	if natOutput, err := tr.runCommandWithOutput("iptables", "-t", "nat", "-L", "HYSTERIA2-WARP", "--line-numbers"); err == nil {
		lines := strings.Split(natOutput, "\n")
		status["nat_rules"] = len(lines) - 2 // Subtract header lines
	}

	// Count filter rules
	if filterOutput, err := tr.runCommandWithOutput("iptables", "-t", "filter", "-L", "HYSTERIA2-FORWARD", "--line-numbers"); err == nil {
		lines := strings.Split(filterOutput, "\n")
		status["filter_rules"] = len(lines) - 2
	}

	// Check if WARP-specific rules exist
	if warpOutput, err := tr.runCommandWithOutput("iptables", "-t", "nat", "-L", "HYSTERIA2-WARP"); err == nil {
		status["warp_rules"] = strings.Contains(warpOutput, "REDIRECT")
	}

	return status, nil
}

// Helper methods

func (tr *TrafficRouterImpl) enableIPForwarding() error {
	return tr.runCommand("sysctl", "-w", "net.ipv4.ip_forward=1")
}

func (tr *TrafficRouterImpl) cleanupExistingRules() error {
	// Flush and delete custom chains
	chains := []string{"HYSTERIA2-WARP", "HYSTERIA2-FORWARD", "HYSTERIA2-QOS"}
	tables := []string{"nat", "filter", "mangle"}

	for _, table := range tables {
		// Flush each chain
		for _, chain := range chains {
			tr.runCommand("iptables", "-t", table, "-F", chain)
		}
	}

	for _, chain := range chains {
		// Delete from OUTPUT/FORWARD
		tr.runCommand("iptables", "-D", "OUTPUT", "-j", chain)
		tr.runCommand("iptables", "-D", "FORWARD", "-j", chain)

		// Delete chains
		tr.runCommand("iptables", "-t", "nat", "-X", chain)
		tr.runCommand("iptables", "-t", "filter", "-X", chain)
		tr.runCommand("iptables", "-t", "mangle", "-X", chain)
	}

	return nil
}

func (tr *TrafficRouterImpl) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	tr.logger.Debugf("Running command: %s %v", name, args)
	return cmd.Run()
}

func (tr *TrafficRouterImpl) runCommandWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	tr.logger.Debugf("Running command with output: %s %v", name, args)
	output, err := cmd.Output()
	return string(output), err
}
