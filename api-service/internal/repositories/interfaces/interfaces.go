package interfaces

import (
	"context"
	"hysteria2_microservices/api-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int, search string, status, role string) ([]*models.User, int64, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateDataUsage(ctx context.Context, id uuid.UUID, dataUsed int64) error
}

type DeviceRepository interface {
	Create(ctx context.Context, device *models.Device) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error)
	GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Device, error)
	Update(ctx context.Context, device *models.Device) error
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
	UpdateDataUsage(ctx context.Context, id uuid.UUID, dataUsed int64) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByToken(ctx context.Context, token string) (*models.Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error
}

type TrafficRepository interface {
	Create(ctx context.Context, traffic *models.TrafficStats) error
	GetByUserID(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*models.TrafficStats, error)
	GetByDeviceID(ctx context.Context, deviceID uuid.UUID, from, to time.Time) ([]*models.TrafficStats, error)
	GetSummary(ctx context.Context, from, to time.Time) (*models.TrafficSummary, error)
	UpdateUserTraffic(ctx context.Context, userID uuid.UUID, upload, download int64) error
	UpdateDeviceTraffic(ctx context.Context, deviceID uuid.UUID, upload, download int64) error
}

type HysteriaConfigRepository interface {
	Create(ctx context.Context, config *models.HysteriaConfig) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.HysteriaConfig, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*models.HysteriaConfig, error)
	Update(ctx context.Context, config *models.HysteriaConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, userID uuid.UUID, deviceID *uuid.UUID, active bool) error
}

type XrayConfigRepository interface {
	Create(ctx context.Context, config *models.XrayConfig) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.XrayConfig, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*models.XrayConfig, error)
	Update(ctx context.Context, config *models.XrayConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, userID uuid.UUID, deviceID *uuid.UUID, active bool) error
}

type NodeRepository interface {
	Create(ctx context.Context, node *models.VPSNode) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.VPSNode, error)
	Update(ctx context.Context, node *models.VPSNode) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, limit int, statusFilter, locationFilter string) ([]*models.VPSNode, int64, error)
	GetOnlineNodes(ctx context.Context) ([]*models.VPSNode, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}
