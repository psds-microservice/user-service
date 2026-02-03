package dto

import "time"

type CreateUserRequest struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Notes     string `json:"notes"`
	CreatedBy int64  `json:"created_by"`
}

type UpdateUserRequest struct {
	Id        uint   `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Password  string `json:"password,omitempty"`
	Notes     string `json:"notes"`
	UpdatedBy int64  `json:"updated_by"`
}

type UserResponse struct {
	Id        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserFilters struct {
	Limit  int
	Offset int
}
