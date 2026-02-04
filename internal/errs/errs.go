package errs

import "errors"

// Доменные сентинель-ошибки для маппинга в gRPC/HTTP коды.
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
