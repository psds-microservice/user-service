package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/errs"
	"github.com/psds-microservice/user-service/internal/mapper"
	"github.com/psds-microservice/user-service/internal/model"
)

// AuthService — контракт сервиса аутентификации.
type AuthService interface {
	Login(ctx context.Context, email, password string) (*dto.UserResponse, error)
}

type authService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) AuthService {
	return &authService{db: db}
}

func (s *authService) getByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*dto.UserResponse, error) {
	u, err := s.getByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errs.ErrInvalidCredentials
	}
	if !checkPassword(u.PasswordHash, password) {
		return nil, errs.ErrInvalidCredentials
	}
	return mapper.UserToResponse(u), nil
}
