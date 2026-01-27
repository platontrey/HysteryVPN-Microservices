package handlers

import (
	"strconv"

	"hysteria2_microservices/api-service/internal/models"
	"hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type NodeHandler struct {
	nodeService interfaces.NodeService
	logger      *logger.Logger
}

type CreateNodeRequest struct {
	Name         string            `json:"name" validate:"required,min=2,max=100"`
	Hostname     string            `json:"hostname" validate:"required,min=2,max=255"`
	IPAddress    string            `json:"ip_address" validate:"required,ip"`
	Location     string            `json:"location" validate:"omitempty,max=100"`
	Country      string            `json:"country" validate:"omitempty,len=2"`
	GRPCPort     int               `json:"grpc_port" validate:"omitempty,min=1,max=65535"`
	Capabilities map[string]string `json:"capabilities"`
	Metadata     map[string]string `json:"metadata"`
}

type UpdateNodeRequest struct {
	Name         *string            `json:"name" validate:"omitempty,min=2,max=100"`
	Hostname     *string            `json:"hostname" validate:"omitempty,min=2,max=255"`
	IPAddress    *string            `json:"ip_address" validate:"omitempty,ip"`
	Location     *string            `json:"location" validate:"omitempty,max=100"`
	Country      *string            `json:"country" validate:"omitempty,len=2"`
	GRPCPort     *int               `json:"grpc_port" validate:"omitempty,min=1,max=65535"`
	Status       *string            `json:"status" validate:"omitempty,oneof=online offline maintenance error"`
	Capabilities *map[string]string `json:"capabilities"`
	Metadata     *map[string]string `json:"metadata"`
}

func NewNodeHandler(nodeService interfaces.NodeService, logger *logger.Logger) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
		logger:      logger,
	}
}

func (h *NodeHandler) GetNodes(c *fiber.Ctx) error {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	statusFilter := c.Query("status")
	locationFilter := c.Query("location")

	nodes, total, err := h.nodeService.ListNodes(c.Context(), page, limit, statusFilter, locationFilter)
	if err != nil {
		h.logger.Error("Failed to get nodes", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get nodes",
		})
	}

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *NodeHandler) GetNode(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}

	node, err := h.nodeService.GetNodeByID(c.Context(), nodeID)
	if err != nil {
		h.logger.Error("Failed to get node", "error", err, "node_id", nodeID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	return c.JSON(node)
}

func (h *NodeHandler) CreateNode(c *fiber.Ctx) error {
	var req CreateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse create node request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set default values
	if req.GRPCPort == 0 {
		req.GRPCPort = 50051
	}

	// Convert string maps to interface{} maps
	capabilities := make(map[string]interface{})
	for k, v := range req.Capabilities {
		capabilities[k] = v
	}
	metadata := make(map[string]interface{})
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	node := &models.VPSNode{
		Name:         req.Name,
		Hostname:     req.Hostname,
		IPAddress:    req.IPAddress,
		Location:     req.Location,
		Country:      req.Country,
		GRPCPort:     req.GRPCPort,
		Status:       "offline", // New nodes start as offline
		Capabilities: capabilities,
		Metadata:     metadata,
	}

	if err := h.nodeService.CreateNode(c.Context(), node); err != nil {
		h.logger.Error("Failed to create node", "error", err, "name", req.Name)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.logger.Info("Node created successfully", "node_id", node.ID, "name", node.Name)

	return c.Status(fiber.StatusCreated).JSON(node)
}

func (h *NodeHandler) UpdateNode(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}

	var req UpdateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse update node request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	node, err := h.nodeService.GetNodeByID(c.Context(), nodeID)
	if err != nil {
		h.logger.Error("Failed to get node for update", "error", err, "node_id", nodeID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	// Update fields if provided
	if req.Name != nil {
		node.Name = *req.Name
	}
	if req.Hostname != nil {
		node.Hostname = *req.Hostname
	}
	if req.IPAddress != nil {
		node.IPAddress = *req.IPAddress
	}
	if req.Location != nil {
		node.Location = *req.Location
	}
	if req.Country != nil {
		node.Country = *req.Country
	}
	if req.GRPCPort != nil {
		node.GRPCPort = *req.GRPCPort
	}
	if req.Status != nil {
		node.Status = *req.Status
	}
	if req.Capabilities != nil {
		capabilities := make(map[string]interface{})
		for k, v := range *req.Capabilities {
			capabilities[k] = v
		}
		node.Capabilities = capabilities
	}
	if req.Metadata != nil {
		metadata := make(map[string]interface{})
		for k, v := range *req.Metadata {
			metadata[k] = v
		}
		node.Metadata = metadata
	}

	if err := h.nodeService.UpdateNode(c.Context(), node); err != nil {
		h.logger.Error("Failed to update node", "error", err, "node_id", nodeID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update node",
		})
	}

	h.logger.Info("Node updated successfully", "node_id", nodeID, "name", node.Name)

	return c.JSON(node)
}

func (h *NodeHandler) DeleteNode(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}

	if err := h.nodeService.DeleteNode(c.Context(), nodeID); err != nil {
		h.logger.Error("Failed to delete node", "error", err, "node_id", nodeID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete node",
		})
	}

	h.logger.Info("Node deleted successfully", "node_id", nodeID)

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *NodeHandler) GetNodeMetrics(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}

	limit := 100 // Default limit
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	metrics, err := h.nodeService.GetNodeMetrics(c.Context(), nodeID, limit)
	if err != nil {
		h.logger.Error("Failed to get node metrics", "error", err, "node_id", nodeID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get node metrics",
		})
	}

	return c.JSON(fiber.Map{
		"metrics": metrics,
		"limit":   limit,
	})
}

func (h *NodeHandler) RestartNode(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}

	if err := h.nodeService.RestartNode(c.Context(), nodeID); err != nil {
		h.logger.Error("Failed to restart node", "error", err, "node_id", nodeID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restart node",
		})
	}

	h.logger.Info("Node restarted successfully", "node_id", nodeID)

	return c.JSON(fiber.Map{
		"message": "Node restart initiated",
	})
}

func (h *NodeHandler) GetNodeLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}

	lines := 100 // Default
	if l := c.Query("lines"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			lines = parsed
		}
	}

	logs, err := h.nodeService.GetNodeLogs(c.Context(), nodeID, lines)
	if err != nil {
		h.logger.Error("Failed to get node logs", "error", err, "node_id", nodeID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get node logs",
		})
	}

	return c.JSON(fiber.Map{
		"logs":  logs,
		"lines": lines,
	})
}
