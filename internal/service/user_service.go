package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/mapper"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/pkg/constants"
)

// Сентинель-ошибки домена user-service для единообразного маппинга в транспортные коды (gRPC/HTTP).
var (
	ErrUserAlreadyExists              = errors.New("user already exists")
	ErrInvalidUserID                  = errors.New("invalid user id")
	ErrUserNotFound                   = errors.New("user not found")
	ErrInvalidCredentials             = errors.New("invalid credentials")
	ErrNotOperator                    = errors.New("user is not an operator")
	ErrInvalidOperatorStatus          = errors.New("invalid operator status")
	ErrClientStreamingLimit           = errors.New("client may have only one active streaming session")
	ErrOperatorNotVerifiedOrAvailable = errors.New("operator must be verified and available")
	ErrMaxSessionsReached             = errors.New("max_sessions reached")
)

type IUserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	GetUser(ctx context.Context, id string) (*dto.UserResponse, error)
	ListUsers(ctx context.Context, filters *dto.UserFilters) ([]*dto.UserResponse, int64, error)
	Login(ctx context.Context, email, password string) (*dto.UserResponse, error)
	ListAvailableOperators(ctx context.Context, limit, offset int) ([]*dto.UserResponse, int64, error)
	UpdateAvailability(ctx context.Context, userID string, available bool) (*dto.UserResponse, error)
	VerifyOperator(ctx context.Context, operatorID string, status string) (*dto.UserResponse, error)
	UpdatePresence(ctx context.Context, userID string, isOnline bool) error
	ValidateUserSession(ctx context.Context, userID, sessionExternalID, participantRole string) (allowed bool, err error)
	GetUserSessions(ctx context.Context, userID string, limit, offset int) ([]*dto.UserSessionResponse, int64, error)
	GetActiveSessions(ctx context.Context, userID string) ([]*dto.UserSessionResponse, error)
	CreateSession(ctx context.Context, userID string, req *dto.CreateSessionRequest) (*dto.UserSessionResponse, error)
}

type UserService struct {
	db *gorm.DB
}

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

func (s *UserService) Login(ctx context.Context, email, password string) (*dto.UserResponse, error) {
	var user model.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !checkPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return mapper.UserToResponse(&user), nil
}

func (s *UserService) ListAvailableOperators(ctx context.Context, limit, offset int) ([]*dto.UserResponse, int64, error) {
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

func (s *UserService) UpdateAvailability(ctx context.Context, userID string, available bool) (*dto.UserResponse, error) {
	uid, err := uuid.Parse(userID)
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
	user.IsAvailable = available
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(&user), nil
}

func (s *UserService) VerifyOperator(ctx context.Context, operatorID string, status string) (*dto.UserResponse, error) {
	if status != constants.OperatorStatusPending && status != constants.OperatorStatusVerified && status != constants.OperatorStatusBlocked {
		return nil, ErrInvalidOperatorStatus
	}
	uid, err := uuid.Parse(operatorID)
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
	if user.Role != constants.RoleOperator {
		return nil, ErrNotOperator
	}
	user.OperatorStatus = status
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return mapper.UserToResponse(&user), nil
}

func (s *UserService) UpdatePresence(ctx context.Context, userID string, isOnline bool) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidUserID
	}
	var user model.User
	err = s.db.WithContext(ctx).Where("id = ?", uid.String()).First(&user).Error
	if err != nil {
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

func (s *UserService) ValidateUserSession(ctx context.Context, userID, sessionExternalID, participantRole string) (allowed bool, err error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, nil
	}
	var user model.User
	err = s.db.WithContext(ctx).Where("id = ?", uid.String()).First(&user).Error
	if err != nil || !user.IsActive {
		return false, nil
	}
	var existing model.UserSession
	err = s.db.WithContext(ctx).Where("user_id = ? AND session_external_id = ? AND left_at IS NULL", userID, sessionExternalID).First(&existing).Error
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	var activeCount int64
	err = s.db.WithContext(ctx).Model(&model.UserSession{}).Where("user_id = ? AND left_at IS NULL", userID).Count(&activeCount).Error
	if err != nil {
		return false, err
	}
	if user.Role == constants.RoleOperator && (user.OperatorStatus != constants.OperatorStatusVerified || !user.IsAvailable) {
		return false, nil
	}
	if int(activeCount) >= user.MaxSessions {
		return false, nil
	}
	return true, nil
}

func (s *UserService) GetUserSessions(ctx context.Context, userID string, limit, offset int) ([]*dto.UserSessionResponse, int64, error) {
	var list []*model.UserSession
	var count int64
	q := s.db.WithContext(ctx).Model(&model.UserSession{}).Where("user_id = ?", userID)
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
	if err := q.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*dto.UserSessionResponse, len(list))
	for i := range list {
		out[i] = mapper.SessionToResponse(list[i])
	}
	return out, count, nil
}

func (s *UserService) GetActiveSessions(ctx context.Context, userID string) ([]*dto.UserSessionResponse, error) {
	var list []*model.UserSession
	err := s.db.WithContext(ctx).Where("user_id = ? AND left_at IS NULL", userID).Order("joined_at DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]*dto.UserSessionResponse, len(list))
	for i := range list {
		out[i] = mapper.SessionToResponse(list[i])
	}
	return out, nil
}

func (s *UserService) CreateSession(ctx context.Context, userID string, req *dto.CreateSessionRequest) (*dto.UserSessionResponse, error) {
	uid, err := uuid.Parse(userID)
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
	var activeCount int64
	err = s.db.WithContext(ctx).Model(&model.UserSession{}).Where("user_id = ? AND left_at IS NULL", userID).Count(&activeCount).Error
	if err != nil {
		return nil, err
	}
	if user.Role == constants.RoleClient && req.SessionType == "streaming" && activeCount >= 1 {
		return nil, ErrClientStreamingLimit
	}
	if user.Role == constants.RoleOperator && (user.OperatorStatus != constants.OperatorStatusVerified || !user.IsAvailable) {
		return nil, ErrOperatorNotVerifiedOrAvailable
	}
	if int(activeCount) >= user.MaxSessions {
		user.IsAvailable = false
		_ = s.db.WithContext(ctx).Save(&user).Error
		return nil, ErrMaxSessionsReached
	}
	session := &model.UserSession{
		ID:                uuid.New().String(),
		UserID:            userID,
		SessionType:       req.SessionType,
		SessionExternalID: req.SessionExternalID,
		ParticipantRole:   req.ParticipantRole,
		JoinedAt:          time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, err
	}
	user.TotalSessions++
	user.IsOnline = true
	_ = s.db.WithContext(ctx).Save(&user).Error
	return mapper.SessionToResponse(session), nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
