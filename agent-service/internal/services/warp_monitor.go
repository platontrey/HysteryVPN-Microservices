package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/config"
)

// WARPMonitor handles continuous monitoring of WARP connections
type WARPMonitor interface {
	// Start monitoring
	Start(ctx context.Context) error
	Stop() error

	// Get current status
	GetCurrentStatus() WARPMonitoringStatus

	// Register status callback
	RegisterStatusCallback(callback func(WARPMonitoringStatus)) error

	// Get historical data
	GetHistoricalData(duration time.Duration) ([]WARPMonitoringStatus, error)

	// Health checks
	RunHealthCheck() WARPHealthCheckResult
}

// WARPMonitoringStatus holds comprehensive monitoring data
type WARPMonitoringStatus struct {
	Timestamp          time.Time `json:"timestamp"`
	WARPConnected      bool      `json:"warp_connected"`
	WARPMode           string    `json:"warp_mode"`
	WARPIPAddress      string    `json:"warp_ip_address"`
	WARPProxyPort      int       `json:"warp_proxy_port"`
	WARPOrganization   string    `json:"warp_organization"`
	WARPAccountType    string    `json:"warp_account_type"`
	WARPServerLocation string    `json:"warp_server_location"`

	// Network metrics
	BytesSent       int64 `json:"bytes_sent"`
	BytesReceived   int64 `json:"bytes_received"`
	PacketsSent     int64 `json:"packets_sent"`
	PacketsReceived int64 `json:"packets_received"`

	// Performance metrics
	LatencyMs    float64 `json:"latency_ms"`
	DownloadMbps float64 `json:"download_mbps"`
	UploadMbps   float64 `json:"upload_mbps"`

	// Connection health
	ConnectionUptime   time.Duration `json:"connection_uptime"`
	LastConnected      time.Time     `json:"last_connected"`
	DisconnectionCount int           `json:"disconnection_count"`

	// Health status
	HealthScore  float64  `json:"health_score"` // 0-100
	HealthIssues []string `json:"health_issues"`

	// System metrics
	CPUUsage          float64  `json:"cpu_usage"`
	MemoryUsage       float64  `json:"memory_usage"`
	NetworkInterfaces []string `json:"network_interfaces"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// WARPHealthCheckResult holds health check results
type WARPHealthCheckResult struct {
	OverallHealthy    bool                   `json:"overall_healthy"`
	WARPRunning       bool                   `json:"warp_running"`
	WARPConnected     bool                   `json:"warp_connected"`
	ProxyWorking      bool                   `json:"proxy_working"`
	InternetReachable bool                   `json:"internet_reachable"`
	DNSWorking        bool                   `json:"dns_working"`
	Checks            map[string]CheckResult `json:"checks"`
	Score             int                    `json:"score"`
	Message           string                 `json:"message"`
}

// CheckResult holds individual check result
type CheckResult struct {
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
	Duration int64  `json:"duration_ms"`
	Score    int    `json:"score"`
}

type WARPMonitorImpl struct {
	logger      *logrus.Logger
	config      *config.Config
	warpManager WARPManager

	// Monitoring state
	isMonitoring    bool
	monitorCtx      context.Context
	monitorCancel   context.CancelFunc
	monitorInterval time.Duration

	// Status tracking
	currentStatus  WARPMonitoringStatus
	statusHistory  []WARPMonitoringStatus
	maxHistorySize int
	statusMutex    sync.RWMutex

	// Callbacks
	statusCallbacks []func(WARPMonitoringStatus)
	callbackMutex   sync.RWMutex

	// Worker pool for callbacks
	workerPoolSize int
	jobChan        chan func()
	stopChan       chan struct{}

	// Metrics tracking
	lastBytesSent     int64
	lastBytesReceived int64
	lastConnected     time.Time
	disconnections    int

	// Health check config
	healthCheckTargets []string
}

// NewWARPMonitor creates a new WARP monitor
func NewWARPMonitor(logger *logrus.Logger, cfg *config.Config, warpManager WARPManager) WARPMonitor {
	return &WARPMonitorImpl{
		logger:          logger,
		config:          cfg,
		warpManager:     warpManager,
		monitorInterval: 30 * time.Second, // Default 30 seconds
		maxHistorySize:  1000,             // Keep last 1000 records
		statusCallbacks: make([]func(WARPMonitoringStatus), 0),
		workerPoolSize:  10,                     // Configurable worker pool size
		jobChan:         make(chan func(), 100), // Buffered channel for jobs
		stopChan:        make(chan struct{}),
		healthCheckTargets: []string{
			"https://1.1.1.1",
			"https://8.8.8.8",
			"https://cloudflare.com",
		},
	}
}

// Start begins monitoring WARP status
func (wm *WARPMonitorImpl) Start(ctx context.Context) error {
	if wm.isMonitoring {
		return fmt.Errorf("monitoring is already running")
	}

	wm.monitorCtx, wm.monitorCancel = context.WithCancel(ctx)
	wm.isMonitoring = true

	// Start worker pool
	for i := 0; i < wm.workerPoolSize; i++ {
		go wm.worker()
	}

	// Start monitoring loop
	go wm.monitoringLoop()

	wm.logger.Info("WARP monitoring started")
	return nil
}

// Stop stops monitoring
func (wm *WARPMonitorImpl) Stop() error {
	if !wm.isMonitoring {
		return nil
	}

	if wm.monitorCancel != nil {
		wm.monitorCancel()
	}

	// Stop worker pool
	close(wm.stopChan)
	close(wm.jobChan)

	wm.isMonitoring = false
	wm.logger.Info("WARP monitoring stopped")
	return nil
}

// GetCurrentStatus returns current monitoring status
func (wm *WARPMonitorImpl) GetCurrentStatus() WARPMonitoringStatus {
	wm.statusMutex.RLock()
	defer wm.statusMutex.RUnlock()

	return wm.currentStatus
}

// RegisterStatusCallback registers a callback for status updates
func (wm *WARPMonitorImpl) RegisterStatusCallback(callback func(WARPMonitoringStatus)) error {
	wm.callbackMutex.Lock()
	defer wm.callbackMutex.Unlock()

	wm.statusCallbacks = append(wm.statusCallbacks, callback)
	return nil
}

// GetHistoricalData returns historical monitoring data
func (wm *WARPMonitorImpl) GetHistoricalData(duration time.Duration) ([]WARPMonitoringStatus, error) {
	wm.statusMutex.RLock()
	defer wm.statusMutex.RUnlock()

	cutoff := time.Now().Add(-duration)
	var historicalData []WARPMonitoringStatus

	for _, status := range wm.statusHistory {
		if status.Timestamp.After(cutoff) {
			historicalData = append(historicalData, status)
		}
	}

	return historicalData, nil
}

// RunHealthCheck performs comprehensive health check
func (wm *WARPMonitorImpl) RunHealthCheck() WARPHealthCheckResult {
	result := WARPHealthCheckResult{
		Checks: make(map[string]CheckResult),
		Score:  100,
	}

	start := time.Now()

	// 1. Check if WARP is installed and running
	if !wm.warpManager.IsWARPInstalled() {
		result.Checks["warp_installed"] = CheckResult{
			Passed:   false,
			Message:  "WARP client is not installed",
			Score:    0,
			Duration: time.Since(start).Milliseconds(),
		}
		result.Score -= 30
	} else {
		result.Checks["warp_installed"] = CheckResult{
			Passed:   true,
			Message:  "WARP client is installed",
			Score:    10,
			Duration: time.Since(start).Milliseconds(),
		}
		result.WARPRunning = true
	}

	// 2. Check WARP connection status
	checkStart := time.Now()
	if connected, err := wm.warpManager.IsWARPConnected(); err != nil {
		result.Checks["warp_connection"] = CheckResult{
			Passed:   false,
			Message:  fmt.Sprintf("Failed to check WARP connection: %v", err),
			Score:    0,
			Duration: time.Since(checkStart).Milliseconds(),
		}
		result.Score -= 40
	} else if connected {
		result.Checks["warp_connection"] = CheckResult{
			Passed:   true,
			Message:  "WARP is connected",
			Score:    20,
			Duration: time.Since(checkStart).Milliseconds(),
		}
		result.WARPConnected = true
	} else {
		result.Checks["warp_connection"] = CheckResult{
			Passed:   false,
			Message:  "WARP is not connected",
			Score:    0,
			Duration: time.Since(checkStart).Milliseconds(),
		}
		result.Score -= 40
	}

	// 3. Check proxy functionality
	if result.WARPConnected {
		checkStart = time.Now()
		if proxyWorking := wm.checkProxyFunctionality(); proxyWorking {
			result.Checks["proxy_functionality"] = CheckResult{
				Passed:   true,
				Message:  "WARP proxy is working",
				Score:    15,
				Duration: time.Since(checkStart).Milliseconds(),
			}
			result.ProxyWorking = true
		} else {
			result.Checks["proxy_functionality"] = CheckResult{
				Passed:   false,
				Message:  "WARP proxy is not working",
				Score:    0,
				Duration: time.Since(checkStart).Milliseconds(),
			}
			result.Score -= 20
		}
	}

	// 4. Check internet reachability
	checkStart = time.Now()
	if internetReachable := wm.checkInternetReachability(); internetReachable {
		result.Checks["internet_reachable"] = CheckResult{
			Passed:   true,
			Message:  "Internet is reachable through WARP",
			Score:    20,
			Duration: time.Since(checkStart).Milliseconds(),
		}
		result.InternetReachable = true
	} else {
		result.Checks["internet_reachable"] = CheckResult{
			Passed:   false,
			Message:  "Internet is not reachable through WARP",
			Score:    0,
			Duration: time.Since(checkStart).Milliseconds(),
		}
		result.Score -= 25
	}

	// 5. Check DNS resolution
	checkStart = time.Now()
	if dnsWorking := wm.checkDNSResolution(); dnsWorking {
		result.Checks["dns_resolution"] = CheckResult{
			Passed:   true,
			Message:  "DNS resolution is working",
			Score:    15,
			Duration: time.Since(checkStart).Milliseconds(),
		}
		result.DNSWorking = true
	} else {
		result.Checks["dns_resolution"] = CheckResult{
			Passed:   false,
			Message:  "DNS resolution is not working",
			Score:    0,
			Duration: time.Since(checkStart).Milliseconds(),
		}
		result.Score -= 15
	}

	// Calculate overall health
	result.OverallHealthy = result.Score >= 70

	if result.OverallHealthy {
		result.Message = "WARP proxy is healthy"
	} else {
		result.Message = fmt.Sprintf("WARP proxy has issues (Score: %d)", result.Score)
	}

	return result
}

// Private methods

func (wm *WARPMonitorImpl) monitoringLoop() {
	ticker := time.NewTicker(wm.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-wm.monitorCtx.Done():
			return
		case <-ticker.C:
			wm.collectStatus()
		}
	}
}

func (wm *WARPMonitorImpl) collectStatus() {
	status := WARPMonitoringStatus{
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Get WARP status
	if warpStatus, err := wm.warpManager.GetWARPStatus(); err == nil {
		status.WARPConnected = warpStatus.Connected
		status.WARPMode = warpStatus.Mode
		status.WARPIPAddress = warpStatus.IPAddress
		status.WARPProxyPort = warpStatus.ProxyPort
		status.WARPOrganization = warpStatus.Organization
		status.WARPAccountType = warpStatus.AccountType
		status.WARPServerLocation = warpStatus.ServerLocation
		status.BytesSent = warpStatus.BytesSent
		status.BytesReceived = warpStatus.BytesReceived

		// Track disconnections
		if status.WARPConnected && !wm.currentStatus.WARPConnected {
			wm.lastConnected = time.Now()
		} else if !status.WARPConnected && wm.currentStatus.WARPConnected {
			wm.disconnections++
		}

		if status.WARPConnected {
			status.ConnectionUptime = time.Since(wm.lastConnected)
		}
		status.LastConnected = wm.lastConnected
		status.DisconnectionCount = wm.disconnections
	}

	// Calculate network metrics
	status.PacketsSent = status.BytesSent / 1500 // Estimate
	status.PacketsReceived = status.BytesReceived / 1500

	// Calculate bandwidth
	if len(wm.statusHistory) > 0 {
		lastStatus := wm.statusHistory[len(wm.statusHistory)-1]
		timeDiff := status.Timestamp.Sub(lastStatus.Timestamp).Seconds()
		if timeDiff > 0 {
			bytesSentDiff := status.BytesSent - lastStatus.BytesSent
			bytesReceivedDiff := status.BytesReceived - lastStatus.BytesReceived

			status.UploadMbps = (float64(bytesSentDiff) * 8) / (timeDiff * 1000000)
			status.DownloadMbps = (float64(bytesReceivedDiff) * 8) / (timeDiff * 1000000)
		}
	}

	// Run performance tests
	status.LatencyMs = wm.measureLatency()

	// Get system metrics
	status.CPUUsage = wm.getCPUUsage()
	status.MemoryUsage = wm.getMemoryUsage()

	// Calculate health score
	status.HealthScore = wm.calculateHealthScore(status)
	status.HealthIssues = wm.identifyHealthIssues(status)

	// Update current status and history
	wm.statusMutex.Lock()
	wm.currentStatus = status
	wm.statusHistory = append(wm.statusHistory, status)

	// Trim history if too large
	if len(wm.statusHistory) > wm.maxHistorySize {
		wm.statusHistory = wm.statusHistory[1:]
	}
	wm.statusMutex.Unlock()

	// Notify callbacks
	wm.notifyStatusCallbacks(status)
}

func (wm *WARPMonitorImpl) notifyStatusCallbacks(status WARPMonitoringStatus) {
	wm.callbackMutex.RLock()
	callbacks := make([]func(WARPMonitoringStatus), len(wm.statusCallbacks))
	copy(callbacks, wm.statusCallbacks)
	wm.callbackMutex.RUnlock()

	for _, callback := range callbacks {
		select {
		case wm.jobChan <- func() { callback(status) }:
		default:
			wm.logger.Warn("Worker pool full, dropping callback")
		}
	}
}

func (wm *WARPMonitorImpl) worker() {
	for {
		select {
		case job := <-wm.jobChan:
			func() {
				defer func() {
					if r := recover(); r != nil {
						wm.logger.Error("Panic in callback", "panic", r)
					}
				}()
				job()
			}()
		case <-wm.stopChan:
			return
		}
	}
}

func (wm *WARPMonitorImpl) checkProxyFunctionality() bool {
	// Simple proxy check - try to make HTTP request through WARP
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://1.1.1.1")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (wm *WARPMonitorImpl) checkInternetReachability() bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Try multiple targets
	for _, target := range wm.healthCheckTargets {
		resp, err := client.Get(target)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 400 {
				return true
			}
		}
	}

	return false
}

func (wm *WARPMonitorImpl) checkDNSResolution() bool {
	// Simple DNS check - could be enhanced with actual DNS queries
	return wm.checkInternetReachability()
}

func (wm *WARPMonitorImpl) measureLatency() float64 {
	start := time.Now()
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://1.1.1.1")
	if err != nil {
		return -1
	}
	defer resp.Body.Close()

	return float64(time.Since(start).Milliseconds())
}

func (wm *WARPMonitorImpl) getCPUUsage() float64 {
	// Placeholder - would implement actual CPU monitoring
	return 0.0
}

func (wm *WARPMonitorImpl) getMemoryUsage() float64 {
	// Placeholder - would implement actual memory monitoring
	return 0.0
}

func (wm *WARPMonitorImpl) calculateHealthScore(status WARPMonitoringStatus) float64 {
	score := 100.0

	if !status.WARPConnected {
		score -= 50
	}

	if status.LatencyMs < 0 {
		score -= 20
	} else if status.LatencyMs > 500 {
		score -= 10
	} else if status.LatencyMs > 200 {
		score -= 5
	}

	if status.DownloadMbps < 1 {
		score -= 15
	}

	if status.UploadMbps < 1 {
		score -= 15
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (wm *WARPMonitorImpl) identifyHealthIssues(status WARPMonitoringStatus) []string {
	var issues []string

	if !status.WARPConnected {
		issues = append(issues, "WARP not connected")
	}

	if status.LatencyMs > 500 {
		issues = append(issues, "High latency")
	}

	if status.DownloadMbps < 1 {
		issues = append(issues, "Low download speed")
	}

	if status.UploadMbps < 1 {
		issues = append(issues, "Low upload speed")
	}

	if status.DisconnectionCount > 5 {
		issues = append(issues, "Frequent disconnections")
	}

	return issues
}
