package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/psds-microservice/user-service/internal/auth"
)

type contextKey string

const ClaimsContextKey contextKey = "claims"

// JWTAuth проверяет Bearer JWT и кладёт claims в контекст.
func JWTAuth(cfg auth.Config, blacklist *auth.Blacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimSpace(parts[1])
			claims, err := cfg.ValidateAccess(tokenString)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}
			if blacklist != nil && blacklist.Contains(claims.ID) {
				http.Error(w, "token revoked", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole возвращает 403 если роль пользователя не в разрешённых.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	set := make(map[string]struct{})
	for _, r := range roles {
		set[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, _ := r.Context().Value(ClaimsContextKey).(*auth.Claims)
			if claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if _, ok := set[claims.Role]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetClaims извлекает claims из контекста.
func GetClaims(ctx context.Context) *auth.Claims {
	c, _ := ctx.Value(ClaimsContextKey).(*auth.Claims)
	return c
}
