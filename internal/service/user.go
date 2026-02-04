package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/errs"
	"github.com/psds-microservice/user-service/internal/mapper"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/pkg/constants"
)

// UserService — контракт сервиса пользователей (CRUD).
type UserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUser(ctx context.Context, id string) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filters *dto.UserFilters) ([]*dto.UserResponse, int64, error)
}

type userService struct {
	db *gorm.DB
}

// NewUserService создаёт сервис пользователей.
func NewUserService(db *gorm.DB) UserService {
	return &userService{db: db}
}

func (s *userService) getByID(ctx context.Context, id string) (*model.User, error) {
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

func (s *userService) getByEmail(ctx context.Context, email string) (*model.User, error) {
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

func (s *userService) getByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	if req.Username != "" {
		existing, err := s.getByUsername(ctx, req.Username)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errs.ErrUserAlreadyExists
		}
	}
	existing, err := s.getByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errs.ErrUserAlreadyExists
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

func (s *userService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	if _, err := uuid.Parse(req.ID); err != nil {
		return nil, errs.ErrInvalidUserID
	}
	user, err := s.getByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errs.ErrUserNotFound
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

	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(user), nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errs.ErrInvalidUserID
	}
	return s.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

func (s *userService) GetUser(ctx context.Context, id string) (*dto.UserResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errs.ErrInvalidUserID
	}
	user, err := s.getByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errs.ErrUserNotFound
	}
	return mapper.UserToResponse(user), nil
}

func (s *userService) ListUsers(ctx context.Context, filters *dto.UserFilters) ([]*dto.UserResponse, int64, error) {
	var list []*model.User
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
	if err := query.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*dto.UserResponse, len(list))
	for i := range list {
		out[i] = mapper.UserToResponse(list[i])
	}
	return out, count, nil
}
