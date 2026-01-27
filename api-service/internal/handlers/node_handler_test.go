package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"hysteria2_microservices/api-service/internal/models"
	"hysteria2_microservices/api-service/internal/services/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockNodeService is a mock implementation of NodeService
type MockNodeService struct {
	mock.Mock
}

func (m *MockNodeService) CreateNode(ctx context.Context, node *models.VPSNode) error {
	args := m.Called(ctx, node)
	return args.Error(0)
}

func (m *MockNodeService) GetNodeByID(ctx context.Context, id uuid.UUID) (*models.VPSNode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VPSNode), args.Error(1)
}

func (m *MockNodeService) UpdateNode(ctx context.Context, node *models.VPSNode) error {
	args := m.Called(ctx, node)
	return args.Error(0)
}

func (m *MockNodeService) DeleteNode(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNodeService) ListNodes(ctx context.Context, page, limit int, statusFilter, locationFilter string) ([]*models.VPSNode, int64, error) {
	args := m.Called(ctx, page, limit, statusFilter, locationFilter)
	return args.Get(0).([]*models.VPSNode), args.Get(1).(int64), args.Error(2)
}

func (m *MockNodeService) GetNodeMetrics(ctx context.Context, nodeID uuid.UUID, limit int) ([]*models.NodeMetric, error) {
	args := m.Called(ctx, nodeID, limit)
	return args.Get(0).([]*models.NodeMetric), args.Error(1)
}

func (m *MockNodeService) RestartNode(ctx context.Context, nodeID uuid.UUID) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

func (m *MockNodeService) GetNodeLogs(ctx context.Context, nodeID uuid.UUID, lines int) ([]string, error) {
	args := m.Called(ctx, nodeID, lines)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockNodeService) UpdateNodeStatus(ctx context.Context, nodeID uuid.UUID, status string) error {
	args := m.Called(ctx, nodeID, status)
	return args.Error(0)
}

func (m *MockNodeService) GetOnlineNodes(ctx context.Context) ([]*models.VPSNode, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.VPSNode), args.Error(1)
}

type NodeHandlerTestSuite struct {
	suite.Suite
	app         *fiber.App
	mockService *MockNodeService
	mockLogger  *MockLogger
	nodeHandler *NodeHandler
	testNode    *models.VPSNode
	testNodeID  uuid.UUID
}

func (suite *NodeHandlerTestSuite) SetupTest() {
	suite.mockService = new(MockNodeService)
	suite.mockLogger = new(MockLogger)
	suite.nodeHandler = NewNodeHandler(suite.mockService, suite.mockLogger)
	suite.app = fiber.New()
	suite.testNodeID = uuid.New()
	suite.testNode = &models.VPSNode{
		ID:            suite.testNodeID,
		Name:          "Test Node",
		Hostname:      "test.example.com",
		IPAddress:     "192.168.1.100",
		Location:      "New York",
		Country:       "US",
		GRPCPort:      50051,
		Status:        "online",
		CreatedAt:     time.Now(),
		LastHeartbeat: &time.Time{},
		Capabilities:  map[string]interface{}{"hysteria": true},
		Metadata:      map[string]interface{}{"version": "1.0.0"},
	}

	// Setup routes
	suite.app.Get("/nodes", suite.nodeHandler.GetNodes)
	suite.app.Get("/nodes/:id", suite.nodeHandler.GetNode)
	suite.app.Post("/nodes", suite.nodeHandler.CreateNode)
	suite.app.Put("/nodes/:id", suite.nodeHandler.UpdateNode)
	suite.app.Delete("/nodes/:id", suite.nodeHandler.DeleteNode)
	suite.app.Get("/nodes/:id/metrics", suite.nodeHandler.GetNodeMetrics)
	suite.app.Post("/nodes/:id/restart", suite.nodeHandler.RestartNode)
	suite.app.Get("/nodes/:id/logs", suite.nodeHandler.GetNodeLogs)
}

