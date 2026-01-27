package interfaces

import (
	"context"
	"hysteria2_microservices/api-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.User, error)
	GenerateTokenPair(userID uuid.UUID) (*TokenPair, error)
	ValidateToken(token string) (*Claims, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
	InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error
}

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, page, limit int, search, status, role string) ([]*models.User, int64, error)
	GetUserDevices(ctx context.Context, userID uuid.UUID) ([]*models.Device, error)
	UpdateUserDataUsage(ctx context.Context, userID uuid.UUID, dataUsed int64) error
}

type NodeService interface {
	CreateNode(ctx context.Context, node *models.VPSNode) error
	GetNodeByID(ctx context.Context, id uuid.UUID) (*models.VPSNode, error)
	UpdateNode(ctx context.Context, node *models.VPSNode) error
	DeleteNode(ctx context.Context, id uuid.UUID) error
	ListNodes(ctx context.Context, page, limit int, statusFilter, locationFilter string) ([]*models.VPSNode, int64, error)
	GetNodeMetrics(ctx context.Context, nodeID uuid.UUID, limit int) ([]*models.NodeMetric, error)
	RestartNode(ctx context.Context, nodeID uuid.UUID) error
	GetNodeLogs(ctx context.Context, nodeID uuid.UUID, lines int) ([]string, error)
	UpdateNodeStatus(ctx context.Context, nodeID uuid.UUID, status string) error
	GetOnlineNodes(ctx context.Context) ([]*models.VPSNode, error)
}

type TrafficService interface {
	RecordTraffic(ctx context.Context, stats *models.TrafficStats) error
	GetUserTraffic(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*models.TrafficStats, error)
	GetTrafficSummary(ctx context.Context, from, to time.Time) (*models.TrafficSummary, error)
	UpdateUserTraffic(ctx context.Context, userID uuid.UUID, upload, download int64) error
	UpdateDeviceTraffic(ctx context.Context, deviceID uuid.UUID, upload, download int64) error
}

type HysteriaService interface {
	GenerateUserConfig(ctx context.Context, userID, deviceID string) (*models.HysteriaConfig, error)
	UpdateUserConfig(ctx context.Context, userID, deviceID string, config *models.HysteriaConfig) error
	ReloadConfiguration(ctx context.Context) error
	GetActiveConnections(ctx context.Context) ([]models.Connection, error)
	DisconnectUser(ctx context.Context, userID, deviceID string) error
	GetServiceStatus(ctx context.Context) (*models.ServiceStatus, error)
}

type XrayService interface {
	GenerateUserConfig(ctx context.Context, userID, deviceID string, protocol string) (*models.XrayConfig, error)
	UpdateUserConfig(ctx context.Context, userID, deviceID string, config *models.XrayConfig) error
	GetSupportedProtocols(ctx context.Context) ([]string, error)
	ReloadConfiguration(ctx context.Context) error
	GetActiveConnections(ctx context.Context) ([]models.Connection, error)
	DisconnectUser(ctx context.Context, userID, deviceID string) error
	GetServiceStatus(ctx context.Context) (*models.ServiceStatus, error)
}

type WebSocketService interface {
	BroadcastTrafficUpdate(userID uuid.UUID, stats *models.TrafficStats)
	BroadcastUserStatus(userID uuid.UUID, status string)
	BroadcastDeviceStatus(deviceID uuid.UUID, userID uuid.UUID, online bool)
	GetConnectedClientsCount() int
	IsUserConnected(userID uuid.UUID) bool
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
