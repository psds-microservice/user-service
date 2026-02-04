// Пути HTTP API: источник — pkg/user_service/user_service.proto (google.api.http).
// Константы CreateUser..UpdateOperatorStatus совпадают с proto. Остальные (Register, Refresh, GetMe, ...)
// — для router до добавления соответствующих RPC в proto. После правок proto: make proto-http-paths.

package user_service

const (
	// BasePathAPI — базовый префикс API (все пути в proto под /api/v1)
	BasePathAPI = "/api/v1"

	// CreateUser
	PathCreateUser   = "/users"
	MethodCreateUser = "POST"

	// GetUser
	PathGetUser   = "/users/{id}"
	MethodGetUser = "GET"

	// UpdateUser
	PathUpdateUser   = "/users/{id}"
	MethodUpdateUser = "PUT"

	// DeleteUser
	PathDeleteUser   = "/users/{id}"
	MethodDeleteUser = "DELETE"

	// Login
	PathLogin   = "/auth/login"
	MethodLogin = "POST"

	// ValidateUserSession
	PathValidateUserSession   = "/sessions/validate"
	MethodValidateUserSession = "POST"

	// UpdateUserPresence
	PathUpdateUserPresence   = "/users/{user_id}/presence"
	MethodUpdateUserPresence = "PUT"

	// GetAvailableOperators
	PathGetAvailableOperators   = "/operators/available"
	MethodGetAvailableOperators = "GET"

	// UpdateOperatorStatus
	PathUpdateOperatorStatus   = "/operators/{user_id}/availability"
	MethodUpdateOperatorStatus = "PUT"

	// Register
	PathRegister   = "/auth/register"
	MethodRegister = "POST"

	// Refresh
	PathRefresh   = "/auth/refresh"
	MethodRefresh = "POST"

	// Logout
	PathLogout   = "/auth/logout"
	MethodLogout = "POST"

	// GetMe
	PathGetMe   = "/users/me"
	MethodGetMe = "GET"

	// UpdateMe
	PathUpdateMe   = "/users/me"
	MethodUpdateMe = "PUT"

	// GetUserSessions
	PathGetUserSessions   = "/users/{id}/sessions"
	MethodGetUserSessions = "GET"

	// GetActiveSessions
	PathGetActiveSessions   = "/users/{id}/active-sessions"
	MethodGetActiveSessions = "GET"

	// CreateSession
	PathCreateSession   = "/users/{id}/sessions"
	MethodCreateSession = "POST"

	// UpdateOperatorAvailability
	PathUpdateOperatorAvailability   = "/operators/availability"
	MethodUpdateOperatorAvailability = "PUT"

	// VerifyOperator
	PathVerifyOperator   = "/operators/{id}/verify"
	MethodVerifyOperator = "POST"

	// GetOperatorStats
	PathGetOperatorStats   = "/operators/stats"
	MethodGetOperatorStats = "GET"
)
