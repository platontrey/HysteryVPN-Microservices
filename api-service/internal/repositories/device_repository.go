package repositories

import (
	"context"

	"hysteria2_microservices/api-service/internal/models"
	repoInterfaces "hysteria2_microservices/api-service/internal/repositories/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) repoInterfaces.DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(ctx context.Context, device *models.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *deviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	var device models.Device
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	var device models.Device
	err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Device, error) {
	var devices []*models.Device
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) Update(ctx context.Context, device *models.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *deviceRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", id).Update("last_seen", gorm.Expr("NOW()")).Error
}

func (r *deviceRepository) UpdateDataUsage(ctx context.Context, id uuid.UUID, dataUsed int64) error {
	return r.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", id).Update("data_used", dataUsed).Error
}

func (r *deviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Device{}, "id = ?", id).Error
}
