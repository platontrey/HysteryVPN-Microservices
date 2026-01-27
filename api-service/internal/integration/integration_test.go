// Integration tests for HysteriaVPN Microservices
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"hysteria2_microservices/api-service/internal/handlers"
	"hysteria2_microservices/api-service/internal/models"
	"hysteria2_microservices/api-service/internal/testfixtures"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockOrchestratorClient simulates gRPC calls to orchestrator
type MockOrchestratorClient struct {
	mock.Mock
}

func (m *MockOrchestratorClient) GetNodeStatus(ctx context.Context, nodeID string) (string, error) {
	args := m.Called(ctx, nodeID)
	return args.String(0), args.Error(1)
}

func (m *MockOrchestratorClient) UpdateNodeConfig(ctx context.Context, nodeID string, config map[string]interface{}) error {
	args := m.Called(ctx, nodeID, config)
	return args.Error(0)
}

type IntegrationTestSuite struct {
	suite.Suite
	app                    *fiber.App
	mockUserService        *handlers.MockUserService
	mockAuthService        *handlers.MockAuthService
	mockNodeService        *handlers.MockNodeService
	mockLogger             *handlers.MockLogger
	userHandler            *handlers.UserHandler
	authHandler            *handlers.AuthHandler
	nodeHandler            *handlers.NodeHandler
	mockOrchestratorClient *MockOrchestratorClient
	testUser               *models.User
	testNode               *models.VPSNode
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Initialize mocks
	suite.mockUserService = &handlers.MockUserService{}
	suite.mockAuthService = &handlers.MockAuthService{}
	suite.mockNodeService = &handlers.MockNodeService{}
	suite.mockLogger = &handlers.MockLogger{}
	suite.mockOrchestratorClient = &MockOrchestratorClient{}

	// Initialize handlers
	suite.userHandler = handlers.NewUserHandler(suite.mockUserService, suite.mockLogger)
	suite.authHandler = handlers.NewAuthHandler(suite.mockAuthService, suite.mockLogger)
	suite.nodeHandler = handlers.NewNodeHandler(suite.mockNodeService, suite.mockLogger)

	// Setup test data
	suite.testUser = testfixtures.CreateTestUser()
	suite.testNode = testfixtures.CreateTestVPSNode()

	// Setup Fiber app with routes
	suite.app = fiber.New()

	// Auth routes
	auth := suite.app.Group("/auth")
	auth.Post("/register", suite.authHandler.Register)
	auth.Post("/login", suite.authHandler.Login)

	// Protected routes
	api := suite.app.Group("/api")
	api.Use(func(c *fiber.Ctx) error {
		// Mock authentication middleware
		c.Locals("user_id", suite.testUser.ID.String())
		c.Locals("username", suite.testUser.Username)
		c.Locals("role", "admin")
		return c.Next()
	})

	// User routes
	api.Get("/users", suite.userHandler.GetUsers)
	api.Post("/users", suite.userHandler.CreateUser)
	api.Get("/users/:id", suite.userHandler.GetUser)
	api.Put("/users/:id", suite.userHandler.UpdateUser)
	api.Delete("/users/:id", suite.userHandler.DeleteUser)

	// Node routes
	api.Get("/nodes", suite.nodeHandler.GetNodes)
	api.Post("/nodes", suite.nodeHandler.CreateNode)
	api.Get("/nodes/:id", suite.nodeHandler.GetNode)
	api.Put("/nodes/:id", suite.nodeHandler.UpdateNode)
	api.Delete("/nodes/:id", suite.nodeHandler.DeleteNode)
}

func (suite *IntegrationTestSuite) TearDownTest() {
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockAuthService.AssertExpectations(suite.T())
	suite.mockNodeService.AssertExpectations(suite.T())
	suite.mockOrchestratorClient.AssertExpectations(suite.T())
}

