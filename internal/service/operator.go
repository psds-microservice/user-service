package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/mapper"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/pkg/constants"
)

type OperatorService struct {
	db *gorm.DB
}

func NewOperatorService(db *gorm.DB) *OperatorService {
	return &OperatorService{db: db}
}

func (s *OperatorService) ListAvailableOperators(ctx context.Context, limit, offset int) ([]*dto.UserResponse, int64, error) {
	var users []*model.User
	var count int64
	query := s.db.WithContext(ctx).Model(&model.User{}).
		Where("role = ? AND operator_status = ? AND is_available = ? AND is_active = ?",
			constants.RoleOperator, constants.OperatorStatusVerified, true, true)
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	query = query.Order("rating DESC NULLS LAST")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}
	var out []*dto.UserResponse
	for _, u := range users {
		out = append(out, mapper.UserToResponse(u))
	}
	return out, count, nil
}

func (s *OperatorService) UpdateAvailability(ctx context.Context, userID string, available bool) (*dto.UserResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", uid.String()).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	user.IsAvailable = available
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(&user), nil
}

func (s *OperatorService) VerifyOperator(ctx context.Context, operatorID string, status string) (*dto.UserResponse, error) {
	if status != constants.OperatorStatusPending && status != constants.OperatorStatusVerified && status != constants.OperatorStatusBlocked {
		return nil, ErrInvalidOperatorStatus
	}
	uid, err := uuid.Parse(operatorID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", uid.String()).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if user.Role != constants.RoleOperator {
		return nil, ErrNotOperator
	}
	user.OperatorStatus = status
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(&user), nil
}
