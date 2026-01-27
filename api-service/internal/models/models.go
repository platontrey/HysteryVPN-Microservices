package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username   string     `json:"username" gorm:"uniqueIndex;not null"`
	Email      string     `json:"email" gorm:"uniqueIndex;not null"`
	Password   string     `json:"-" gorm:"not null"` // Never return password in JSON
	FullName   *string    `json:"full_name"`
	Status     string     `json:"status" gorm:"default:'active';check:status IN ('active','suspended','deleted')"`
	Role       string     `json:"role" gorm:"default:'user';check:role IN ('admin','user')"`
	DataLimit  int64      `json:"data_limit" gorm:"default:0"`
	DataUsed   int64      `json:"data_used" gorm:"default:0"`
	ExpiryDate *time.Time `json:"expiry_date"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastLogin  *time.Time `json:"last_login"`
	Notes      *string    `json:"notes"`

	// Relations
	Devices []Device `json:"devices,omitempty" gorm:"foreignKey:UserID"`
}

type Device struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID  `json:"user_id" gorm:"not null"`
	Name      string     `json:"name" gorm:"not null"`
	DeviceID  string     `json:"device_id" gorm:"uniqueIndex;not null"`
	PublicKey string     `json:"public_key" gorm:"not null"`
	IPAddress *string    `json:"ip_address"`
	Status    string     `json:"status" gorm:"default:'active';check:status IN ('active','inactive','blocked')"`
	DataUsed  int64      `json:"data_used" gorm:"default:0"`
	CreatedAt time.Time  `json:"created_at"`
	LastSeen  *time.Time `json:"last_seen"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type Session struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID  `json:"user_id" gorm:"not null"`
	DeviceID     *uuid.UUID `json:"device_id"`
	SessionToken string     `json:"-" gorm:"uniqueIndex;not null"`
	RefreshToken *string    `json:"-" gorm:"uniqueIndex"`
	IPAddress    *string    `json:"ip_address"`
	UserAgent    *string    `json:"user_agent"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null"`
	CreatedAt    time.Time  `json:"created_at"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`

	// Relations
	User   User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Device *Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

type TrafficStats struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID  `json:"user_id" gorm:"not null"`
	DeviceID   *uuid.UUID `json:"device_id"`
	Upload     int64      `json:"upload" gorm:"default:0"`
	Download   int64      `json:"download" gorm:"default:0"`
	Total      int64      `json:"total" gorm:"default:0"`
	RecordedAt time.Time  `json:"recorded_at" gorm:"default:now()"`
	CreatedAt  time.Time  `json:"created_at"`

	// Relations
	User   User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Device *Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

type HysteriaConfig struct {
	ID         uuid.UUID              `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID              `json:"user_id" gorm:"not null"`
	DeviceID   *uuid.UUID             `json:"device_id"`
	ConfigName string                 `json:"config_name" gorm:"not null"`
	Protocol   string                 `json:"protocol" gorm:"default:'hysteria2';check:protocol IN ('hysteria2')"`
	ConfigData map[string]interface{} `json:"config_data" gorm:"type:jsonb;not null"`
	IsActive   bool                   `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`

	// Relations
	User   User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Device *Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

type XrayConfig struct {
	ID         uuid.UUID              `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID              `json:"user_id" gorm:"not null"`
	DeviceID   *uuid.UUID             `json:"device_id"`
	ConfigName string                 `json:"config_name" gorm:"not null"`
	Protocol   string                 `json:"protocol" gorm:"not null;check:protocol IN ('vless','vless-reality','vmess','trojan','shadowsocks')"`
	ConfigData map[string]interface{} `json:"config_data" gorm:"type:jsonb;not null"`
	IsActive   bool                   `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`

	// Relations
	User   User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Device *Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

type TrafficSummary struct {
	TotalUsers        int64               `json:"total_users"`
	ActiveUsers       int64               `json:"active_users"`
	TotalUpload       int64               `json:"total_upload"`
	TotalDownload     int64               `json:"total_download"`
	TotalDataTransfer int64               `json:"total_data_transfer"`
	TopUsers          []UserTrafficRank   `json:"top_users"`
	TopDevices        []DeviceTrafficRank `json:"top_devices"`
	From              time.Time           `json:"from"`
	To                time.Time           `json:"to"`
}

type UserTrafficRank struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Upload      int64     `json:"upload"`
	Download    int64     `json:"download"`
	Total       int64     `json:"total"`
	DeviceCount int       `json:"device_count"`
}

type DeviceTrafficRank struct {
	DeviceID   uuid.UUID `json:"device_id"`
	UserID     uuid.UUID `json:"user_id"`
	DeviceName string    `json:"device_name"`
	Username   string    `json:"username"`
	Upload     int64     `json:"upload"`
	Download   int64     `json:"download"`
	Total      int64     `json:"total"`
}

