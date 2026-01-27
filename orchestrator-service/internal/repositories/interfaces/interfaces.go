package interfaces

import (
	"time"

	"hysteryVPN/orchestrator-service/internal/models"
)

// NodeRepository defines operations for VPS node management
type NodeRepository interface {
	Create(node *models.VPSNode) error
	GetByID(id string) (*models.VPSNode, error)
	GetByIPAddress(ip string) (*models.VPSNode, error)
	Update(node *models.VPSNode) error
	Delete(id string) error
	List(offset, limit int, statusFilter, locationFilter string) ([]*models.VPSNode, int64, error)
	UpdateStatus(id, status string) error
	UpdateLastHeartbeat(id string, heartbeat time.Time) error
	GetOnlineNodes() ([]*models.VPSNode, error)
}

// NodeAssignmentRepository defines operations for user-node assignments
type NodeAssignmentRepository interface {
	Create(assignment *models.NodeAssignment) error
	GetByUserID(userID string) ([]*models.NodeAssignment, error)
	GetByNodeID(nodeID string) ([]*models.NodeAssignment, error)
	Update(assignment *models.NodeAssignment) error
	Delete(id string) error
	DeleteByUserID(userID string) error
	DeleteByNodeID(nodeID string) error
	GetActiveAssignments(userID string) (*models.NodeAssignment, error)
}

// NodeMetricRepository defines operations for node metrics
type NodeMetricRepository interface {
	Create(metric *models.NodeMetric) error
	GetByNodeID(nodeID string, limit int) ([]*models.NodeMetric, error)
	GetLatest(nodeID string) (*models.NodeMetric, error)
	GetByTimeRange(nodeID string, startTime, endTime time.Time) ([]*models.NodeMetric, error)
	DeleteOldMetrics(before time.Time) error
	GetAverageMetrics(nodeID string, duration time.Duration) (*models.NodeMetric, error)
}

// DeploymentRepository defines operations for deployment tracking
type DeploymentRepository interface {
	Create(deployment *models.Deployment) error
	GetByID(id string) (*models.Deployment, error)
	GetByNodeID(nodeID string, limit int) ([]*models.Deployment, error)
	Update(deployment *models.Deployment) error
	GetLatestDeployment(nodeID string) (*models.Deployment, error)
	GetPendingDeployments() ([]*models.Deployment, error)
}

// UserRepository defines operations for user management
type UserRepository interface {
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id string) error
	List(offset, limit int) ([]*models.User, int64, error)
	Search(query string, offset, limit int) ([]*models.User, int64, error)
}