func (suite *IntegrationTestSuite) TestFullUserLifecycle() {
	// 1. Register a new user
	suite.mockAuthService.On("Register", mock.Anything, "integrationuser", "integration@example.com", "password123").
		Return(suite.testUser, nil)
	suite.mockAuthService.On("GenerateTokenPair", suite.testUser.ID).
		Return(testfixtures.CreateTestTokenPair(), nil)

	registerReq := handlers.RegisterRequest{
		Username: "integrationuser",
		Email:    "integration@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusCreated, resp.StatusCode)

	// 2. Login with the user
	suite.mockAuthService.On("Login", mock.Anything, "integration@example.com", "password123").
		Return(suite.testUser, nil)
	suite.mockAuthService.On("GenerateTokenPair", suite.testUser.ID).
		Return(testfixtures.CreateTestTokenPair(), nil)

	loginReq := handlers.LoginRequest{
		Email:    "integration@example.com",
		Password: "password123",
	}
	body, _ = json.Marshal(loginReq)
	req = httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// 3. Get users list
	suite.mockUserService.On("ListUsers", mock.Anything, 1, 10, "", "", "").
		Return([]*models.User{suite.testUser}, int64(1), nil)

	req = httptest.NewRequest("GET", "/api/users", nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// 4. Get specific user
	suite.mockUserService.On("GetUserByID", mock.Anything, suite.testUser.ID).
		Return(suite.testUser, nil)

	req = httptest.NewRequest("GET", "/api/users/"+suite.testUser.ID.String(), nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// 5. Update user
	updateReq := handlers.UpdateUserRequest{
		FullName: stringPtr("Updated Integration User"),
	}
	suite.mockUserService.On("GetUserByID", mock.Anything, suite.testUser.ID).
		Return(suite.testUser, nil)
	suite.mockUserService.On("UpdateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.ID == suite.testUser.ID && *user.FullName == "Updated Integration User"
	})).Return(nil)

	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest("PUT", "/api/users/"+suite.testUser.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// 6. Delete user
	suite.mockUserService.On("DeleteUser", mock.Anything, suite.testUser.ID).Return(nil)

	req = httptest.NewRequest("DELETE", "/api/users/"+suite.testUser.ID.String(), nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusNoContent, resp.StatusCode)
}

func (suite *IntegrationTestSuite) TestNodeManagementWorkflow() {
	// 1. Create a new node
	suite.mockNodeService.On("CreateNode", mock.Anything, mock.MatchedBy(func(node *models.VPSNode) bool {
		return node.Name == "Integration Node"
	})).Return(nil).Run(func(args mock.Arguments) {
		node := args.Get(1).(*models.VPSNode)
		node.ID = suite.testNode.ID
		node.CreatedAt = time.Now()
	})

	createReq := handlers.CreateNodeRequest{
		Name:      "Integration Node",
		Hostname:  "integration.example.com",
		IPAddress: "192.168.1.200",
		Location:  "Integration City",
		Country:   "IC",
		GRPCPort:  50052,
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/nodes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusCreated, resp.StatusCode)

	// 2. Get nodes list
	suite.mockNodeService.On("ListNodes", mock.Anything, 1, 10, "", "").
		Return([]*models.VPSNode{suite.testNode}, int64(1), nil)

	req = httptest.NewRequest("GET", "/api/nodes", nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// 3. Get specific node
	suite.mockNodeService.On("GetNodeByID", mock.Anything, suite.testNode.ID).
		Return(suite.testNode, nil)

	req = httptest.NewRequest("GET", "/api/nodes/"+suite.testNode.ID.String(), nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// 4. Update node status via orchestrator
	suite.mockNodeService.On("UpdateNode", mock.Anything, mock.MatchedBy(func(node *models.VPSNode) bool {
		return node.ID == suite.testNode.ID && node.Status == "maintenance"
	})).Return(nil)
	suite.mockNodeService.On("GetNodeByID", mock.Anything, suite.testNode.ID).
		Return(suite.testNode, nil)

	updateReq := handlers.UpdateNodeRequest{
		Status: stringPtr("maintenance"),
	}
	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest("PUT", "/api/nodes/"+suite.testNode.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// 5. Delete node
	suite.mockNodeService.On("DeleteNode", mock.Anything, suite.testNode.ID).Return(nil)

	req = httptest.NewRequest("DELETE", "/api/nodes/"+suite.testNode.ID.String(), nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusNoContent, resp.StatusCode)
}

func (suite *IntegrationTestSuite) TestAPIOrchestratorCommunication() {
	// Simulate API calling orchestrator for node metrics
	metrics := testfixtures.CreateTestNodeMetrics(suite.testNode.ID)

	suite.mockNodeService.On("GetNodeMetrics", mock.Anything, suite.testNode.ID, 100).
		Return(metrics, nil)

	req := httptest.NewRequest("GET", "/api/nodes/"+suite.testNode.ID.String()+"/metrics", nil)
	resp, _ := suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	assert.Len(response["metrics"], 2)
}

func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	// Test concurrent API calls
	done := make(chan bool, 2)

	// Mock user service for concurrent calls
	suite.mockUserService.On("ListUsers", mock.Anything, 1, 10, "", "", "").
		Return([]*models.User{suite.testUser}, int64(1), nil).Times(2)

	go func() {
		req := httptest.NewRequest("GET", "/api/users", nil)
		resp, _ := suite.app.Test(req)
		assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)
		done <- true
	}()

	go func() {
		req := httptest.NewRequest("GET", "/api/users", nil)
		resp, _ := suite.app.Test(req)
		assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)
		done <- true
	}()

	// Wait for both requests to complete
	<-done
	<-done
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
