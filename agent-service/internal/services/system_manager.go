package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// SystemManagerImpl implements SystemManager interface
type SystemManagerImpl struct {
	logger *logrus.Logger
}

// NewSystemManager creates a new SystemManager
func NewSystemManager(logger *logrus.Logger) SystemManager {
	return &SystemManagerImpl{
		logger: logger,
	}
}

// GetSystemInfo returns system information
func (sm *SystemManagerImpl) GetSystemInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"cpu_cores":  runtime.NumCPU(),
		"goroutines": runtime.NumGoroutine(),
		"go_version": runtime.Version(),
	}

	return info, nil
}

// RestartService restarts the service (placeholder)
func (sm *SystemManagerImpl) RestartService() error {
	sm.logger.Info("Service restart requested (not implemented)")
	return nil
}

// UpdateSoftware updates software (placeholder)
func (sm *SystemManagerImpl) UpdateSoftware() error {
	sm.logger.Info("Software update requested (not implemented)")
	return nil
}

// IsBBREnabled checks if BBR congestion control is enabled
func (sm *SystemManagerImpl) IsBBREnabled() (bool, error) {
	file, err := os.Open("/proc/sys/net/ipv4/tcp_congestion_control")
	if err != nil {
		sm.logger.Errorf("Failed to read congestion control: %v", err)
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		return strings.Contains(line, "bbr"), nil
	}

	return false, nil
}

// EnableBBR enables BBR congestion control
func (sm *SystemManagerImpl) EnableBBR() error {
	sm.logger.Info("Enabling BBR congestion control")

	// Check if bbr module is available
	if _, err := os.Stat("/lib/modules/$(uname -r)/kernel/net/ipv4/tcp_bbr.ko"); os.IsNotExist(err) {
		// Try to load bbr module
		cmd := exec.Command("modprobe", "tcp_bbr")
		if err := cmd.Run(); err != nil {
			sm.logger.Warnf("Failed to load tcp_bbr module: %v", err)
		}
	}

	// Set bbr as default congestion control
	cmd := exec.Command("sysctl", "-w", "net.ipv4.tcp_congestion_control=bbr")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set BBR: %w", err)
	}

	// Make it persistent
	cmd = exec.Command("echo", "net.ipv4.tcp_congestion_control=bbr", ">>", "/etc/sysctl.conf")
	if err := cmd.Run(); err != nil {
		sm.logger.Warnf("Failed to make BBR persistent: %v", err)
	}

	sm.logger.Info("BBR enabled successfully")
	return nil
}

// CheckAndEnableBBR checks if BBR is enabled and enables it if not
func (sm *SystemManagerImpl) CheckAndEnableBBR() error {
	enabled, err := sm.IsBBREnabled()
	if err != nil {
		return err
	}

	if !enabled {
		sm.logger.Info("BBR not enabled, enabling...")
		return sm.EnableBBR()
	}

	sm.logger.Info("BBR already enabled")
	return nil
}
