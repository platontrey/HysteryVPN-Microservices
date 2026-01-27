package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	MasterServer string          `mapstructure:"master_server"`
	Node         NodeConfig      `mapstructure:"node"`
	Metrics      MetricsConfig   `mapstructure:"metrics"`
	Logging      LoggingConfig   `mapstructure:"logging"`
	Network      NetworkConfig   `mapstructure:"network"`
	Hysteria2    Hysteria2Config `mapstructure:"hysteria2"`
	Xray         XrayConfig      `mapstructure:"xray"`
}

type NodeConfig struct {
	ID           string            `mapstructure:"id"`
	Name         string            `mapstructure:"name"`
	Hostname     string            `mapstructure:"hostname"`
	IPAddress    string            `mapstructure:"ip_address"`
	Location     string            `mapstructure:"location"`
	Country      string            `mapstructure:"country"`
	GRPCPort     int               `mapstructure:"grpc_port"`
	Capabilities map[string]string `mapstructure:"capabilities"`
	Metadata     map[string]string `mapstructure:"metadata"`
}

type MetricsConfig struct {
	CollectInterval int `mapstructure:"collect_interval"` // seconds
	ReportInterval  int `mapstructure:"report_interval"`  // seconds
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json, text
}

type NetworkConfig struct {
	EnableMasquerading bool   `mapstructure:"enable_masquerading"`
	DefaultInterface   string `mapstructure:"default_interface"`
}

type Hysteria2Config struct {
	EnableBBR          bool   `mapstructure:"enable_bbr"`
	EnableSystemd      bool   `mapstructure:"enable_systemd"`
	PortHopping        bool   `mapstructure:"port_hopping"`
	HopInterval        int    `mapstructure:"hop_interval"` // seconds
	HopStartPort       int    `mapstructure:"hop_start_port"`
	HopEndPort         int    `mapstructure:"hop_end_port"`
	SalamanderEnabled  bool   `mapstructure:"salamander_enabled"`
	SalamanderPassword string `mapstructure:"salamander_password"`
	ObfsType           string `mapstructure:"obfs_type"`     // "salamander" or others
	CustomConfig       string `mapstructure:"custom_config"` // JSON string for custom settings
	ListenPorts        []int  `mapstructure:"listen_ports"`
	DefaultListenPort  int    `mapstructure:"default_listen_port"`
	AuthType           string `mapstructure:"auth_type"` // "password", "userpass", etc.
	AuthPassword       string `mapstructure:"auth_password"`
	UpMbps             int    `mapstructure:"up_mbps"`
	DownMbps           int    `mapstructure:"down_mbps"`

	// Advanced Obfuscation Settings for Russian DPI Bypass
	AdvancedObfuscationEnabled bool     `mapstructure:"advanced_obfuscation_enabled"`
	QUICObfuscationEnabled     bool     `mapstructure:"quic_obfuscation_enabled"`
	QUICScrambleTransform      bool     `mapstructure:"quic_scramble_transform"`
	QUICPacketPadding          int      `mapstructure:"quic_packet_padding"` // 1200-1500 bytes
	QUICTimingRandomization    bool     `mapstructure:"quic_timing_randomization"`
	TLSFingerprintRotation     bool     `mapstructure:"tls_fingerprint_rotation"`
	TLSFingerprints            []string `mapstructure:"tls_fingerprints"` // ["chrome", "firefox", "safari"]
	VLESSRealityEnabled        bool     `mapstructure:"vless_reality_enabled"`
	VLESSRealityTargets        []string `mapstructure:"vless_reality_targets"` // ["apple.com", "google.com"]
	MultiHopEnabled            bool     `mapstructure:"multi_hop_enabled"`
	TrafficShapingEnabled      bool     `mapstructure:"traffic_shaping_enabled"`
	BehavioralRandomization    bool     `mapstructure:"behavioral_randomization"`

	// SNI Configuration
	SNIEnabled            bool     `mapstructure:"sni_enabled"`
	SNIDomains            []string `mapstructure:"sni_domains"`
	DefaultSNI            string   `mapstructure:"default_sni"`
	SNICertPath           string   `mapstructure:"sni_cert_path"`
	SNIKeyPath            string   `mapstructure:"sni_key_path"`
	SNIAutoMode           bool     `mapstructure:"sni_auto_mode"`           // Auto-generate certs for domains
	SNIAutoRenew          bool     `mapstructure:"sni_auto_renew"`          // Auto-renew certificates
	SNIEmail              string   `mapstructure:"sni_email"`               // Email for Let's Encrypt
	SNILetsEncrypt        bool     `mapstructure:"sni_lets_encrypt"`        // Use Let's Encrypt for certificates
	SNIPreferredChallenge string   `mapstructure:"sni_preferred_challenge"` // Preferred ACME challenge
	SNIValidateDNS        bool     `mapstructure:"sni_validate_dns"`        // Validate DNS before certificate generation

	// WARP Configuration
	WARPEnabled      bool   `mapstructure:"warp_enabled"`
	WARPProxyPort    int    `mapstructure:"warp_proxy_port"`
	WARPAutoConnect  bool   `mapstructure:"warp_auto_connect"`
	WARPNotifyOnFail bool   `mapstructure:"warp_notify_on_fail"`
	WARPClientType   string `mapstructure:"warp_client_type"` // "local", "docker"
	WARPLicenseKey   string `mapstructure:"warp_license_key"`
	WARPOrganization string `mapstructure:"warp_organization"`
}

