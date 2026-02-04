package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/mapper"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/internal/repository"
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
	repo        repository.IUserRepository
	sessionRepo repository.IUserSessionRepository
}

func NewUserService(repo repository.IUserRepository, sessionRepo repository.IUserSessionRepository) *UserService {
	return &UserService{repo: repo, sessionRepo: sessionRepo}
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	if req.Username != "" {
		existing, _ := s.repo.FindByUsername(ctx, req.Username)
		if existing != nil {
			return nil, ErrUserAlreadyExists
		}
	}
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	username := req.Username
	if username == "" {
		username = req.Email // fallback to email if no username
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

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return mapper.UserToResponse(createdUser), nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
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

	updatedUser, err := s.repo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return mapper.UserToResponse(updatedUser), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidUserID
	}
	return s.repo.Delete(ctx, uid)
}

func (s *UserService) GetUser(ctx context.Context, id string) (*dto.UserResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	user, err := s.repo.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return mapper.UserToResponse(user), nil
}

func (s *UserService) ListUsers(ctx context.Context, filters *dto.UserFilters) ([]*dto.UserResponse, int64, error) {
	users, count, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	var responses []*dto.UserResponse
	for _, u := range users {
		responses = append(responses, mapper.UserToResponse(u))
	}

	return responses, count, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !checkPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	return mapper.UserToResponse(user), nil
}

func (s *UserService) ListAvailableOperators(ctx context.Context, limit, offset int) ([]*dto.UserResponse, int64, error) {
	users, count, err := s.repo.ListAvailableOperators(ctx, limit, offset)
	if err != nil {
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
	user, err := s.repo.Get(ctx, uid)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	if user.Role != constants.RoleOperator {
		return nil, ErrNotOperator
	}
	user.IsAvailable = available
	updated, err := s.repo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return mapper.UserToResponse(updated), nil
}

func (s *UserService) VerifyOperator(ctx context.Context, operatorID string, status string) (*dto.UserResponse, error) {
	if status != constants.OperatorStatusPending && status != constants.OperatorStatusVerified && status != constants.OperatorStatusBlocked {
		return nil, ErrInvalidOperatorStatus
	}
	uid, err := uuid.Parse(operatorID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	user, err := s.repo.Get(ctx, uid)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	if user.Role != constants.RoleOperator {
		return nil, ErrNotOperator
	}
	user.OperatorStatus = status
	updated, err := s.repo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return mapper.UserToResponse(updated), nil
}

func (s *UserService) UpdatePresence(ctx context.Context, userID string, isOnline bool) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidUserID
	}
	user, err := s.repo.Get(ctx, uid)
	if err != nil || user == nil {
		return ErrUserNotFound
	}
	now := time.Now()
	user.IsOnline = isOnline
	user.LastSeenAt = &now
	_, err = s.repo.Update(ctx, user)
	return err
}

func (s *UserService) ValidateUserSession(ctx context.Context, userID, sessionExternalID, participantRole string) (allowed bool, err error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, nil
	}
	user, err := s.repo.Get(ctx, uid)
	if err != nil || user == nil || !user.IsActive {
		return false, nil
	}
	existing, err := s.sessionRepo.FindActiveByUserAndExternalID(ctx, userID, sessionExternalID)
	if err != nil {
		return false, err
	}
	if existing != nil {
		return true, nil
	}
	activeCount, err := s.sessionRepo.CountActiveByUserID(ctx, userID)
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
	list, count, err := s.sessionRepo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*dto.UserSessionResponse, len(list))
	for i := range list {
		out[i] = mapper.SessionToResponse(list[i])
	}
	return out, count, nil
}

func (s *UserService) GetActiveSessions(ctx context.Context, userID string) ([]*dto.UserSessionResponse, error) {
	list, err := s.sessionRepo.ListActiveByUserID(ctx, userID)
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
	user, err := s.repo.Get(ctx, uid)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	activeCount, err := s.sessionRepo.CountActiveByUserID(ctx, userID)
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
		_, _ = s.repo.Update(ctx, user)
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
	created, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, err
	}
	user.TotalSessions++
	user.IsOnline = true
	_, _ = s.repo.Update(ctx, user)
	return mapper.SessionToResponse(created), nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
