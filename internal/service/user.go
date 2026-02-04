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

// UserService — CRUD пользователей.
type UserService struct {
	db *gorm.DB
}

// NewUserService создаёт сервис пользователей.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	if req.Username != "" {
		var existing model.User
		err := s.db.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error
		if err == nil {
			return nil, ErrUserAlreadyExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	var existing model.User
	err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existing).Error
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	username := req.Username
	if username == "" {
		username = req.Email
	}

	role := req.Role
	if role == "" {
		role = constants.RoleClient
	}
	if role != constants.RoleClient && role != constants.RoleOperator && role != constants.RoleAdmin {
		role = constants.RoleClient
	}
	operatorStatus := ""
	if role == constants.RoleOperator {
		operatorStatus = constants.OperatorStatusPending
	}

	user := &model.User{
		ID:             uuid.New().String(),
		Username:       username,
		Email:          req.Email,
		Phone:          req.Phone,
		PasswordHash:   hashedPassword,
		Status:         constants.UserStatusActive,
		Role:           role,
		OperatorStatus: operatorStatus,
		MaxSessions:    1,
		IsAvailable:    false,
		IsActive:       true,
		Language:       "en",
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(user), nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	var user model.User
	err = s.db.WithContext(ctx).Where("id = ?", id.String()).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	user.Phone = req.Phone
	if req.Status != "" {
		user.Status = req.Status
	}
	if req.Password != "" {
		hashed, err := hashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashed
	}
	user.FullName = req.FullName
	user.AvatarURL = req.AvatarURL
	user.Timezone = req.Timezone
	user.Language = req.Language
	user.Company = req.Company
	user.Specialization = req.Specialization

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(&user), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidUserID
	}
	return s.db.WithContext(ctx).Delete(&model.User{}, "id = ?", uid.String()).Error
}

func (s *UserService) GetUser(ctx context.Context, id string) (*dto.UserResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	var user model.User
	err = s.db.WithContext(ctx).Where("id = ?", uid.String()).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return mapper.UserToResponse(&user), nil
}

func (s *UserService) ListUsers(ctx context.Context, filters *dto.UserFilters) ([]*dto.UserResponse, int64, error) {
	var users []*model.User
	var count int64
	query := s.db.WithContext(ctx).Model(&model.User{})

	if filters != nil {
		if filters.Status != "" {
			query = query.Where("status = ?", filters.Status)
		}
		if filters.Role != "" {
			query = query.Where("role = ?", filters.Role)
		}
		if filters.Search != "" {
			search := "%" + filters.Search + "%"
			query = query.Where("username ILIKE ? OR email ILIKE ?", search, search)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if filters != nil {
		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var responses []*dto.UserResponse
	for _, u := range users {
		responses = append(responses, mapper.UserToResponse(u))
	}
	return responses, count, nil
}
