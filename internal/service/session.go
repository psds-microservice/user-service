package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/errs"
	"github.com/psds-microservice/user-service/internal/mapper"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/pkg/constants"
)

// SessionService — контракт сервиса сессий.
type SessionService interface {
	GetUserSessions(ctx context.Context, userID string, limit, offset int) ([]*dto.UserSessionResponse, int64, error)
	GetActiveSessions(ctx context.Context, userID string) ([]*dto.UserSessionResponse, error)
	CreateSession(ctx context.Context, userID string, req *dto.CreateSessionRequest) (*dto.UserSessionResponse, error)
	ValidateUserSession(ctx context.Context, userID, sessionExternalID, participantRole string) (bool, error)
}

type sessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) SessionService {
	return &sessionService{db: db}
}

func (s *sessionService) getUserByID(ctx context.Context, userID string) (*model.User, error) {
	var u model.User
	err := s.db.WithContext(ctx).Where("id = ?", userID).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *sessionService) countActiveByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.UserSession{}).Where("user_id = ? AND left_at IS NULL", userID).Count(&count).Error
	return count, err
}

func (s *sessionService) findActiveByUserAndExternalID(ctx context.Context, userID, sessionExternalID string) (*model.UserSession, error) {
	var sess model.UserSession
	err := s.db.WithContext(ctx).Where("user_id = ? AND session_external_id = ? AND left_at IS NULL", userID, sessionExternalID).First(&sess).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sess, nil
}

func (s *sessionService) ValidateUserSession(ctx context.Context, userID, sessionExternalID, participantRole string) (allowed bool, err error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, nil
	}
	user, err := s.getUserByID(ctx, userID)
	if err != nil || user == nil || !user.IsActive {
		return false, nil
	}
	existing, err := s.findActiveByUserAndExternalID(ctx, userID, sessionExternalID)
	if err != nil {
		return false, err
	}
	if existing != nil {
		return true, nil
	}
	activeCount, err := s.countActiveByUser(ctx, userID)
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

func (s *sessionService) GetUserSessions(ctx context.Context, userID string, limit, offset int) ([]*dto.UserSessionResponse, int64, error) {
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

func (s *sessionService) GetActiveSessions(ctx context.Context, userID string) ([]*dto.UserSessionResponse, error) {
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

func (s *sessionService) CreateSession(ctx context.Context, userID string, req *dto.CreateSessionRequest) (*dto.UserSessionResponse, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, errs.ErrInvalidUserID
	}
	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errs.ErrUserNotFound
	}
	activeCount, err := s.countActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Role == constants.RoleClient && req.SessionType == "streaming" && activeCount >= 1 {
		return nil, errs.ErrClientStreamingLimit
	}
	if user.Role == constants.RoleOperator && (user.OperatorStatus != constants.OperatorStatusVerified || !user.IsAvailable) {
		return nil, ErrOperatorNotVerifiedOrAvailable
	}
	if int(activeCount) >= user.MaxSessions {
		user.IsAvailable = false
		_ = s.db.WithContext(ctx).Save(user)
		return nil, errs.ErrMaxSessionsReached
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
	_ = s.db.WithContext(ctx).Save(user)
	return mapper.SessionToResponse(session), nil
}
