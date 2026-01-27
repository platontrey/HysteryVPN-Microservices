package repositories

import (
	"hysteria2_microservices/orchestrator-service/internal/repositories/interfaces"
)

type Repositories struct {
	NodeRepo       interfaces.NodeRepository
	AssignmentRepo interfaces.NodeAssignmentRepository
	MetricRepo     interfaces.NodeMetricRepository
	DeploymentRepo interfaces.DeploymentRepository
	UserRepo       interfaces.UserRepository
}
