package constants

// Роли пользователей (домен user-service). Источник истины — этот сервис (БД, бизнес-логика).
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
