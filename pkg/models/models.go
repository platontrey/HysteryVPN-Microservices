package models

import (
	"time"

	"github.com/google/uuid"
)

// BaseUser contains common fields for User models across services
type BaseUser struct {
	ID         uuid.UUID  `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	FullName   string     `json:"full_name"`
	Status     string     `json:"status"`
	Role       string     `json:"role"`
	DataLimit  int64      `json:"data_limit"`
	DataUsed   int64      `json:"data_used"`
	ExpiryDate *time.Time `json:"expiry_date"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastLogin  *time.Time `json:"last_login"`
	Notes      string     `json:"notes"`
}

// VPSNode represents a VPS node in the system
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
}

// NodeAssignment represents the assignment of a user to a node
type NodeAssignment struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	NodeID     uuid.UUID `json:"node_id" gorm:"not null;index"`
	UserID     uuid.UUID `json:"user_id" gorm:"not null;index"`
	AssignedAt time.Time `json:"assigned_at" gorm:"default:CURRENT_TIMESTAMP"`
	Status     string    `json:"status" gorm:"size:20;default:'active'"`
}

// NodeMetric represents metrics collected from a node
type NodeMetric struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	NodeID    uuid.UUID `json:"node_id" gorm:"not null;index"`
	Metric    string    `json:"metric" gorm:"size:100;not null"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP"`
}
