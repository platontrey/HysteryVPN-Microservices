package middleware

import (
	"hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTAuth(authService interfaces.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
				"code":  "AUTH_HEADER_REQUIRED",
			})
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
				"code":  "INVALID_AUTH_FORMAT",
			})
		}

		token := tokenParts[1]

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func Logging(logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Log request
		logger.Info("HTTP Request",
			"method", c.Method(),
			"path", c.Path(),
			"ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
		)

		// Process request
		err := c.Next()

		// Log response
		status := c.Response().StatusCode()
		if status >= 400 {
			logger.Warn("HTTP Response",
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
				"ip", c.IP(),
			)
		} else {
			logger.Debug("HTTP Response",
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
			)
		}

		return err
	}
}

func RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		if userRole == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "User role not found",
				"code":  "ROLE_NOT_FOUND",
			})
		}

		role, ok := userRole.(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid user role",
				"code":  "INVALID_ROLE",
			})
		}

		if role != "admin" && role != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
		}

		return c.Next()
	}
}
