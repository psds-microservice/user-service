package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/mapper"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/pkg/constants"
)

// SessionService — сессии пользователя.
type SessionService struct {
	db *gorm.DB
}

// NewSessionService создаёт сервис сессий.
func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) ValidateUserSession(ctx context.Context, userID, sessionExternalID, participantRole string) (allowed bool, err error) {
	_, err = uuid.Parse(userID)
	if err != nil {
		return false, nil
	}
	var user model.User
	err = s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
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

func (s *SessionService) GetUserSessions(ctx context.Context, userID string, limit, offset int) ([]*dto.UserSessionResponse, int64, error) {
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

func (s *SessionService) GetActiveSessions(ctx context.Context, userID string) ([]*dto.UserSessionResponse, error) {
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

func (s *SessionService) CreateSession(ctx context.Context, userID string, req *dto.CreateSessionRequest) (*dto.UserSessionResponse, error) {
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
