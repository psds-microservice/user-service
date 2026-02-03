package dto

import "time"

// CreateUserRequest — запрос на создание пользователя.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"` // client, operator, admin
}

// UpdateUserRequest — запрос на обновление пользователя.
type UpdateUserRequest struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Password       string `json:"password,omitempty"`
	Status         string `json:"status"`
	FullName       string `json:"full_name"`
	AvatarURL      string `json:"avatar_url"`
	Timezone       string `json:"timezone"`
	Language       string `json:"language"`
	Company        string `json:"company"`
	Specialization string `json:"specialization"`
}

// UserResponse — ответ с данными пользователя.
type UserResponse struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
	Status         string     `json:"status"`
	Role           string     `json:"role"`
	OperatorStatus string     `json:"operator_status,omitempty"`
	MaxSessions    int        `json:"max_sessions,omitempty"`
	IsAvailable    bool       `json:"is_available,omitempty"`
	FullName       string     `json:"full_name,omitempty"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	Timezone       string     `json:"timezone,omitempty"`
	Language       string     `json:"language,omitempty"`
	Company        string     `json:"company,omitempty"`
	Specialization string     `json:"specialization,omitempty"`
	TotalSessions  int        `json:"total_sessions,omitempty"`
	Rating         float64    `json:"rating,omitempty"`
	IsActive       bool       `json:"is_active,omitempty"`
	IsOnline       bool       `json:"is_online,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastLogin      *time.Time `json:"last_login,omitempty"`
	LastActivity   *time.Time `json:"last_activity,omitempty"`
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"`
}

// UserFilters — фильтры для списка пользователей.
type UserFilters struct {
	Limit  int
	Offset int
	Status string
	Role   string
	Search string
}
