package repositories

import (
	"hysteryVPN/orchestrator-service/internal/database"
)

type Repositories struct {
	NodeRepo       NodeRepository
	AssignmentRepo NodeAssignmentRepository
	MetricRepo     NodeMetricRepository
	DeploymentRepo DeploymentRepository
	UserRepo       UserRepository
}

type NodeRepository interface {
	Create(node *database.Node) error
	GetByID(id string) (*database.Node, error)
	GetByIPAddress(ip string) (*database.Node, error)
	Update(node *database.Node) error
	Delete(id string) error
	List(offset, limit int, statusFilter, locationFilter string) ([]*database.Node, int64, error)
	UpdateStatus(id, status string) error
	UpdateLastHeartbeat(id string, heartbeat time.Time) error
	GetOnlineNodes() ([]*database.Node, error)
}

type NodeAssignmentRepository interface {
	Create(assignment *database.NodeAssignment) error
	GetByUserID(userID string) ([]*database.NodeAssignment, error)
	GetByNodeID(nodeID string) ([]*database.NodeAssignment, error)
	Update(assignment *database.NodeAssignment) error
	Delete(id string) error
}

type NodeMetricRepository interface {
	Create(metric *database.NodeMetric) error
	GetByNodeID(nodeID string, limit int) ([]*database.NodeMetric, error)
	GetLatest(nodeID string) (*database.NodeMetric, error)
}

type DeploymentRepository interface {
	Create(deployment *database.Deployment) error
	GetByNodeID(nodeID string, limit int) ([]*database.Deployment, error)
}

type UserRepository interface {
	GetByID(id string) (*database.User, error)
}