func (suite *NodeHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

func (suite *NodeHandlerTestSuite) TestGetNodes_Success() {
	nodes := []*models.VPSNode{suite.testNode}
	total := int64(1)

	suite.mockService.On("ListNodes", mock.Anything, 1, 10, "", "").Return(nodes, total, nil)

	req := httptest.NewRequest("GET", "/nodes", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal(float64(1), response["total"])
}

func (suite *NodeHandlerTestSuite) TestGetNode_Success() {
	suite.mockService.On("GetNodeByID", mock.Anything, suite.testNodeID).Return(suite.testNode, nil)

	req := httptest.NewRequest("GET", "/nodes/"+suite.testNodeID.String(), nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response models.VPSNode
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal(suite.testNode.ID, response.ID)
	suite.Equal(suite.testNode.Name, response.Name)
}

func (suite *NodeHandlerTestSuite) TestCreateNode_Success() {
	createReq := CreateNodeRequest{
		Name:         "New Node",
		Hostname:     "new.example.com",
		IPAddress:    "192.168.1.101",
		Location:     "London",
		Country:      "GB",
		GRPCPort:     50052,
		Capabilities: map[string]string{"hysteria": "true"},
		Metadata:     map[string]string{"version": "1.1.0"},
	}

	suite.mockService.On("CreateNode", mock.Anything, mock.MatchedBy(func(node *models.VPSNode) bool {
		return node.Name == createReq.Name && node.Hostname == createReq.Hostname
	})).Return(nil).Run(func(args mock.Arguments) {
		node := args.Get(1).(*models.VPSNode)
		node.ID = uuid.New()
		node.CreatedAt = time.Now()
	})

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/nodes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusCreated, resp.StatusCode)
}

func (suite *NodeHandlerTestSuite) TestUpdateNode_Success() {
	updateReq := UpdateNodeRequest{
		Name:     stringPtr("Updated Node"),
		Location: stringPtr("Updated Location"),
		Status:   stringPtr("maintenance"),
	}

	updatedNode := *suite.testNode
	updatedNode.Name = *updateReq.Name
	updatedNode.Location = *updateReq.Location
	updatedNode.Status = *updateReq.Status

	suite.mockService.On("GetNodeByID", mock.Anything, suite.testNodeID).Return(suite.testNode, nil)
	suite.mockService.On("UpdateNode", mock.Anything, mock.MatchedBy(func(node *models.VPSNode) bool {
		return node.ID == suite.testNodeID && node.Name == "Updated Node"
	})).Return(nil)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/nodes/"+suite.testNodeID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)
}

func (suite *NodeHandlerTestSuite) TestDeleteNode_Success() {
	suite.mockService.On("DeleteNode", mock.Anything, suite.testNodeID).Return(nil)

	req := httptest.NewRequest("DELETE", "/nodes/"+suite.testNodeID.String(), nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusNoContent, resp.StatusCode)
}

func (suite *NodeHandlerTestSuite) TestGetNodeMetrics_Success() {
	metrics := []*models.NodeMetric{
		{
			ID:            uuid.New(),
			NodeID:        suite.testNodeID,
			CPUUsage:      45.5,
			MemoryUsage:   67.8,
			BandwidthUp:   1024000,
			BandwidthDown: 2048000,
			RecordedAt:    time.Now(),
		},
	}

	suite.mockService.On("GetNodeMetrics", mock.Anything, suite.testNodeID, 100).Return(metrics, nil)

	req := httptest.NewRequest("GET", "/nodes/"+suite.testNodeID.String()+"/metrics", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Len(response["metrics"], 1)
}

func (suite *NodeHandlerTestSuite) TestRestartNode_Success() {
	suite.mockService.On("RestartNode", mock.Anything, suite.testNodeID).Return(nil)

	req := httptest.NewRequest("POST", "/nodes/"+suite.testNodeID.String()+"/restart", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Node restart initiated", response["message"])
}

func (suite *NodeHandlerTestSuite) TestGetNodeLogs_Success() {
	logs := []string{
		"[2023-01-01 12:00:00] Server started",
		"[2023-01-01 12:01:00] Client connected",
	}

	suite.mockService.On("GetNodeLogs", mock.Anything, suite.testNodeID, 100).Return(logs, nil)

	req := httptest.NewRequest("GET", "/nodes/"+suite.testNodeID.String()+"/logs", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Len(response["logs"], 2)
	suite.Equal(float64(100), response["lines"])
}

func TestNodeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(NodeHandlerTestSuite))
}
