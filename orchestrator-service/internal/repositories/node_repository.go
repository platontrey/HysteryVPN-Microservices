package repositories

import (
	"time"

	"hysteryVPN/orchestrator-service/internal/models"
	"hysteryVPN/orchestrator-service/internal/repositories/interfaces"
)

type NodeRepository struct {
	db interfaces.Database
}

func NewNodeRepository(db interfaces.Database) interfaces.NodeRepository {
	return &NodeRepository{db: db}
}

func (r *NodeRepository) Create(node *models.VPSNode) error {
	return r.db.Create(node).Error
}

func (r *NodeRepository) GetByID(id string) (*models.VPSNode, error) {
	var node models.VPSNode
	err := r.db.First(&node, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *NodeRepository) GetByIPAddress(ip string) (*models.VPSNode, error) {
	var node models.VPSNode
	err := r.db.First(&node, "ip_address = ?", ip).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *NodeRepository) Update(node *models.VPSNode) error {
	return r.db.Save(node).Error
}

func (r *NodeRepository) Delete(id string) error {
	return r.db.Delete(&models.VPSNode{}, "id = ?", id).Error
}

func (r *NodeRepository) List(offset, limit int, statusFilter, locationFilter string) ([]*models.VPSNode, int64, error) {
	var nodes []*models.VPSNode
	var total int64

	query := r.db.Model(&models.VPSNode{})

	// Apply filters
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	if locationFilter != "" {
		query = query.Where("location ILIKE ?", "%"+locationFilter+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get nodes with pagination
	err := query.Offset(offset).Limit(limit).Find(&nodes).Error
	if err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

func (r *NodeRepository) UpdateStatus(id, status string) error {
	return r.db.Model(&models.VPSNode{}).Where("id = ?", id).Update("status", status).Error
}

func (r *NodeRepository) UpdateLastHeartbeat(id string, heartbeat time.Time) error {
	return r.db.Model(&models.VPSNode{}).Where("id = ?", id).Update("last_heartbeat", heartbeat).Error
}

func (r *NodeRepository) GetOnlineNodes() ([]*models.VPSNode, error) {
	var nodes []*models.VPSNode
	err := r.db.Where("status = ?", models.NodeStatusOnline).Find(&nodes).Error
	return nodes, err
}
