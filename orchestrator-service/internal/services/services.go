package services

import (
	"hysteria2_microservices/orchestrator-service/internal/repositories"
)

type Services struct {
	NodeService       NodeService
	AssignmentService AssignmentService
	MetricsService    MetricsService
	DeploymentService DeploymentService
	UserService       UserService
}

func NewServices(repos *repositories.Repositories) *Services {
	return &Services{
		NodeService:       NewNodeService(repos.NodeRepo),
		AssignmentService: NewAssignmentService(repos.AssignmentRepo),
		MetricsService:    NewMetricsService(repos.MetricRepo),
		DeploymentService: NewDeploymentService(repos.DeploymentRepo),
		UserService:       NewUserService(repos.UserRepo),
	}
}
