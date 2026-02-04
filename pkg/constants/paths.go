package constants

// Пути, не описанные в user_service.proto: health, ready, swagger.
// Все пути API ( /api/v1/... ) берутся из proto: pkg/gen/user_service/http_paths.go.
const (
	PathHealth  = "/health"
	PathReady   = "/ready"
	PathSwagger = "/swagger"
)
