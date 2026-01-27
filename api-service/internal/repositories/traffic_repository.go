package repositories

import (
	"context"
	"time"

	"hysteria2_microservices/api-service/internal/models"
	repoInterfaces "hysteria2_microservices/api-service/internal/repositories/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type trafficRepository struct {
	db *gorm.DB
}

func NewTrafficRepository(db *gorm.DB) repoInterfaces.TrafficRepository {
	return &trafficRepository{db: db}
}

func (r *trafficRepository) Create(ctx context.Context, traffic *models.TrafficStats) error {
	return r.db.WithContext(ctx).Create(traffic).Error
}

func (r *trafficRepository) GetByUserID(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*models.TrafficStats, error) {
	var traffic []*models.TrafficStats
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND recorded_at BETWEEN ? AND ?", userID, from, to).
		Order("recorded_at DESC").
		Find(&traffic).Error
	return traffic, err
}

func (r *trafficRepository) GetByDeviceID(ctx context.Context, deviceID uuid.UUID, from, to time.Time) ([]*models.TrafficStats, error) {
	var traffic []*models.TrafficStats
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND recorded_at BETWEEN ? AND ?", deviceID, from, to).
		Order("recorded_at DESC").
		Find(&traffic).Error
	return traffic, err
}

func (r *trafficRepository) GetSummary(ctx context.Context, from, to time.Time) (*models.TrafficSummary, error) {
	summary := &models.TrafficSummary{
		From: from,
		To:   to,
	}

	// Get total users and active users
	var totalUsers int64
	r.db.WithContext(ctx).Model(&models.User{}).Count(&totalUsers)
	summary.TotalUsers = totalUsers

	var activeUsers int64
	r.db.WithContext(ctx).Model(&models.User{}).Where("status = ?", "active").Count(&activeUsers)
	summary.ActiveUsers = activeUsers

	// Get total traffic
	var trafficResult struct {
		TotalUpload   int64
		TotalDownload int64
	}
	r.db.WithContext(ctx).Model(&models.TrafficStats{}).
		Where("recorded_at BETWEEN ? AND ?", from, to).
		Select("COALESCE(SUM(upload), 0) as total_upload, COALESCE(SUM(download), 0) as total_download").
		Scan(&trafficResult)

	summary.TotalUpload = trafficResult.TotalUpload
	summary.TotalDownload = trafficResult.TotalDownload
	summary.TotalDataTransfer = trafficResult.TotalUpload + trafficResult.TotalDownload

	// Get top users by traffic
	var topUsers []models.UserTrafficRank
	r.db.WithContext(ctx).Model(&models.TrafficStats{}).
		Select("users.id, users.username, COALESCE(SUM(traffic_stats.upload), 0) as upload, COALESCE(SUM(traffic_stats.download), 0) as download").
		Joins("JOIN users ON traffic_stats.user_id = users.id").
		Where("traffic_stats.recorded_at BETWEEN ? AND ?", from, to).
		Group("users.id, users.username").
		Order("COALESCE(SUM(traffic_stats.upload + traffic_stats.download), 0) DESC").
		Limit(10).
		Scan(&topUsers)

	// Calculate total for top users
	for i := range topUsers {
		topUsers[i].Total = topUsers[i].Upload + topUsers[i].Download
	}
	summary.TopUsers = topUsers

	// Get top devices by traffic
	var topDevices []models.DeviceTrafficRank
	r.db.WithContext(ctx).Model(&models.TrafficStats{}).
		Select("devices.id as device_id, devices.name as device_name, users.id as user_id, users.username, COALESCE(SUM(traffic_stats.upload), 0) as upload, COALESCE(SUM(traffic_stats.download), 0) as download").
		Joins("JOIN devices ON traffic_stats.device_id = devices.id").
		Joins("JOIN users ON traffic_stats.user_id = users.id").
		Where("traffic_stats.recorded_at BETWEEN ? AND ?", from, to).
		Group("devices.id, devices.name, users.id, users.username").
		Order("COALESCE(SUM(traffic_stats.upload + traffic_stats.download), 0) DESC").
		Limit(10).
		Scan(&topDevices)

	// Calculate total for top devices
	for i := range topDevices {
		topDevices[i].Total = topDevices[i].Upload + topDevices[i].Download
	}
	summary.TopDevices = topDevices

	return summary, nil
}

func (r *trafficRepository) UpdateUserTraffic(ctx context.Context, userID uuid.UUID, upload, download int64) error {
	// Insert new traffic record
	traffic := &models.TrafficStats{
		UserID:     userID,
		Upload:     upload,
		Download:   download,
		RecordedAt: time.Now(),
	}
	return r.Create(ctx, traffic)
}

func (r *trafficRepository) UpdateDeviceTraffic(ctx context.Context, deviceID uuid.UUID, upload, download int64) error {
	// Insert new traffic record for device
	traffic := &models.TrafficStats{
		DeviceID:   &deviceID,
		Upload:     upload,
		Download:   download,
		RecordedAt: time.Now(),
	}
	return r.Create(ctx, traffic)
}
