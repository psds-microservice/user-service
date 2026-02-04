package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/model"
)

type PresenceService struct {
	db *gorm.DB
}

func NewPresenceService(db *gorm.DB) *PresenceService {
	return &PresenceService{db: db}
}

func (s *PresenceService) UpdatePresence(ctx context.Context, userID string, isOnline bool) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidUserID
	}
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", uid.String()).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	now := time.Now()
	user.IsOnline = isOnline
	user.LastSeenAt = &now
	return s.db.WithContext(ctx).Save(&user).Error
}