type XrayConfig struct {
	EnableAPI          bool     `mapstructure:"enable_api"`
	ListenPort         int      `mapstructure:"listen_port"`
	LogLevel           string   `mapstructure:"log_level"`
	SupportedProtocols []string `mapstructure:"supported_protocols"` // ["vless", "reality", "vmess", etc.]
	DefaultProtocol    string   `mapstructure:"default_protocol"`    // "vless"
	EnableStatistics   bool     `mapstructure:"enable_statistics"`

	// VLESS specific
	VLESSUUID string `mapstructure:"vless_uuid"`
	VLESSFlow string `mapstructure:"vless_flow"`

	// Reality specific
	RealityDest        string   `mapstructure:"reality_dest"`
	RealityServerNames []string `mapstructure:"reality_server_names"`
	RealityPrivateKey  string   `mapstructure:"reality_private_key"`
	RealityPublicKey   string   `mapstructure:"reality_public_key"`
	RealityShortIds    []string `mapstructure:"reality_short_ids"`

	// Certificate management
	CertPath string `mapstructure:"cert_path"`
	KeyPath  string `mapstructure:"key_path"`
}

func LoadConfig() (*Config, error) {
	setDefaults()

	viper.AutomaticEnv()
	bindEnvVars()

	// Load config file if exists
	viper.SetConfigName("agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, using environment variables and defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("master_server", "")
	viper.SetDefault("node.grpc_port", 50051)
	viper.SetDefault("metrics.collect_interval", 30)
	viper.SetDefault("metrics.report_interval", 60)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("network.enable_masquerading", false)
	viper.SetDefault("network.default_interface", "eth0")
	viper.SetDefault("hysteria2.enable_bbr", true)
	viper.SetDefault("hysteria2.enable_systemd", true)
	viper.SetDefault("hysteria2.port_hopping", false)
	viper.SetDefault("hysteria2.hop_interval", 30)
	viper.SetDefault("hysteria2.hop_start_port", 10000)
	viper.SetDefault("hysteria2.hop_end_port", 20000)
	viper.SetDefault("hysteria2.salamander_enabled", false)
	viper.SetDefault("hysteria2.obfs_type", "")
	viper.SetDefault("hysteria2.default_listen_port", 8080)
	viper.SetDefault("hysteria2.auth_type", "password")
	viper.SetDefault("hysteria2.up_mbps", 100)
	viper.SetDefault("hysteria2.down_mbps", 100)
	viper.SetDefault("hysteria2.sni_enabled", false)
	viper.SetDefault("hysteria2.sni_domains", []string{})
	viper.SetDefault("hysteria2.default_sni", "")
	viper.SetDefault("hysteria2.sni_cert_path", "/etc/hysteria/sni")
	viper.SetDefault("hysteria2.sni_key_path", "/etc/hysteria/sni")
	viper.SetDefault("hysteria2.sni_auto_mode", true)
	viper.SetDefault("hysteria2.sni_auto_renew", true)
	viper.SetDefault("hysteria2.sni_email", "")
	viper.SetDefault("hysteria2.sni_lets_encrypt", false)
	viper.SetDefault("hysteria2.sni_preferred_challenge", "http-01")
	viper.SetDefault("hysteria2.sni_validate_dns", true)

	// WARP defaults
	viper.SetDefault("hysteria2.warp_enabled", false)
	viper.SetDefault("hysteria2.warp_proxy_port", 1080)
	viper.SetDefault("hysteria2.warp_auto_connect", true)
	viper.SetDefault("hysteria2.warp_notify_on_fail", true)
	viper.SetDefault("hysteria2.warp_client_type", "local")
	viper.SetDefault("hysteria2.warp_license_key", "")
	viper.SetDefault("hysteria2.warp_organization", "")

	// Advanced obfuscation defaults for Russian DPI bypass
	viper.SetDefault("hysteria2.advanced_obfuscation_enabled", false)
	viper.SetDefault("hysteria2.quic_obfuscation_enabled", false)
	viper.SetDefault("hysteria2.quic_scramble_transform", false)
	viper.SetDefault("hysteria2.quic_packet_padding", 1300)
	viper.SetDefault("hysteria2.quic_timing_randomization", false)
	viper.SetDefault("hysteria2.tls_fingerprint_rotation", false)
	viper.SetDefault("hysteria2.tls_fingerprints", []string{"chrome"})
	viper.SetDefault("hysteria2.vless_reality_enabled", false)
	viper.SetDefault("hysteria2.vless_reality_targets", []string{"apple.com"})
	viper.SetDefault("hysteria2.multi_hop_enabled", false)
	viper.SetDefault("hysteria2.traffic_shaping_enabled", false)
	viper.SetDefault("hysteria2.behavioral_randomization", false)

	// Xray defaults
	viper.SetDefault("xray.enable_api", false)
	viper.SetDefault("xray.listen_port", 443)
	viper.SetDefault("xray.log_level", "warning")
	viper.SetDefault("xray.supported_protocols", []string{"vless", "reality"})
	viper.SetDefault("xray.default_protocol", "vless")
	viper.SetDefault("xray.enable_statistics", false)
	viper.SetDefault("xray.cert_path", "/etc/xray/cert.pem")
	viper.SetDefault("xray.key_path", "/etc/xray/key.pem")
}

func bindEnvVars() {
	viper.BindEnv("master_server", "MASTER_SERVER")
	viper.BindEnv("node.id", "NODE_ID")
	viper.BindEnv("node.name", "NODE_NAME")
	viper.BindEnv("node.hostname", "NODE_HOSTNAME")
	viper.BindEnv("node.ip_address", "NODE_IP_ADDRESS")
	viper.BindEnv("node.location", "NODE_LOCATION")
	viper.BindEnv("node.country", "NODE_COUNTRY")
	viper.BindEnv("node.grpc_port", "NODE_GRPC_PORT")
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")

	// WARP environment variables
	viper.BindEnv("hysteria2.warp_enabled", "WARP_ENABLED")
	viper.BindEnv("hysteria2.warp_proxy_port", "WARP_PROXY_PORT")
	viper.BindEnv("hysteria2.warp_auto_connect", "WARP_AUTO_CONNECT")
	viper.BindEnv("hysteria2.warp_notify_on_fail", "WARP_NOTIFY_ON_FAIL")
	viper.BindEnv("hysteria2.warp_client_type", "WARP_CLIENT_TYPE")
	viper.BindEnv("hysteria2.warp_license_key", "WARP_LICENSE_KEY")
	viper.BindEnv("hysteria2.warp_organization", "WARP_ORGANIZATION")
}

func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
