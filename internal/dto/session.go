package dto

import "time"

// UserSessionResponse — одна запись сессии.
type UserSessionResponse struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	SessionType          string     `json:"session_type"`
	SessionExternalID    string     `json:"session_external_id"`
	ParticipantRole      string     `json:"participant_role"`
	JoinedAt             time.Time  `json:"joined_at"`
	LeftAt               *time.Time `json:"left_at,omitempty"`
	DurationSeconds      int        `json:"duration_seconds"`
	ConsultationRating   *int       `json:"consultation_rating,omitempty"`
	ConsultationFeedback string     `json:"consultation_feedback,omitempty"`
}

// CreateSessionRequest — POST /api/v1/users/{id}/sessions.
type CreateSessionRequest struct {
	SessionType       string `json:"session_type"` // streaming, consultation, viewing
	SessionExternalID string `json:"session_external_id"`
	ParticipantRole   string `json:"participant_role"` // host, operator, viewer
}
