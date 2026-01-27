package handlers

import (
	"github.com/gofiber/fiber/v2"

	"hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"
)

type AuthHandler struct {
	authService interfaces.AuthService
	logger      *logger.Logger
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func NewAuthHandler(authService interfaces.AuthService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse register request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	user, err := h.authService.Register(c.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to register user", "error", err, "username", req.Username, "email", req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  "REGISTRATION_FAILED",
		})
	}

	// Generate token pair
	tokenPair, err := h.authService.GenerateTokenPair(user.ID)
	if err != nil {
		h.logger.Error("Failed to generate token pair", "error", err, "user_id", user.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
			"code":  "TOKEN_GENERATION_FAILED",
		})
	}

	h.logger.Info("User registered successfully", "user_id", user.ID, "username", user.Username)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"status":   user.Status,
		},
		"token": tokenPair,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse login request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Failed login attempt", "email", req.Email, "error", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Generate token pair
	tokenPair, err := h.authService.GenerateTokenPair(user.ID)
	if err != nil {
		h.logger.Error("Failed to generate token pair", "error", err, "user_id", user.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	h.logger.Info("User logged in successfully", "user_id", user.ID, "username", user.Username)

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"status":   user.Status,
		},
		"token": tokenPair,
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse refresh request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Failed token refresh", "error", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	h.logger.Info("Token refreshed successfully")

	return c.JSON(fiber.Map{
		"token": tokenPair,
	})
}
