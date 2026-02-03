package mapper

import (
	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/model"
)

// UserToResponse преобразует entity User в DTO UserResponse.
func UserToResponse(u *model.User) *dto.UserResponse {
	if u == nil {
		return nil
	}
	return &dto.UserResponse{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		Phone:          u.Phone,
		Status:         u.Status,
		Role:           u.Role,
		OperatorStatus: u.OperatorStatus,
		MaxSessions:    u.MaxSessions,
		IsAvailable:    u.IsAvailable,
		FullName:       u.FullName,
		AvatarURL:      u.AvatarURL,
		Timezone:       u.Timezone,
		Language:       u.Language,
		Company:        u.Company,
		Specialization: u.Specialization,
		TotalSessions:  u.TotalSessions,
		Rating:         u.Rating,
		IsActive:       u.IsActive,
		IsOnline:       u.IsOnline,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
		LastLogin:      u.LastLogin,
		LastActivity:   u.LastActivity,
		LastSeenAt:     u.LastSeenAt,
	}
}
