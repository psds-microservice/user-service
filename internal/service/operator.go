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

// OperatorService — контракт сервиса операторов.
type OperatorService interface {
	ListAvailableOperators(ctx context.Context, limit, offset int) ([]*dto.UserResponse, int64, error)
	UpdateAvailability(ctx context.Context, userID string, available bool) (*dto.UserResponse, error)
	VerifyOperator(ctx context.Context, operatorID string, status string) (*dto.UserResponse, error)
}

type operatorService struct {
	db *gorm.DB
}

func NewOperatorService(db *gorm.DB) OperatorService {
	return &operatorService{db: db}
}

func (s *operatorService) getByID(ctx context.Context, id string) (*model.User, error) {
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

func (s *operatorService) ListAvailableOperators(ctx context.Context, limit, offset int) ([]*dto.UserResponse, int64, error) {
	var list []*model.User
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
	if err := query.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*dto.UserResponse, len(list))
	for i := range list {
		out[i] = mapper.UserToResponse(list[i])
	}
	return out, count, nil
}

func (s *operatorService) UpdateAvailability(ctx context.Context, userID string, available bool) (*dto.UserResponse, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserID
	}
	user, err := s.getByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	user.IsAvailable = available
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(user), nil
}

func (s *operatorService) VerifyOperator(ctx context.Context, operatorID string, status string) (*dto.UserResponse, error) {
	if status != constants.OperatorStatusPending && status != constants.OperatorStatusVerified && status != constants.OperatorStatusBlocked {
		return nil, ErrInvalidOperatorStatus
	}
	if _, err := uuid.Parse(operatorID); err != nil {
		return nil, ErrInvalidUserID
	}
	user, err := s.getByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.Role != constants.RoleOperator {
		return nil, ErrNotOperator
	}
	user.OperatorStatus = status
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(user), nil
}
