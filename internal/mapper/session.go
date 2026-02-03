package mapper

import (
	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/model"
)

// SessionToResponse преобразует entity UserSession в DTO UserSessionResponse.
func SessionToResponse(s *model.UserSession) *dto.UserSessionResponse {
	if s == nil {
		return nil
	}
	return &dto.UserSessionResponse{
		ID:                   s.ID,
		UserID:               s.UserID,
		SessionType:          s.SessionType,
		SessionExternalID:    s.SessionExternalID,
		ParticipantRole:      s.ParticipantRole,
		JoinedAt:             s.JoinedAt,
		LeftAt:               s.LeftAt,
		DurationSeconds:      s.DurationSeconds,
		ConsultationRating:   s.ConsultationRating,
		ConsultationFeedback: s.ConsultationFeedback,
	}
}
