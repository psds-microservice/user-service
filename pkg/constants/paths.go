package constants

// Базовые пути API
const (
	BasePathAPI = "/api/v1"
	BasePathV1  = BasePathAPI
)

// Health
const (
	PathHealth = "/health"
	PathReady  = "/ready"
)

// Swagger
const (
	PathSwagger = "/swagger"
)

// Auth (относительно BasePathAPI)
const (
	PathAuthRegister = "/auth/register"
	PathAuthLogin    = "/auth/login"
	PathAuthRefresh  = "/auth/refresh"
	PathAuthLogout   = "/auth/logout"
)

// Users (относительно BasePathAPI)
const (
	PathUsers               = "/users"
	PathUsersMe             = "/users/me"
	PathUsersID             = "/users/{id}"
	PathUsersPresence       = "/users/{id}/presence"
	PathUsersSessions       = "/users/{id}/sessions"
	PathUsersActiveSessions = "/users/{id}/active-sessions"
)

// Operators (относительно BasePathAPI)
const (
	PathOperatorsAvailable      = "/operators/available"
	PathOperatorsAvailability   = "/operators/availability"
	PathOperatorsIDAvailability = "/operators/{id}/availability"
	PathOperatorsIDVerify       = "/operators/{id}/verify"
	PathOperatorsStats          = "/operators/stats"
)

// Sessions (относительно BasePathAPI)
const (
	PathSessionsValidate = "/sessions/validate"
)
