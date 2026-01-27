package repositories

import (
	"context"
	"hysteria2_microservices/api-service/internal/models"
	"hysteria2_microservices/api-service/internal/repositories/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type nodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) interfaces.NodeRepository {
	return &nodeRepository{db: db}
}

func (r *nodeRepository) Create(ctx context.Context, node *models.VPSNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *nodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.VPSNode, error) {
	var node models.VPSNode
	err := r.db.WithContext(ctx).Preload("Assignments").Preload("Metrics").Where("id = ?", id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *nodeRepository) Update(ctx context.Context, node *models.VPSNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

func (r *nodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.VPSNode{}, "id = ?", id).Error
}

func (r *nodeRepository) List(ctx context.Context, page, limit int, statusFilter, locationFilter string) ([]*models.VPSNode, int64, error) {
	var nodes []*models.VPSNode
	var total int64

	query := r.db.WithContext(ctx).Model(&models.VPSNode{})

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	if locationFilter != "" {
		query = query.Where("location ILIKE ?", "%"+locationFilter+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&nodes).Error
	if err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

func (r *nodeRepository) GetOnlineNodes(ctx context.Context) ([]*models.VPSNode, error) {
	var nodes []*models.VPSNode
	err := r.db.WithContext(ctx).Where("status = ?", "online").Find(&nodes).Error
	return nodes, err
}

func (r *nodeRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&models.VPSNode{}).Where("id = ?", id).Update("status", status).Error
}