type Connection struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	DeviceID    string    `json:"device_id"`
	Address     string    `json:"address"`
	Upload      int64     `json:"upload"`
	Download    int64     `json:"download"`
	Duration    int64     `json:"duration"`
	ConnectedAt time.Time `json:"connected_at"`
}

type VPSNode struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name          string                 `json:"name" gorm:"size:100;not null"`
	Hostname      string                 `json:"hostname" gorm:"size:255;not null"`
	IPAddress     string                 `json:"ip_address" gorm:"size:45;not null;index"`
	Location      string                 `json:"location" gorm:"size:100"`
	Country       string                 `json:"country" gorm:"size:2"`
	GRPCPort      int                    `json:"grpc_port" gorm:"default:50051"`
	Status        string                 `json:"status" gorm:"size:20;default:'offline';index"`
	Version       string                 `json:"version" gorm:"size:50"`
	Capabilities  map[string]interface{} `json:"capabilities" gorm:"type:jsonb"`
	CreatedAt     time.Time              `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	LastHeartbeat *time.Time             `json:"last_heartbeat" gorm:"index"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"type:jsonb"`

	// Relations
	Assignments []NodeAssignment `json:"assignments,omitempty" gorm:"foreignKey:NodeID"`
	Metrics     []NodeMetric     `json:"metrics,omitempty" gorm:"foreignKey:NodeID"`
	Deployments []Deployment     `json:"deployments,omitempty" gorm:"foreignKey:NodeID"`
}

type NodeAssignment struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `json:"user_id" gorm:"not null;index"`
	NodeID     uuid.UUID `json:"node_id" gorm:"not null;index"`
	AssignedAt time.Time `json:"assigned_at" gorm:"default:CURRENT_TIMESTAMP"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`

	// Relations
	User *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Node *VPSNode `json:"node,omitempty" gorm:"foreignKey:NodeID"`
}

type NodeMetric struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	NodeID            uuid.UUID `json:"node_id" gorm:"not null;index"`
	CPUUsage          float64   `json:"cpu_usage" gorm:"type:decimal(5,2)"`
	MemoryUsage       float64   `json:"memory_usage" gorm:"type:decimal(5,2)"`
	BandwidthUp       int64     `json:"bandwidth_up"`
	BandwidthDown     int64     `json:"bandwidth_down"`
	ActiveConnections int       `json:"active_connections"`
	RecordedAt        time.Time `json:"recorded_at" gorm:"default:CURRENT_TIMESTAMP;index"`

	// Relations
	Node *VPSNode `json:"node,omitempty" gorm:"foreignKey:NodeID"`
}

type Deployment struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	NodeID        uuid.UUID  `json:"node_id" gorm:"not null;index"`
	ConfigVersion string     `json:"config_version" gorm:"size:50;not null"`
	Status        string     `json:"status" gorm:"size:20;default:'pending'"`
	DeployedAt    *time.Time `json:"deployed_at"`
	RollbackAt    *time.Time `json:"rollback_at"`
	ErrorMessage  string     `json:"error_message" gorm:"type:text"`

	// Relations
	Node *VPSNode `json:"node,omitempty" gorm:"foreignKey:NodeID"`
}

type ServiceStatus struct {
	IsRunning         bool          `json:"is_running"`
	Version           string        `json:"version"`
	Uptime            time.Duration `json:"uptime"`
	ActiveConnections int64         `json:"active_connections"`
	TotalTraffic      int64         `json:"total_traffic"`
	MemoryUsage       int64         `json:"memory_usage"`
	CPUUsage          float64       `json:"cpu_usage"`
}

// TableName overrides
func (User) TableName() string {
	return "users"
}

func (Device) TableName() string {
	return "devices"
}

func (Session) TableName() string {
	return "sessions"
}

func (TrafficStats) TableName() string {
	return "traffic_stats"
}

func (HysteriaConfig) TableName() string {
	return "hysteria_configs"
}

func (XrayConfig) TableName() string {
	return "xray_configs"
}

func (VPSNode) TableName() string {
	return "vps_nodes"
}

func (NodeAssignment) TableName() string {
	return "node_assignments"
}

func (NodeMetric) TableName() string {
	return "node_metrics"
}

func (Deployment) TableName() string {
	return "deployments"
}

// Hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (t *TrafficStats) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Total == 0 {
		t.Total = t.Upload + t.Download
	}
	return nil
}

func (h *HysteriaConfig) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

func (v *VPSNode) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

func (na *NodeAssignment) BeforeCreate(tx *gorm.DB) error {
	if na.ID == uuid.Nil {
		na.ID = uuid.New()
	}
	return nil
}

func (nm *NodeMetric) BeforeCreate(tx *gorm.DB) error {
	if nm.ID == uuid.Nil {
		nm.ID = uuid.New()
	}
	return nil
}

func (d *Deployment) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
