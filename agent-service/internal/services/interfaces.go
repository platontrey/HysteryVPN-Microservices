package services

import (
	"context"
	"time"
)

// ConfigManager handles configuration management
type ConfigManager interface {
	GetConfig() (map[string]interface{}, error)
	UpdateConfig(config map[string]interface{}) error
}

// MetricsCollector collects system metrics
type MetricsCollector interface {
	Collect() (map[string]interface{}, error)
	StartCollection(ctx context.Context) error
	StopCollection() error
}

// SystemManager handles system operations
type SystemManager interface {
	GetSystemInfo() (map[string]interface{}, error)
	RestartService() error
	UpdateSoftware() error
	IsBBREnabled() (bool, error)
	EnableBBR() error
	CheckAndEnableBBR() error

	// WARP installation methods
	InstallWARPClient() error
	IsWARPClientInstalled() bool
	SetupWARPSystemdService() error
	ConfigureWARPAutoStart(autoStart bool) error
}

// NetworkManager handles network operations including masquerading
type NetworkManager interface {
	EnableMasquerading(interfaceName string) error
	DisableMasquerading(interfaceName string) error
	IsMasqueradingEnabled(interfaceName string) (bool, error)
	GetNetworkInterfaces() ([]string, error)

	// WARP-specific methods
	InstallWARPClient() error
	ConfigureWARPProxy(port int) error
	StartWARPService() error
	StopWARPService() error
	RestartWARPService() error
	IsWARPConnected() (bool, error)
	GetWARPStatus() (map[string]interface{}, error)
	RouteTrafficThroughWARP(interfaceName string) error
	DisableWarpRouting() error
}

// WARPManager handles Cloudflare WARP client operations
type WARPManager interface {
	// Installation and setup
	InstallWARPClient() error
	UninstallWARPClient() error
	IsWARPInstalled() bool
	SetupWARPSystemdService() error

	// Connection management
	ConnectWARP() error
	DisconnectWARP() error
	RestartWARP() error
	IsWARPConnected() (bool, error)

	// Proxy configuration
	EnableProxyMode(port int) error
	DisableProxyMode() error
	GetProxyPort() (int, error)
	SetProxyPort(port int) error

	// Configuration management
	ConfigureWARP(cfg WARPConfig) error
	GetWARPConfiguration() (WARPConfig, error)
	ValidateConfiguration(cfg WARPConfig) error

	// Status and monitoring
	GetWARPStatus() (WARPStatus, error)
	StartStatusMonitoring(ctx context.Context, interval time.Duration) (<-chan WARPStatus, error)
	StopStatusMonitoring() error

	// Traffic routing
	EnableTrafficRouting(interfaceName string) error
	DisableTrafficRouting() error
	GetRoutingRules() ([]string, error)

	// License and organization management
	SetLicenseKey(licenseKey string) error
	SetOrganization(organization string) error
	GetLicenseInfo() (WARPLicenseInfo, error)
}

// WARPConfig holds WARP configuration
type WARPConfig struct {
	Enabled         bool     `json:"enabled"`
	ProxyPort       int      `json:"proxy_port"`
	AutoConnect     bool     `json:"auto_connect"`
	NotifyOnFail    bool     `json:"notify_on_fail"`
	ClientType      string   `json:"client_type"` // "local", "docker"
	LicenseKey      string   `json:"license_key"`
	Organization    string   `json:"organization"`
	Mode            string   `json:"mode"` // "proxy", "warp"
	DnsEnabled      bool     `json:"dns_enabled"`
	DnsServers      []string `json:"dns_servers"`
	ExcludeLAN      bool     `json:"exclude_lan"`
	SplitTunnel     bool     `json:"split_tunnel"`
	SplitTunnelApps []string `json:"split_tunnel_apps"`
}

// WARPStatus holds WARP status information
type WARPStatus struct {
	Installed      bool          `json:"installed"`
	Connected      bool          `json:"connected"`
	Mode           string        `json:"mode"`
	ProxyPort      int           `json:"proxy_port"`
	AccountType    string        `json:"account_type"`
	Organization   string        `json:"organization"`
	IPAddress      string        `json:"ip_address"`
	Location       string        `json:"location"`
	ServerLocation string        `json:"server_location"`
	LastConnected  time.Time     `json:"last_connected"`
	Uptime         time.Duration `json:"uptime"`
	BytesSent      int64         `json:"bytes_sent"`
	BytesReceived  int64         `json:"bytes_received"`
	Health         string        `json:"health"` // "good", "warning", "error"
	Error          string        `json:"error,omitempty"`
}

// WARPLicenseInfo holds license information
type WARPLicenseInfo struct {
	HasLicense   bool      `json:"has_license"`
	LicenseType  string    `json:"license_type"`
	Organization string    `json:"organization"`
	Seats        int       `json:"seats"`
	UsedSeats    int       `json:"used_seats"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsValid      bool      `json:"is_valid"`
	Error        string    `json:"error,omitempty"`
}

// XrayManager handles Xray-core VPN server management for VLESS, Reality, and other protocols
type XrayManager interface {
	// Installation and lifecycle
	InstallXray() error
	IsXrayInstalled() bool
	StartXray(configPath string) error
	StopXray() error
	RestartXray(configPath string) error
	GetXrayStatus() (map[string]interface{}, error)

	// Configuration management
	GenerateConfig(protocol string, configTemplate string) (string, error)
	ValidateConfig(protocol string, config map[string]interface{}) error

	// Protocol-specific configuration
	ConfigureVLESS(uuid, dest string, flow string) error
	ConfigureReality(dest, serverNames string, privateKey string, shortIds []string) error
	GenerateRealityKeys() (privateKey, publicKey string, err error)

	// Certificate management for Reality
	GenerateRealityCert(domain string) error
	ValidateRealityCert(domain string) (bool, error)
}

// LocalServices aggregates all local services
type LocalServices struct {
	ConfigManager    ConfigManager
	MetricsCollector MetricsCollector
	SystemManager    SystemManager
	NetworkManager   NetworkManager
	HysteriaManager  HysteriaManager
	XrayManager      XrayManager
	WARPManager      WARPManager
}
