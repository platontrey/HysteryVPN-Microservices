package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VPSNode represents a VPS server node
type VPSNode struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Hostname      string    `gorm:"size:255;not null" json:"hostname"`
	IPAddress     string    `gorm:"size:45;not null;index" json:"ip_address"`
	Location      string    `gorm:"size:100" json:"location"`
	Country       string    `gorm:"size:2" json:"country"`
	GRPCPort      int       `gorm:"default:50051" json:"grpc_port"`
	Status        string    `gorm:"size:20;default:'offline';index" json:"status"`
	Version       string    `gorm:"size:50" json:"version"`
	Capabilities  JSONB     `gorm:"type:jsonb" json:"capabilities"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	LastHeartbeat time.Time `gorm:"index" json:"last_heartbeat"`
	Metadata      JSONB     `gorm:"type:jsonb" json:"metadata"`

	// SNI-specific fields
	PrimaryDomain   string `gorm:"size:255" json:"primary_domain"`
	SNIEnabled      bool   `gorm:"default:false" json:"sni_enabled"`
	SNIDomains      JSONB  `gorm:"type:jsonb" json:"sni_domains"` // []string
	CertificatePath string `gorm:"size:500" json:"certificate_path"`
	KeyPath         string `gorm:"size:500" json:"key_path"`
	SNIAutoRenew    bool   `gorm:"default:true" json:"sni_auto_renew"`
	SNIEmail        string `gorm:"size:255" json:"sni_email"`

	// Relations
	Assignments []NodeAssignment `gorm:"foreignKey:NodeID" json:"assignments,omitempty"`
	Metrics     []NodeMetric     `gorm:"foreignKey:NodeID" json:"metrics,omitempty"`
	Deployments []Deployment     `gorm:"foreignKey:NodeID" json:"deployments,omitempty"`
}

// NodeAssignment represents user assignment to a node
type NodeAssignment struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	NodeID     uuid.UUID `gorm:"type:uuid;not null;index" json:"node_id"`
	AssignedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"assigned_at"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`

	// Relations
	User *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Node *VPSNode `gorm:"foreignKey:NodeID" json:"node,omitempty"`
}

// NodeMetric represents metrics collected from a node
type NodeMetric struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	NodeID            uuid.UUID `gorm:"type:uuid;not null;index" json:"node_id"`
	CPUUsage          float64   `gorm:"type:decimal(5,2)" json:"cpu_usage"`
	MemoryUsage       float64   `gorm:"type:decimal(5,2)" json:"memory_usage"`
	BandwidthUp       int64     `json:"bandwidth_up"`
	BandwidthDown     int64     `json:"bandwidth_down"`
	ActiveConnections int       `json:"active_connections"`
	RecordedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP;index" json:"recorded_at"`

	// Relations
	Node *VPSNode `gorm:"foreignKey:NodeID" json:"node,omitempty"`
}

// Deployment represents configuration deployment to a node
type Deployment struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	NodeID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"node_id"`
	ConfigVersion string     `gorm:"size:50;not null" json:"config_version"`
	Status        string     `gorm:"size:20;default:'pending'" json:"status"`
	DeployedAt    *time.Time `json:"deployed_at"`
	RollbackAt    *time.Time `json:"rollback_at"`
	ErrorMessage  string     `gorm:"type:text" json:"error_message"`

	// Relations
	Node *VPSNode `gorm:"foreignKey:NodeID" json:"node,omitempty"`
}

// User model (simplified version for this service)
type User struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	Username   string     `gorm:"size:50;unique;not null;index" json:"username"`
	Email      string     `gorm:"size:255;unique;not null;index" json:"email"`
	FullName   string     `gorm:"size:100" json:"full_name"`
	Status     string     `gorm:"size:20;default:'active';index" json:"status"`
	Role       string     `gorm:"size:20;default:'user'" json:"role"`
	DataLimit  int64      `gorm:"default:0" json:"data_limit"`
	DataUsed   int64      `gorm:"default:0" json:"data_used"`
	ExpiryDate *time.Time `json:"expiry_date"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	LastLogin  *time.Time `json:"last_login"`
	Notes      string     `gorm:"type:text" json:"notes"`

	// Relations
	Assignments []NodeAssignment `gorm:"foreignKey:UserID" json:"assignments,omitempty"`
}

// JSONB type for PostgreSQL JSONB fields
type JSONB map[string]interface{}

// Value implements driver.Valuer interface
func (j JSONB) Value() (interface{}, error) {
	if j == nil {
		return nil, nil
	}
	return j, nil
}

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case map[string]interface{}:
		*j = v
	case []byte:
		// Handle byte array conversion if needed
		*j = make(map[string]interface{})
		// Simple JSON parsing could be added here
	default:
		*j = make(map[string]interface{})
	}

	return nil
}

// BeforeCreate hook for UUID generation
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

// TableName methods for custom table names
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

// Helper methods
func (n *VPSNode) IsOnline() bool {
	return n.Status == "online"
}

func (n *VPSNode) GetCapability(key string) (interface{}, bool) {
	if n.Capabilities == nil {
		return nil, false
	}
	value, exists := n.Capabilities[key]
	return value, exists
}

func (n *VPSNode) GetMetadata(key string) (interface{}, bool) {
	if n.Metadata == nil {
		return nil, false
	}
	value, exists := n.Metadata[key]
	return value, exists
}

// SNI-related helper methods
func (n *VPSNode) GetSNIDomains() []string {
	if n.SNIDomains == nil {
		return []string{}
	}

	// Try to extract domains array from JSONB
	if domainsArray, ok := n.SNIDomains["domains"]; ok {
		if domains, ok := domainsArray.([]interface{}); ok {
			result := make([]string, len(domains))
			for i, domain := range domains {
				if str, ok := domain.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return []string{}
}

func (n *VPSNode) SetSNIDomains(domains []string) {
	if len(domains) == 0 {
		n.SNIDomains = JSONB{}
		return
	}

	interfaceDomains := make([]interface{}, len(domains))
	for i, domain := range domains {
		interfaceDomains[i] = domain
	}
	n.SNIDomains = JSONB{
		"domains": interfaceDomains,
	}
}

func (n *VPSNode) HasSNIDomain(domain string) bool {
	domains := n.GetSNIDomains()
	for _, d := range domains {
		if d == domain {
			return true
		}
	}
	return false
}

func (n *VPSNode) AddSNIDomain(domain string) {
	domains := n.GetSNIDomains()
	if !n.HasSNIDomain(domain) {
		domains = append(domains, domain)
		n.SetSNIDomains(domains)
	}
}

func (n *VPSNode) RemoveSNIDomain(domain string) {
	domains := n.GetSNIDomains()
	for i, d := range domains {
		if d == domain {
			domains = append(domains[:i], domains[i+1:]...)
			break
		}
	}
	n.SetSNIDomains(domains)
}

// Constants
const (
	NodeStatusOffline     = "offline"
	NodeStatusOnline      = "online"
	NodeStatusMaintenance = "maintenance"
	NodeStatusError       = "error"

	DeploymentStatusPending   = "pending"
	DeploymentStatusDeploying = "deploying"
	DeploymentStatusSuccess   = "success"
	DeploymentStatusFailed    = "failed"

	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusDeleted   = "deleted"

	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)
