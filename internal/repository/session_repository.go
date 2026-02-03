package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/model"
)

type IUserSessionRepository interface {
	Create(ctx context.Context, s *model.UserSession) (*model.UserSession, error)
	Update(ctx context.Context, s *model.UserSession) (*model.UserSession, error)
	Get(ctx context.Context, id uuid.UUID) (*model.UserSession, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.UserSession, int64, error)
	CountActiveByUserID(ctx context.Context, userID string) (int64, error)
	ListActiveByUserID(ctx context.Context, userID string) ([]*model.UserSession, error)
}

type UserSessionRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(db *gorm.DB) *UserSessionRepository {
	return &UserSessionRepository{db: db}
}

func (r *UserSessionRepository) Create(ctx context.Context, s *model.UserSession) (*model.UserSession, error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	err := r.db.WithContext(ctx).Create(s).Error
	return s, err
}

func (r *UserSessionRepository) Update(ctx context.Context, s *model.UserSession) (*model.UserSession, error) {
	err := r.db.WithContext(ctx).Save(s).Error
	return s, err
}

func (r *UserSessionRepository) Get(ctx context.Context, id uuid.UUID) (*model.UserSession, error) {
	var s model.UserSession
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *UserSessionRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.UserSession, int64, error) {
	var list []*model.UserSession
	var count int64
	q := r.db.WithContext(ctx).Model(&model.UserSession{}).Where("user_id = ?", userID)
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	q = q.Order("joined_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	err := q.Find(&list).Error
	return list, count, err
}

func (r *UserSessionRepository) CountActiveByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserSession{}).
		Where("user_id = ? AND left_at IS NULL", userID).Count(&count).Error
	return count, err
}

func (r *UserSessionRepository) ListActiveByUserID(ctx context.Context, userID string) ([]*model.UserSession, error) {
	var list []*model.UserSession
	err := r.db.WithContext(ctx).Where("user_id = ? AND left_at IS NULL", userID).
		Order("joined_at DESC").Find(&list).Error
	return list, err
}
