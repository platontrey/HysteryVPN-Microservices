package handlers

import (
	"strconv"

	"hysteria2_microservices/api-service/internal/models"
	"hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService interfaces.UserService
	logger      *logger.Logger
}

type CreateUserRequest struct {
	Username  string  `json:"username" validate:"required,min=3,max=50"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=8"`
	FullName  *string `json:"full_name"`
	Role      string  `json:"role" validate:"omitempty,oneof=admin user"`
	DataLimit int64   `json:"data_limit" validate:"min=0"`
	Notes     *string `json:"notes"`
}

type UpdateUserRequest struct {
	Username  *string `json:"username" validate:"omitempty,min=3,max=50"`
	Email     *string `json:"email" validate:"omitempty,email"`
	FullName  *string `json:"full_name"`
	Status    *string `json:"status" validate:"omitempty,oneof=active suspended deleted"`
	Role      *string `json:"role" validate:"omitempty,oneof=admin user"`
	DataLimit *int64  `json:"data_limit" validate:"omitempty,min=0"`
	Notes     *string `json:"notes"`
}

func NewUserHandler(userService interfaces.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
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

	search := c.Query("search")
	status := c.Query("status")
	role := c.Query("role")

	users, total, err := h.userService.ListUsers(c.Context(), page, limit, search, status, role)
	if err != nil {
		h.logger.Error("Failed to get users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get users",
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user, err := h.userService.GetUserByID(c.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err, "user_id", userID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse create user request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password, // Will be hashed in service
		FullName:  req.FullName,
		Role:      req.Role,
		DataLimit: req.DataLimit,
		Status:    "active",
		Notes:     req.Notes,
	}

	if err := h.userService.CreateUser(c.Context(), user); err != nil {
		h.logger.Error("Failed to create user", "error", err, "username", req.Username)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.logger.Info("User created successfully", "user_id", user.ID, "username", user.Username)

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse update user request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := h.userService.GetUserByID(c.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user for update", "error", err, "user_id", userID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Update fields if provided
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FullName != nil {
		user.FullName = req.FullName
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.DataLimit != nil {
		user.DataLimit = *req.DataLimit
	}
	if req.Notes != nil {
		user.Notes = req.Notes
	}

	if err := h.userService.UpdateUser(c.Context(), user); err != nil {
		h.logger.Error("Failed to update user", "error", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	h.logger.Info("User updated successfully", "user_id", userID, "username", user.Username)

	return c.JSON(user)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := h.userService.DeleteUser(c.Context(), userID); err != nil {
		h.logger.Error("Failed to delete user", "error", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	h.logger.Info("User deleted successfully", "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *UserHandler) GetUserDevices(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	devices, err := h.userService.GetUserDevices(c.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user devices", "error", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user devices",
		})
	}

	return c.JSON(fiber.Map{
		"devices": devices,
	})
}
