package repositories

import (
	"context"

	"hysteria2_microservices/api-service/internal/models"
	repoInterfaces "hysteria2_microservices/api-service/internal/repositories/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type XrayConfigRepositoryImpl struct {
	db     *gorm.DB
	logger *logger.Logger
}

func NewXrayConfigRepository(db *gorm.DB, logger *logger.Logger) repoInterfaces.XrayConfigRepository {
	return &XrayConfigRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *XrayConfigRepositoryImpl) Create(ctx context.Context, config *models.XrayConfig) error {
	r.logger.Infof("Creating Xray config for user %s", config.UserID)

	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		r.logger.Errorf("Failed to create Xray config: %v", err)
		return err
	}

	r.logger.Infof("Xray config created successfully: %s", config.ID)
	return nil
}

func (r *XrayConfigRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.XrayConfig, error) {
	r.logger.Infof("Getting Xray configs for user %s", userID)

	var configs []*models.XrayConfig
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("User").
		Preload("Device").
		Find(&configs).Error; err != nil {
		r.logger.Errorf("Failed to get Xray configs for user %s: %v", userID, err)
		return nil, err
	}

	return configs, nil
}

func (r *XrayConfigRepositoryImpl) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*models.XrayConfig, error) {
	r.logger.Infof("Getting active Xray configs for user %s", userID)

	var configs []*models.XrayConfig
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Preload("User").
		Preload("Device").
		Find(&configs).Error; err != nil {
		r.logger.Errorf("Failed to get active Xray configs for user %s: %v", userID, err)
		return nil, err
	}

	return configs, nil
}

func (r *XrayConfigRepositoryImpl) Update(ctx context.Context, config *models.XrayConfig) error {
	r.logger.Infof("Updating Xray config %s", config.ID)

	if err := r.db.WithContext(ctx).Save(config).Error; err != nil {
		r.logger.Errorf("Failed to update Xray config %s: %v", config.ID, err)
		return err
	}

	r.logger.Infof("Xray config updated successfully: %s", config.ID)
	return nil
}

func (r *XrayConfigRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Infof("Deleting Xray config %s", id)

	if err := r.db.WithContext(ctx).Delete(&models.XrayConfig{}, id).Error; err != nil {
		r.logger.Errorf("Failed to delete Xray config %s: %v", id, err)
		return err
	}

	r.logger.Infof("Xray config deleted successfully: %s", id)
	return nil
}

func (r *XrayConfigRepositoryImpl) SetActive(ctx context.Context, userID uuid.UUID, deviceID *uuid.UUID, active bool) error {
	r.logger.Infof("Setting Xray configs active status for user %s, device %v, active: %v", userID, deviceID, active)

	query := r.db.WithContext(ctx).Model(&models.XrayConfig{}).Where("user_id = ?", userID)

	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}

	if err := query.Update("is_active", active).Error; err != nil {
		r.logger.Errorf("Failed to update active status for Xray configs: %v", err)
		return err
	}

	r.logger.Info("Xray config active status updated successfully")
	return nil
}
