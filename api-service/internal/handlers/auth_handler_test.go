package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"hysteria2_microservices/api-service/internal/models"
	"hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, username, email, password string) (*models.User, error) {
	args := m.Called(ctx, username, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) GenerateTokenPair(userID uuid.UUID) (*interfaces.TokenPair, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.TokenPair), args.Error(1)
}

func (m *MockAuthService) ValidateToken(token string) (*interfaces.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.Claims), args.Error(1)
}

func (m *MockAuthService) RefreshToken(refreshToken string) (*interfaces.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.TokenPair), args.Error(1)
}

func (m *MockAuthService) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type AuthHandlerTestSuite struct {
	suite.Suite
	app           *fiber.App
	mockService   *MockAuthService
	mockLogger    *MockLogger
	authHandler   *AuthHandler
	testUser      *models.User
	testUserID    uuid.UUID
	testTokenPair *interfaces.TokenPair
	testClaims    *interfaces.Claims
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	suite.mockService = new(MockAuthService)
	suite.mockLogger = new(MockLogger)
	suite.authHandler = NewAuthHandler(suite.mockService, suite.mockLogger)
	suite.app = fiber.New()
	suite.testUserID = uuid.New()
	suite.testUser = &models.User{
		ID:       suite.testUserID,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Status:   "active",
	}

	suite.testTokenPair = &interfaces.TokenPair{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		ExpiresIn:    3600,
	}

	suite.testClaims = &interfaces.Claims{
		UserID:   suite.testUserID.String(),
		Username: "testuser",
		Role:     "user",
	}

	// Setup routes
	suite.app.Post("/auth/register", suite.authHandler.Register)
	suite.app.Post("/auth/login", suite.authHandler.Login)
	suite.app.Post("/auth/refresh", suite.authHandler.RefreshToken)
}

func (suite *AuthHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestRegister_Success() {
	registerReq := RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	newUser := &models.User{
		ID:       uuid.New(),
		Username: registerReq.Username,
		Email:    registerReq.Email,
		Role:     "user",
		Status:   "active",
	}

	suite.mockService.On("Register", mock.Anything, registerReq.Username, registerReq.Email, registerReq.Password).Return(newUser, nil)
	suite.mockService.On("GenerateTokenPair", newUser.ID).Return(suite.testTokenPair, nil)

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)

	userData := response["user"].(map[string]interface{})
	suite.Equal(newUser.ID.String(), userData["id"])
	suite.Equal(newUser.Username, userData["username"])
	suite.Equal(newUser.Email, userData["email"])
	suite.Equal(newUser.Role, userData["role"])
	suite.Equal(newUser.Status, userData["status"])

	tokenData := response["token"].(map[string]interface{})
	suite.Equal(suite.testTokenPair.AccessToken, tokenData["access_token"])
	suite.Equal(suite.testTokenPair.RefreshToken, tokenData["refresh_token"])
	suite.Equal(float64(suite.testTokenPair.ExpiresIn), tokenData["expires_in"])
}

func (suite *AuthHandlerTestSuite) TestRegister_InvalidBody() {
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Invalid request body", response["error"])
}

func (suite *AuthHandlerTestSuite) TestRegister_ServiceError() {
	registerReq := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	suite.mockService.On("Register", mock.Anything, registerReq.Username, registerReq.Email, registerReq.Password).Return(nil, errors.New("email already exists"))

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("email already exists", response["error"])
}

func (suite *AuthHandlerTestSuite) TestRegister_TokenGenerationError() {
	registerReq := RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	newUser := &models.User{
		ID:       uuid.New(),
		Username: registerReq.Username,
		Email:    registerReq.Email,
		Role:     "user",
		Status:   "active",
	}

	suite.mockService.On("Register", mock.Anything, registerReq.Username, registerReq.Email, registerReq.Password).Return(newUser, nil)
	suite.mockService.On("GenerateTokenPair", newUser.ID).Return(nil, errors.New("token generation failed"))

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusInternalServerError, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Failed to generate tokens", response["error"])
}

func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	suite.mockService.On("Login", mock.Anything, loginReq.Email, loginReq.Password).Return(suite.testUser, nil)
	suite.mockService.On("GenerateTokenPair", suite.testUser.ID).Return(suite.testTokenPair, nil)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)

	userData := response["user"].(map[string]interface{})
	suite.Equal(suite.testUser.ID.String(), userData["id"])
	suite.Equal(suite.testUser.Username, userData["username"])
	suite.Equal(suite.testUser.Email, userData["email"])

	tokenData := response["token"].(map[string]interface{})
	suite.Equal(suite.testTokenPair.AccessToken, tokenData["access_token"])
}

func (suite *AuthHandlerTestSuite) TestLogin_InvalidCredentials() {
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	suite.mockService.On("Login", mock.Anything, loginReq.Email, loginReq.Password).Return(nil, errors.New("invalid credentials"))

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Invalid credentials", response["error"])
}

func (suite *AuthHandlerTestSuite) TestLogin_TokenGenerationError() {
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	suite.mockService.On("Login", mock.Anything, loginReq.Email, loginReq.Password).Return(suite.testUser, nil)
	suite.mockService.On("GenerateTokenPair", suite.testUser.ID).Return(nil, errors.New("token generation failed"))

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusInternalServerError, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Failed to generate tokens", response["error"])
}

func (suite *AuthHandlerTestSuite) TestRefreshToken_Success() {
	refreshReq := RefreshRequest{
		RefreshToken: "refresh_token_456",
	}

	suite.mockService.On("RefreshToken", refreshReq.RefreshToken).Return(suite.testTokenPair, nil)

	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)

	tokenData := response["token"].(map[string]interface{})
	suite.Equal(suite.testTokenPair.AccessToken, tokenData["access_token"])
	suite.Equal(suite.testTokenPair.RefreshToken, tokenData["refresh_token"])
}

func (suite *AuthHandlerTestSuite) TestRefreshToken_InvalidToken() {
	refreshReq := RefreshRequest{
		RefreshToken: "invalid_token",
	}

	suite.mockService.On("RefreshToken", refreshReq.RefreshToken).Return(nil, errors.New("invalid refresh token"))

	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Invalid refresh token", response["error"])
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
