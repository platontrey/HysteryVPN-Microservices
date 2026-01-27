package repositories

import (
	"context"

	"hysteria2_microservices/api-service/internal/models"
	repoInterfaces "hysteria2_microservices/api-service/internal/repositories/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) repoInterfaces.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("session_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	var sessions []*models.Session
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error
	return sessions, err
}

func (r *sessionRepository) Update(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id).Error
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < NOW()").Delete(&models.Session{}).Error
}

func (r *sessionRepository) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Update("is_active", false).Error
}
