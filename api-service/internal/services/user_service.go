package services

import (
	"context"
	"fmt"

	"hysteria2_microservices/api-service/internal/models"
	repoInterfaces "hysteria2_microservices/api-service/internal/repositories/interfaces"
	serviceInterfaces "hysteria2_microservices/api-service/internal/services/interfaces"
	"hysteria2_microservices/api-service/pkg/cache"

	"github.com/google/uuid"
)

type userService struct {
	userRepo   repoInterfaces.UserRepository
	deviceRepo repoInterfaces.DeviceRepository
	redis      *cache.RedisClient
}

func NewUserService(userRepo repoInterfaces.UserRepository, deviceRepo repoInterfaces.DeviceRepository, redis *cache.RedisClient) serviceInterfaces.UserService {
	return &userService{
		userRepo:   userRepo,
		deviceRepo: deviceRepo,
		redis:      redis,
	}
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	return s.userRepo.Create(ctx, user)
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", id.String())
	var cachedUser models.User
	if err := s.redis.Get(ctx, cacheKey, &cachedUser); err == nil {
		return &cachedUser, nil
	}

	// Get from database
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache result
	s.redis.Set(ctx, cacheKey, user, 0) // No expiration for user data

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", user.ID.String())
	s.redis.Del(ctx, cacheKey)

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Delete from database
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", id.String())
	s.redis.Del(ctx, cacheKey)

	return nil
}

func (s *userService) ListUsers(ctx context.Context, page, limit int, search, status, role string) ([]*models.User, int64, error) {
	offset := (page - 1) * limit
	return s.userRepo.List(ctx, offset, limit, search, status, role)
}

func (s *userService) GetUserDevices(ctx context.Context, userID uuid.UUID) ([]*models.Device, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user_devices:%s", userID.String())
	var cachedDevices []*models.Device
	if err := s.redis.Get(ctx, cacheKey, &cachedDevices); err == nil {
		return cachedDevices, nil
	}

	// Get from database
	devices, err := s.deviceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache result for 5 minutes
	s.redis.Set(ctx, cacheKey, devices, 300)

	return devices, nil
}

func (s *userService) UpdateUserDataUsage(ctx context.Context, userID uuid.UUID, dataUsed int64) error {
	return s.userRepo.UpdateDataUsage(ctx, userID, dataUsed)
}
