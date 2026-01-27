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
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, page, limit int, search, status, role string) ([]*models.User, int64, error) {
	args := m.Called(ctx, page, limit, search, status, role)
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) GetUserDevices(ctx context.Context, userID uuid.UUID) ([]*models.Device, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Device), args.Error(1)
}

func (m *MockUserService) UpdateUserDataUsage(ctx context.Context, userID uuid.UUID, dataUsed int64) error {
	args := m.Called(ctx, userID, dataUsed)
	return args.Error(0)
}

// MockLogger is a mock implementation of logger.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Fatal(msg string, fields ...interface{}) {}

type UserHandlerTestSuite struct {
	suite.Suite
	app         *fiber.App
	mockService *MockUserService
	mockLogger  *MockLogger
	userHandler *UserHandler
	testUser    *models.User
	testUserID  uuid.UUID
}

func (suite *UserHandlerTestSuite) SetupTest() {
	suite.mockService = new(MockUserService)
	suite.mockLogger = new(MockLogger)
	suite.userHandler = NewUserHandler(suite.mockService, suite.mockLogger)
	suite.app = fiber.New()
	suite.testUserID = uuid.New()
	suite.testUser = &models.User{
		ID:        suite.testUserID,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FullName:  stringPtr("Test User"),
		Status:    "active",
		Role:      "user",
		DataLimit: 1000000,
		Notes:     stringPtr("Test notes"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Setup routes
	suite.app.Get("/users", suite.userHandler.GetUsers)
	suite.app.Get("/users/:id", suite.userHandler.GetUser)
	suite.app.Post("/users", suite.userHandler.CreateUser)
	suite.app.Put("/users/:id", suite.userHandler.UpdateUser)
	suite.app.Delete("/users/:id", suite.userHandler.DeleteUser)
	suite.app.Get("/users/:userId/devices", suite.userHandler.GetUserDevices)
}

func (suite *UserHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestGetUsers_Success() {
	users := []*models.User{suite.testUser}
	total := int64(1)

	suite.mockService.On("ListUsers", mock.Anything, 1, 10, "", "", "").Return(users, total, nil)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)

	suite.Equal(float64(1), response["total"])
	suite.Equal(float64(1), response["page"])
	suite.Equal(float64(10), response["limit"])
}

func (suite *UserHandlerTestSuite) TestGetUsers_WithPagination() {
	users := []*models.User{suite.testUser}
	total := int64(25)

	suite.mockService.On("ListUsers", mock.Anything, 2, 5, "search", "active", "user").Return(users, total, nil)

	req := httptest.NewRequest("GET", "/users?page=2&limit=5&search=search&status=active&role=user", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestGetUsers_InvalidPage() {
	req := httptest.NewRequest("GET", "/users?page=invalid", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode) // Should default to page 1
}

func (suite *UserHandlerTestSuite) TestGetUsers_LimitTooHigh() {
	req := httptest.NewRequest("GET", "/users?limit=200", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode) // Should cap at 100
}

func (suite *UserHandlerTestSuite) TestGetUsers_ServiceError() {
	suite.mockService.On("ListUsers", mock.Anything, 1, 10, "", "", "").Return([]*models.User{}, int64(0), errors.New("database error"))

	req := httptest.NewRequest("GET", "/users", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusInternalServerError, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Failed to get users", response["error"])
}

func (suite *UserHandlerTestSuite) TestGetUser_Success() {
	suite.mockService.On("GetUserByID", mock.Anything, suite.testUserID).Return(suite.testUser, nil)

	req := httptest.NewRequest("GET", "/users/"+suite.testUserID.String(), nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response models.User
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal(suite.testUser.ID, response.ID)
	suite.Equal(suite.testUser.Username, response.Username)
}

func (suite *UserHandlerTestSuite) TestGetUser_InvalidID() {
	req := httptest.NewRequest("GET", "/users/invalid-uuid", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Invalid user ID", response["error"])
}

func (suite *UserHandlerTestSuite) TestGetUser_NotFound() {
	suite.mockService.On("GetUserByID", mock.Anything, suite.testUserID).Return(nil, errors.New("user not found"))

	req := httptest.NewRequest("GET", "/users/"+suite.testUserID.String(), nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusNotFound, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("User not found", response["error"])
}

func (suite *UserHandlerTestSuite) TestCreateUser_Success() {
	createReq := CreateUserRequest{
		Username:  "newuser",
		Email:     "new@example.com",
		Password:  "password123",
		FullName:  stringPtr("New User"),
		Role:      "user",
		DataLimit: 500000,
		Notes:     stringPtr("New user notes"),
	}

	expectedUser := &models.User{
		ID:        uuid.New(),
		Username:  createReq.Username,
		Email:     createReq.Email,
		Password:  createReq.Password,
		FullName:  createReq.FullName,
		Status:    "active",
		Role:      createReq.Role,
		DataLimit: createReq.DataLimit,
		Notes:     createReq.Notes,
	}

	suite.mockService.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Username == createReq.Username && user.Email == createReq.Email
	})).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*models.User)
		user.ID = expectedUser.ID
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	})

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusCreated, resp.StatusCode)

	var response models.User
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal(expectedUser.Username, response.Username)
	suite.Equal(expectedUser.Email, response.Email)
}

func (suite *UserHandlerTestSuite) TestCreateUser_InvalidBody() {
	req := httptest.NewRequest("POST", "/users", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Invalid request body", response["error"])
}

func (suite *UserHandlerTestSuite) TestCreateUser_ServiceError() {
	createReq := CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	suite.mockService.On("CreateUser", mock.Anything, mock.Anything).Return(errors.New("username already exists"))

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("username already exists", response["error"])
}

func (suite *UserHandlerTestSuite) TestUpdateUser_Success() {
	updateReq := UpdateUserRequest{
		FullName:  stringPtr("Updated Name"),
		DataLimit: int64Ptr(2000000),
		Notes:     stringPtr("Updated notes"),
	}

	updatedUser := *suite.testUser
	updatedUser.FullName = updateReq.FullName
	updatedUser.DataLimit = *updateReq.DataLimit
	updatedUser.Notes = updateReq.Notes

	suite.mockService.On("GetUserByID", mock.Anything, suite.testUserID).Return(suite.testUser, nil)
	suite.mockService.On("UpdateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.ID == suite.testUserID && *user.FullName == "Updated Name"
	})).Return(nil)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/users/"+suite.testUserID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestUpdateUser_InvalidID() {
	updateReq := UpdateUserRequest{
		FullName: stringPtr("Updated Name"),
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/users/invalid-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Invalid user ID", response["error"])
}

func (suite *UserHandlerTestSuite) TestUpdateUser_NotFound() {
	updateReq := UpdateUserRequest{
		FullName: stringPtr("Updated Name"),
	}

	suite.mockService.On("GetUserByID", mock.Anything, suite.testUserID).Return(nil, errors.New("user not found"))

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/users/"+suite.testUserID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusNotFound, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestDeleteUser_Success() {
	suite.mockService.On("DeleteUser", mock.Anything, suite.testUserID).Return(nil)

	req := httptest.NewRequest("DELETE", "/users/"+suite.testUserID.String(), nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusNoContent, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestDeleteUser_InvalidID() {
	req := httptest.NewRequest("DELETE", "/users/invalid-uuid", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestDeleteUser_ServiceError() {
	suite.mockService.On("DeleteUser", mock.Anything, suite.testUserID).Return(errors.New("delete failed"))

	req := httptest.NewRequest("DELETE", "/users/"+suite.testUserID.String(), nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusInternalServerError, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestGetUserDevices_Success() {
	devices := []*models.Device{
		{
			ID:       uuid.New(),
			UserID:   suite.testUserID,
			Name:     "Device 1",
			DeviceID: "device1",
			Status:   "active",
		},
	}

	suite.mockService.On("GetUserDevices", mock.Anything, suite.testUserID).Return(devices, nil)

	req := httptest.NewRequest("GET", "/users/"+suite.testUserID.String()+"/devices", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Len(response["devices"], 1)
}

func (suite *UserHandlerTestSuite) TestGetUserDevices_InvalidUserID() {
	req := httptest.NewRequest("GET", "/users/invalid-uuid/devices", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestGetUserDevices_ServiceError() {
	suite.mockService.On("GetUserDevices", mock.Anything, suite.testUserID).Return([]*models.Device{}, errors.New("service error"))

	req := httptest.NewRequest("GET", "/users/"+suite.testUserID.String()+"/devices", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusInternalServerError, resp.StatusCode)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
