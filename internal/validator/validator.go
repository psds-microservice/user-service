package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/pkg/constants"
)

const (
	minPasswordLength = 6
	maxUsernameLength = 128
	maxEmailLength    = 256
)

var (
	emailRegex = regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)
)

// Validator — валидация входящих DTO перед вызовом сервиса.
type Validator struct{}

func New() *Validator {
	return &Validator{}
}

// ValidateRegisterRequest проверяет RegisterRequest (POST /api/v1/auth/register).
func (v *Validator) ValidateRegisterRequest(req *dto.RegisterRequest) error {
	var errs []string
	if strings.TrimSpace(req.Email) == "" {
		errs = append(errs, "email is required")
	} else if len(req.Email) > maxEmailLength {
		errs = append(errs, "email too long")
	} else if !emailRegex.MatchString(req.Email) {
		errs = append(errs, "email format is invalid")
	}
	if strings.TrimSpace(req.Password) == "" {
		errs = append(errs, "password is required")
	} else if len(req.Password) < minPasswordLength {
		errs = append(errs, fmt.Sprintf("password must be at least %d characters", minPasswordLength))
	}
	if req.Role != "" && req.Role != constants.RoleClient && req.Role != constants.RoleOperator && req.Role != constants.RoleAdmin {
		errs = append(errs, "role must be one of: client, operator, admin")
	}
	if len(req.Username) > maxUsernameLength {
		errs = append(errs, "username too long")
	}
	if len(errs) > 0 {
		return errors.New("validation: " + strings.Join(errs, "; "))
	}
	return nil
}

// ValidateLoginRequest проверяет LoginRequest (POST /api/v1/auth/login).
func (v *Validator) ValidateLoginRequest(req *dto.LoginRequest) error {
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("validation: email is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("validation: password is required")
	}
	return nil
}

// ValidateRefreshRequest проверяет RefreshRequest (POST /api/v1/auth/refresh).
func (v *Validator) ValidateRefreshRequest(req *dto.RefreshRequest) error {
	if strings.TrimSpace(req.RefreshToken) == "" {
		return errors.New("validation: refresh_token is required")
	}
	return nil
}

// ValidateCreateUserRequest проверяет CreateUserRequest.
func (v *Validator) ValidateCreateUserRequest(req *dto.CreateUserRequest) error {
	var errs []string
	if strings.TrimSpace(req.Email) == "" {
		errs = append(errs, "email is required")
	} else if len(req.Email) > maxEmailLength {
		errs = append(errs, "email too long")
	} else if !emailRegex.MatchString(req.Email) {
		errs = append(errs, "email format is invalid")
	}
	if strings.TrimSpace(req.Password) == "" {
		errs = append(errs, "password is required")
	} else if len(req.Password) < minPasswordLength {
		errs = append(errs, fmt.Sprintf("password must be at least %d characters", minPasswordLength))
	}
	if req.Role != "" && req.Role != constants.RoleClient && req.Role != constants.RoleOperator && req.Role != constants.RoleAdmin {
		errs = append(errs, "role must be one of: client, operator, admin")
	}
	if len(req.Username) > maxUsernameLength {
		errs = append(errs, "username too long")
	}
	if len(errs) > 0 {
		return errors.New("validation: " + strings.Join(errs, "; "))
	}
	return nil
}

// ValidateUpdateUserRequest проверяет UpdateUserRequest (ID и опциональные поля).
func (v *Validator) ValidateUpdateUserRequest(req *dto.UpdateUserRequest) error {
	if strings.TrimSpace(req.ID) == "" {
		return errors.New("validation: id is required")
	}
	if _, err := uuid.Parse(req.ID); err != nil {
		return errors.New("validation: id must be a valid UUID")
	}
	if req.Email != "" && !emailRegex.MatchString(req.Email) {
		return errors.New("validation: email format is invalid")
	}
	if req.Status != "" && req.Status != constants.UserStatusActive && req.Status != constants.UserStatusInactive && req.Status != constants.UserStatusBlocked {
		return errors.New("validation: status must be one of: active, inactive, blocked")
	}
	if len(req.Username) > maxUsernameLength {
		return errors.New("validation: username too long")
	}
	if req.Password != "" && len(req.Password) < minPasswordLength {
		return fmt.Errorf("validation: password must be at least %d characters", minPasswordLength)
	}
	return nil
}

// ValidateCreateSessionRequest проверяет CreateSessionRequest.
func (v *Validator) ValidateCreateSessionRequest(req *dto.CreateSessionRequest) error {
	allowedTypes := map[string]bool{"streaming": true, "consultation": true, "viewing": true}
	if strings.TrimSpace(req.SessionType) == "" {
		return errors.New("validation: session_type is required")
	}
	if !allowedTypes[req.SessionType] {
		return errors.New("validation: session_type must be one of: streaming, consultation, viewing")
	}
	if strings.TrimSpace(req.SessionExternalID) == "" {
		return errors.New("validation: session_external_id is required")
	}
	allowedRoles := map[string]bool{"host": true, "operator": true, "viewer": true}
	if req.ParticipantRole != "" && !allowedRoles[req.ParticipantRole] {
		return errors.New("validation: participant_role must be one of: host, operator, viewer")
	}
	return nil
}

// ValidateSessionValidateRequest проверяет запрос на валидацию сессии (user_id, session_external_id).
func (v *Validator) ValidateSessionValidateRequest(userID, sessionExternalID string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("validation: user_id is required")
	}
	if strings.TrimSpace(sessionExternalID) == "" {
		return errors.New("validation: session_external_id is required")
	}
	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("validation: user_id must be a valid UUID")
	}
	return nil
}
