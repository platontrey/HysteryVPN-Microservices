// Test fixtures for HysteriaVPN Microservices
package testfixtures

import (
	"time"

	"hysteria2_microservices/api-service/internal/models"
	"hysteria2_microservices/api-service/internal/services/interfaces"

	"github.com/google/uuid"
)

// User fixtures
func CreateTestUser() *models.User {
	return &models.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FullName:  stringPtr("Test User"),
		Status:    "active",
		Role:      "user",
		DataLimit: 1000000,
		DataUsed:  500000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func CreateTestUserList() []*models.User {
	return []*models.User{
		CreateTestUser(),
		{
			ID:        uuid.New(),
			Username:  "user2",
			Email:     "user2@example.com",
			Password:  "hashedpassword2",
			Status:    "active",
			Role:      "user",
			DataLimit: 2000000,
			DataUsed:  1000000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// Device fixtures
func CreateTestDevice(userID uuid.UUID) *models.Device {
	return &models.Device{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "Test Device",
		DeviceID:  "device123",
		PublicKey: "testpublickey",
		Status:    "active",
		DataUsed:  100000,
		CreatedAt: time.Now(),
		LastSeen:  &time.Time{},
	}
}

// VPS Node fixtures
func CreateTestVPSNode() *models.VPSNode {
	return &models.VPSNode{
		ID:        uuid.New(),
		Name:      "Test Node",
		Hostname:  "test.example.com",
		IPAddress: "192.168.1.100",
		Location:  "New York",
		Country:   "US",
		GRPCPort:  50051,
		Status:    "online",
		Version:   "1.0.0",
		Capabilities: map[string]interface{}{
			"hysteria": true,
			"network":  true,
		},
		Metadata: map[string]interface{}{
			"datacenter": "aws-us-east",
		},
		CreatedAt:     time.Now(),
		LastHeartbeat: &time.Time{},
	}
}

// Token fixtures
func CreateTestTokenPair() *interfaces.TokenPair {
	return &interfaces.TokenPair{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature",
		RefreshToken: "refresh_token_123456789",
		ExpiresIn:    3600,
	}
}

func CreateTestClaims() *interfaces.Claims {
	return &interfaces.Claims{
		UserID:   uuid.New().String(),
		Username: "testuser",
		Role:     "user",
	}
}

// Traffic fixtures
func CreateTestTrafficStats(userID, deviceID uuid.UUID) *models.TrafficStats {
	return &models.TrafficStats{
		ID:         uuid.New(),
		UserID:     userID,
		DeviceID:   &deviceID,
		Upload:     1000000,
		Download:   2000000,
		Total:      3000000,
		RecordedAt: time.Now(),
		CreatedAt:  time.Now(),
	}
}

// Node metrics fixtures
func CreateTestNodeMetrics(nodeID uuid.UUID) []*models.NodeMetric {
	return []*models.NodeMetric{
		{
			ID:                uuid.New(),
			NodeID:            nodeID,
			CPUUsage:          45.5,
			MemoryUsage:       67.8,
			BandwidthUp:       1024000,
			BandwidthDown:     2048000,
			ActiveConnections: 25,
			RecordedAt:        time.Now(),
		},
		{
			ID:                uuid.New(),
			NodeID:            nodeID,
			CPUUsage:          42.1,
			MemoryUsage:       65.3,
			BandwidthUp:       980000,
			BandwidthDown:     1950000,
			ActiveConnections: 22,
			RecordedAt:        time.Now().Add(-time.Minute * 5),
		},
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
