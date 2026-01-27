package services

import (
	"context"
	"hysteria2_microservices/api-service/internal/models"
	repoInterfaces "hysteria2_microservices/api-service/internal/repositories/interfaces"
	serviceInterfaces "hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/google/uuid"
)

type nodeService struct {
	nodeRepo repoInterfaces.NodeRepository
	logger   *logger.Logger
}

func NewNodeService(nodeRepo repoInterfaces.NodeRepository, logger *logger.Logger) serviceInterfaces.NodeService {
	return &nodeService{
		nodeRepo: nodeRepo,
		logger:   logger,
	}
}

func (s *nodeService) CreateNode(ctx context.Context, node *models.VPSNode) error {
	s.logger.Info("Creating new VPS node", "name", node.Name, "ip", node.IPAddress)
	return s.nodeRepo.Create(ctx, node)
}

func (s *nodeService) GetNodeByID(ctx context.Context, id uuid.UUID) (*models.VPSNode, error) {
	s.logger.Debug("Getting node by ID", "id", id)
	return s.nodeRepo.GetByID(ctx, id)
}

func (s *nodeService) UpdateNode(ctx context.Context, node *models.VPSNode) error {
	s.logger.Info("Updating node", "id", node.ID, "name", node.Name)
	return s.nodeRepo.Update(ctx, node)
}

func (s *nodeService) DeleteNode(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("Deleting node", "id", id)
	return s.nodeRepo.Delete(ctx, id)
}

func (s *nodeService) ListNodes(ctx context.Context, page, limit int, statusFilter, locationFilter string) ([]*models.VPSNode, int64, error) {
	s.logger.Debug("Listing nodes", "page", page, "limit", limit, "status", statusFilter, "location", locationFilter)
	return s.nodeRepo.List(ctx, page, limit, statusFilter, locationFilter)
}

func (s *nodeService) GetNodeMetrics(ctx context.Context, nodeID uuid.UUID, limit int) ([]*models.NodeMetric, error) {
	s.logger.Debug("Getting node metrics", "node_id", nodeID, "limit", limit)
	// This would need a metrics repository - for now return empty
	return []*models.NodeMetric{}, nil
}

func (s *nodeService) RestartNode(ctx context.Context, nodeID uuid.UUID) error {
	s.logger.Info("Restarting node", "node_id", nodeID)
	// This would involve sending a command to the agent
	// For now, just update status
	return s.nodeRepo.UpdateStatus(ctx, nodeID, "restarting")
}

func (s *nodeService) GetNodeLogs(ctx context.Context, nodeID uuid.UUID, lines int) ([]string, error) {
	s.logger.Debug("Getting node logs", "node_id", nodeID, "lines", lines)
	// This would involve fetching logs from the agent
	// For now, return mock logs
	return []string{"Node started", "Configuration loaded", "Service running"}, nil
}

func (s *nodeService) UpdateNodeStatus(ctx context.Context, nodeID uuid.UUID, status string) error {
	s.logger.Info("Updating node status", "node_id", nodeID, "status", status)
	return s.nodeRepo.UpdateStatus(ctx, nodeID, status)
}

func (s *nodeService) GetOnlineNodes(ctx context.Context) ([]*models.VPSNode, error) {
	s.logger.Debug("Getting online nodes")
	return s.nodeRepo.GetOnlineNodes(ctx)
}
