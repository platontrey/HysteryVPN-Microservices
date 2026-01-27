package handlers

import (
	"strconv"

	"hysteria2_microservices/api-service/internal/models"
	serviceInterfaces "hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type XrayHandler struct {
	xrayService serviceInterfaces.XrayService
	logger      *logger.Logger
}

func NewXrayHandler(xrayService serviceInterfaces.XrayService, logger *logger.Logger) *XrayHandler {
	return &XrayHandler{
		xrayService: xrayService,
		logger:      logger,
	}
}

// GenerateXrayConfig generates a new Xray configuration for a user
func (h *XrayHandler) GenerateXrayConfig(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	protocol := c.Query("protocol", "vless")
	deviceID := c.Query("deviceId", "")

	h.logger.Infof("Generating Xray config for user %s, protocol %s", userID, protocol)

	config, err := h.xrayService.GenerateUserConfig(c.Context(), userID, deviceID, protocol)
	if err != nil {
		h.logger.Errorf("Failed to generate Xray config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate configuration",
		})
	}

	return c.JSON(fiber.Map{
		"config": config,
	})
}

// UpdateXrayConfig updates an existing Xray configuration
func (h *XrayHandler) UpdateXrayConfig(c *fiber.Ctx) error {
	configID := c.Params("configId")
	if configID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Config ID is required",
		})
	}

	var config models.XrayConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	configIDParsed, err := uuid.Parse(configID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid config ID",
		})
	}
	config.ID = configIDParsed

	userID := c.Params("userId")
	deviceID := c.Query("deviceId", "")

	h.logger.Infof("Updating Xray config %s for user %s", configID, userID)

	if err := h.xrayService.UpdateUserConfig(c.Context(), userID, deviceID, &config); err != nil {
		h.logger.Errorf("Failed to update Xray config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update configuration",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Configuration updated successfully",
	})
}

// GetXrayConfigs gets all Xray configurations for a user
func (h *XrayHandler) GetXrayConfigs(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// This would need to be implemented in the service layer
	// For now, return empty array
	h.logger.Infof("Getting Xray configs for user %s", userID)

	return c.JSON(fiber.Map{
		"configs": []interface{}{},
	})
}

// GetSupportedProtocols returns the list of supported Xray protocols
func (h *XrayHandler) GetSupportedProtocols(c *fiber.Ctx) error {
	h.logger.Info("Getting supported Xray protocols")

	protocols, err := h.xrayService.GetSupportedProtocols(c.Context())
	if err != nil {
		h.logger.Errorf("Failed to get supported protocols: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get supported protocols",
		})
	}

	return c.JSON(fiber.Map{
		"protocols": protocols,
	})
}

// GetXrayStatus returns the status of Xray services
func (h *XrayHandler) GetXrayStatus(c *fiber.Ctx) error {
	h.logger.Info("Getting Xray service status")

	status, err := h.xrayService.GetServiceStatus(c.Context())
	if err != nil {
		h.logger.Errorf("Failed to get Xray status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get service status",
		})
	}

	return c.JSON(fiber.Map{
		"status": status,
	})
}

// GetXrayConnections returns active Xray connections
func (h *XrayHandler) GetXrayConnections(c *fiber.Ctx) error {
	h.logger.Info("Getting active Xray connections")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	connections, err := h.xrayService.GetActiveConnections(c.Context())
	if err != nil {
		h.logger.Errorf("Failed to get active connections: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get active connections",
		})
	}

	// Simple pagination (in real implementation, this would be more sophisticated)
	start := (page - 1) * limit
	end := start + limit
	if start > len(connections) {
		start = len(connections)
	}
	if end > len(connections) {
		end = len(connections)
	}

	paginatedConnections := connections[start:end]

	return c.JSON(fiber.Map{
		"connections": paginatedConnections,
		"total":       len(connections),
		"page":        page,
		"limit":       limit,
	})
}

// DisconnectXrayUser disconnects a specific user from Xray
func (h *XrayHandler) DisconnectXrayUser(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	deviceID := c.Query("deviceId", "")

	h.logger.Infof("Disconnecting Xray user %s, device %s", userID, deviceID)

	if err := h.xrayService.DisconnectUser(c.Context(), userID, deviceID); err != nil {
		h.logger.Errorf("Failed to disconnect user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to disconnect user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User disconnected successfully",
	})
}

// ReloadXrayConfiguration reloads Xray configuration across all nodes
func (h *XrayHandler) ReloadXrayConfiguration(c *fiber.Ctx) error {
	h.logger.Info("Reloading Xray configuration")

	if err := h.xrayService.ReloadConfiguration(c.Context()); err != nil {
		h.logger.Errorf("Failed to reload configuration: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reload configuration",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Configuration reloaded successfully",
	})
}
