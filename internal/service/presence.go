package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/errs"
	"github.com/psds-microservice/user-service/internal/model"
)

// PresenceService — контракт сервиса presence (онлайн-статус).
type PresenceService interface {
	UpdatePresence(ctx context.Context, userID string, isOnline bool) error
}

type presenceService struct {
	db *gorm.DB
}

func NewPresenceService(db *gorm.DB) PresenceService {
	return &presenceService{db: db}
}

func (s *presenceService) getByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *presenceService) UpdatePresence(ctx context.Context, userID string, isOnline bool) error {
	if _, err := uuid.Parse(userID); err != nil {
		return errs.ErrInvalidUserID
	}
	user, err := s.getByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errs.ErrUserNotFound
	}
	now := time.Now()
	user.IsOnline = isOnline
	user.LastSeenAt = &now
	return s.db.WithContext(ctx).Save(user).Error
}
