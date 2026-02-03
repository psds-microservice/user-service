package constants

// Роли пользователей
const (
	RoleClient   = "client"
	RoleOperator = "operator"
	RoleAdmin    = "admin"
)

// Статусы оператора (operator_status)
const (
	OperatorStatusPending  = "pending"
	OperatorStatusVerified = "verified"
	OperatorStatusBlocked  = "blocked"
)

// Статусы пользователя (status)
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBlocked  = "blocked"
)
