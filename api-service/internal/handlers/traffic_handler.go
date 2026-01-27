package handlers

import (
	"time"

	"hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TrafficHandler struct {
	trafficService interfaces.TrafficService
	logger         *logger.Logger
}

func NewTrafficHandler(trafficService interfaces.TrafficService, logger *logger.Logger) *TrafficHandler {
	return &TrafficHandler{
		trafficService: trafficService,
		logger:         logger,
	}
}

func (h *TrafficHandler) GetUserTraffic(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Parse time range
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "from and to parameters are required",
		})
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid from time format (use RFC3339)",
		})
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid to time format (use RFC3339)",
		})
	}

	traffic, err := h.trafficService.GetUserTraffic(c.Context(), userID, from, to)
	if err != nil {
		h.logger.Error("Failed to get user traffic", "error", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user traffic",
		})
	}

	return c.JSON(fiber.Map{
		"traffic": traffic,
		"user_id": userID,
		"from":    from,
		"to":      to,
	})
}

func (h *TrafficHandler) GetTrafficSummary(c *fiber.Ctx) error {
	// Default to last 30 days
	now := time.Now()
	from := now.AddDate(0, 0, -30)
	to := now

	// Parse optional parameters
	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsed
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsed
		}
	}

	summary, err := h.trafficService.GetTrafficSummary(c.Context(), from, to)
	if err != nil {
		h.logger.Error("Failed to get traffic summary", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get traffic summary",
		})
	}

	return c.JSON(summary)
}
