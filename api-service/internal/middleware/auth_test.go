package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"hysteria2_microservices/api-service/internal/models"
	"hysteria2_microservices/api-service/internal/services/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockAuthServiceForMiddleware is a mock implementation of AuthService for middleware testing
type MockAuthServiceForMiddleware struct {
	mock.Mock
}

func (m *MockAuthServiceForMiddleware) Register(ctx context.Context, username, email, password string) (*models.User, error) {
	args := m.Called(ctx, username, email, password)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthServiceForMiddleware) Login(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthServiceForMiddleware) GenerateTokenPair(userID uuid.UUID) (*interfaces.TokenPair, error) {
	args := m.Called(userID)
	return args.Get(0).(*interfaces.TokenPair), args.Error(1)
}

func (m *MockAuthServiceForMiddleware) ValidateToken(token string) (*interfaces.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.Claims), args.Error(1)
}

func (m *MockAuthServiceForMiddleware) RefreshToken(refreshToken string) (*interfaces.TokenPair, error) {
	args := m.Called(refreshToken)
	return args.Get(0).(*interfaces.TokenPair), args.Error(1)
}

func (m *MockAuthServiceForMiddleware) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockLoggerForMiddleware is a mock implementation of logger.Logger for middleware testing
type MockLoggerForMiddleware struct {
	mock.Mock
}

func (m *MockLoggerForMiddleware) Debug(msg string, fields ...interface{}) {}
func (m *MockLoggerForMiddleware) Info(msg string, fields ...interface{})  {}
func (m *MockLoggerForMiddleware) Warn(msg string, fields ...interface{})  {}
func (m *MockLoggerForMiddleware) Error(msg string, fields ...interface{}) {}

type MiddlewareTestSuite struct {
	suite.Suite
	app         *fiber.App
	mockService *MockAuthServiceForMiddleware
	mockLogger  *MockLoggerForMiddleware
	testClaims  *interfaces.Claims
}

func (suite *MiddlewareTestSuite) SetupTest() {
	suite.mockService = new(MockAuthServiceForMiddleware)
	suite.mockLogger = new(MockLoggerForMiddleware)
	suite.app = fiber.New()

	suite.testClaims = &interfaces.Claims{
		UserID:   uuid.New().String(),
		Username: "testuser",
		Role:     "user",
	}

	// Setup a test route with middleware
	suite.app.Use(JWTAuth(suite.mockService))
	suite.app.Use(Logging(suite.mockLogger))
	suite.app.Get("/protected", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		username := c.Locals("username")
		role := c.Locals("role")
		return c.JSON(fiber.Map{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})
}

func (suite *MiddlewareTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

func (suite *MiddlewareTestSuite) TestJWTAuth_ValidToken() {
	validToken := "valid.jwt.token"

	suite.mockService.On("ValidateToken", validToken).Return(suite.testClaims, nil)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal(suite.testClaims.UserID, response["user_id"])
	suite.Equal(suite.testClaims.Username, response["username"])
	suite.Equal(suite.testClaims.Role, response["role"])
}

func (suite *MiddlewareTestSuite) TestJWTAuth_MissingAuthorizationHeader() {
	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Authorization header required", response["error"])
}

func (suite *MiddlewareTestSuite) TestJWTAuth_InvalidHeaderFormat() {
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Invalid authorization header format", response["error"])
}

func (suite *MiddlewareTestSuite) TestJWTAuth_InvalidToken() {
	invalidToken := "invalid.jwt.token"

	suite.mockService.On("ValidateToken", invalidToken).Return(nil, errors.New("invalid token"))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)
	resp, err := suite.app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Invalid or expired token", response["error"])
}

func (suite *MiddlewareTestSuite) TestRequireRole_SufficientRole() {
	app := fiber.New()

	// Override the auth middleware to set role
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Use(RequireRole("user"))
	app.Get("/admin-only", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/admin-only", nil)
	resp, err := app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)
}

func (suite *MiddlewareTestSuite) TestRequireRole_InsufficientRole() {
	app := fiber.New()

	// Override the auth middleware to set role
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("role", "user")
		return c.Next()
	})

	app.Use(RequireRole("admin"))
	app.Get("/admin-only", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/admin-only", nil)
	resp, err := app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusForbidden, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Insufficient permissions", response["error"])
}

func (suite *MiddlewareTestSuite) TestRequireRole_MissingRole() {
	app := fiber.New()

	app.Use(RequireRole("admin"))
	app.Get("/admin-only", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/admin-only", nil)
	resp, err := app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusForbidden, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("User role not found", response["error"])
}

func (suite *MiddlewareTestSuite) TestLogging_RequestLogging() {
	// Create a simple app for logging test
	app := fiber.New()
	app.Use(Logging(suite.mockLogger))

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("logged")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent")
	req.RemoteAddr = "127.0.0.1:12345"

	resp, err := app.Test(req)

	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
