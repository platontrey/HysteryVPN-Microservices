package handlers

import (
	"fmt"
	"time"

	"hysteria2_microservices/api-service/internal/models"
	serviceInterfaces "hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/gofiber/fiber/v2"
	ws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// WebSocket message types
type WSMessageType string

const (
	WSTrafficUpdate WSMessageType = "traffic_update"
	WSUserStatus    WSMessageType = "user_status"
	WSDeviceOnline  WSMessageType = "device_online"
	WSError         WSMessageType = "error"
)

type WSMessage struct {
	Type      WSMessageType `json:"type"`
	UserID    string        `json:"user_id,omitempty"`
	Data      interface{}   `json:"data,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

type WebSocketHandler struct {
	trafficService serviceInterfaces.TrafficService
	logger         *logger.Logger
	clients        map[string]*ws.Conn
}

func NewWebSocketHandler(trafficService serviceInterfaces.TrafficService, logger *logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		trafficService: trafficService,
		logger:         logger,
		clients:        make(map[string]*ws.Conn),
	}
}

// SetTrafficService sets the traffic service (for breaking circular dependency)
func (h *WebSocketHandler) SetTrafficService(trafficService serviceInterfaces.TrafficService) {
	h.trafficService = trafficService
}

// WebSocket upgrade middleware
func (h *WebSocketHandler) WebSocketUpgrade() func(*fiber.Ctx) error {
	return ws.New(func(c *ws.Conn) {
		// Get user ID from context (set by JWT middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			h.logger.Warn("WebSocket connection without user ID")
			c.Close()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			h.logger.Warn("Invalid user ID in WebSocket connection")
			c.Close()
			return
		}

		// Register client
		clientKey := fmt.Sprintf("user_%s", userIDStr)
		h.clients[clientKey] = c
		h.logger.Info("WebSocket client connected", "user_id", userIDStr)

		// Clean up on disconnect
		defer func() {
			delete(h.clients, clientKey)
			h.logger.Info("WebSocket client disconnected", "user_id", userIDStr)
		}()

		// Send welcome message
		welcomeMsg := WSMessage{
			Type:      WSUserStatus,
			UserID:    userIDStr,
			Data:      map[string]string{"status": "connected"},
			Timestamp: time.Now(),
		}
		if err := h.sendMessage(c, welcomeMsg); err != nil {
			h.logger.Error("Failed to send welcome message", "error", err)
			return
		}

		// Handle incoming messages
		for {
			var msg WSMessage
			err := c.ReadJSON(&msg)
			if err != nil {
				if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
					h.logger.Error("WebSocket error", "error", err)
				}
				break
			}

			// Handle client messages (ping, subscribe, etc.)
			h.handleClientMessage(c, userIDStr, msg)
		}
	})
}

func (h *WebSocketHandler) handleClientMessage(c *ws.Conn, userID string, msg WSMessage) {
	switch msg.Type {
	case "ping":
		// Respond to ping
		pongMsg := WSMessage{
			Type:      "pong",
			UserID:    userID,
			Data:      map[string]string{"timestamp": time.Now().Format(time.RFC3339)},
			Timestamp: time.Now(),
		}
		h.sendMessage(c, pongMsg)

	case "subscribe_traffic":
		// Client wants to subscribe to traffic updates
		h.logger.Info("Client subscribed to traffic updates", "user_id", userID)

	default:
		h.logger.Warn("Unknown WebSocket message type", "type", msg.Type, "user_id", userID)
	}
}

func (h *WebSocketHandler) sendMessage(c *ws.Conn, msg WSMessage) error {
	return c.WriteJSON(msg)
}

// Broadcast traffic update to specific user
func (h *WebSocketHandler) BroadcastTrafficUpdate(userID uuid.UUID, stats *models.TrafficStats) {
	clientKey := fmt.Sprintf("user_%s", userID.String())
	if conn, exists := h.clients[clientKey]; exists {
		msg := WSMessage{
			Type:      WSTrafficUpdate,
			UserID:    userID.String(),
			Data:      stats,
			Timestamp: time.Now(),
		}

		if err := h.sendMessage(conn, msg); err != nil {
			h.logger.Error("Failed to send traffic update", "error", err, "user_id", userID)
			// Remove dead connection
			delete(h.clients, clientKey)
		} else {
			h.logger.Debug("Traffic update sent", "user_id", userID, "upload", stats.Upload, "download", stats.Download)
		}
	}
}

// Broadcast user status update
func (h *WebSocketHandler) BroadcastUserStatus(userID uuid.UUID, status string) {
	clientKey := fmt.Sprintf("user_%s", userID.String())
	if conn, exists := h.clients[clientKey]; exists {
		msg := WSMessage{
			Type:      WSUserStatus,
			UserID:    userID.String(),
			Data:      map[string]string{"status": status},
			Timestamp: time.Now(),
		}

		if err := h.sendMessage(conn, msg); err != nil {
			h.logger.Error("Failed to send user status", "error", err, "user_id", userID)
			delete(h.clients, clientKey)
		}
	}
}

// Broadcast device online/offline status
func (h *WebSocketHandler) BroadcastDeviceStatus(deviceID uuid.UUID, userID uuid.UUID, online bool) {
	clientKey := fmt.Sprintf("user_%s", userID.String())
	if conn, exists := h.clients[clientKey]; exists {
		status := "offline"
		if online {
			status = "online"
		}

		msg := WSMessage{
			Type:      WSDeviceOnline,
			UserID:    userID.String(),
			Data:      map[string]interface{}{"device_id": deviceID.String(), "status": status},
			Timestamp: time.Now(),
		}

		if err := h.sendMessage(conn, msg); err != nil {
			h.logger.Error("Failed to send device status", "error", err, "device_id", deviceID)
			delete(h.clients, clientKey)
		}
	}
}

// Get connected clients count
func (h *WebSocketHandler) GetConnectedClientsCount() int {
	return len(h.clients)
}

// Get connected clients for specific user
func (h *WebSocketHandler) IsUserConnected(userID uuid.UUID) bool {
	clientKey := fmt.Sprintf("user_%s", userID.String())
	_, exists := h.clients[clientKey]
	return exists
}

// Ensure WebSocketHandler implements WebSocketService interface
var _ serviceInterfaces.WebSocketService = (*WebSocketHandler)(nil)
