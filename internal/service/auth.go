package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/mapper"
	"github.com/psds-microservice/user-service/internal/model"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*dto.UserResponse, error) {
	var u model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !checkPassword(u.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return mapper.UserToResponse(&u), nil
}
