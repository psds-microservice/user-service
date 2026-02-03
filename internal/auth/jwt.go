package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Permissions по ролям (promt.txt).
var permissionsByRole = map[string][]string{
	"client":   {"stream:create", "stream:join", "chat:send", "file:upload"},
	"operator": {"stream:join", "chat:send", "file:upload", "consultation:join"},
	"admin":    {"stream:create", "stream:join", "chat:send", "file:upload", "consultation:join", "operator:verify", "operator:stats"},
}

// Claims — JWT claims (promt.txt).
type Claims struct {
	jwt.RegisteredClaims
	UserID         string   `json:"user_id"`
	Email          string   `json:"email"`
	Role           string   `json:"role"`
	OperatorStatus string   `json:"operator_status,omitempty"`
	IsAvailable    bool     `json:"is_available"`
	Permissions    []string `json:"permissions"`
}

// RefreshClaims — только sub (user_id) и exp для refresh токена.
type RefreshClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// Config для генерации/проверки JWT.
type Config struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// NewConfig создаёт конфиг из строк (secret, accessTTL, refreshTTL).
func NewConfig(secret, accessTTL, refreshTTL string) (Config, error) {
	at, err := time.ParseDuration(accessTTL)
	if err != nil {
		at = 15 * time.Minute
	}
	rt, err := time.ParseDuration(refreshTTL)
	if err != nil {
		rt = 168 * time.Hour
	}
	return Config{
		Secret:     []byte(secret),
		AccessTTL:  at,
		RefreshTTL: rt,
	}, nil
}

// GeneratePair выдаёт access и refresh токены.
func (c Config) GeneratePair(userID, email, role, operatorStatus string, isAvailable bool) (access, refresh string, err error) {
	now := time.Now()
	perms := permissionsByRole[role]
	if perms == nil {
		perms = permissionsByRole["client"]
	}

	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(c.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("at-%s-%d", userID, now.UnixNano()),
		},
		UserID:         userID,
		Email:          email,
		Role:           role,
		OperatorStatus: operatorStatus,
		IsAvailable:    isAvailable,
		Permissions:    perms,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	access, err = accessToken.SignedString(c.Secret)
	if err != nil {
		return "", "", err
	}

	refreshClaims := &RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(c.RefreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("rt-%s-%d", userID, now.UnixNano()),
		},
		UserID: userID,
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refresh, err = refreshToken.SignedString(c.Secret)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// ValidateAccess проверяет access токен и возвращает claims.
func (c Config) ValidateAccess(tokenString string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return c.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ValidateRefresh проверяет refresh токен и возвращает userID.
func (c Config) ValidateRefresh(tokenString string) (userID string, err error) {
	tok, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return c.Secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := tok.Claims.(*RefreshClaims)
	if !ok || !tok.Valid {
		return "", errors.New("invalid refresh token")
	}
	return claims.UserID, nil
}

// HasPermission проверяет наличие разрешения в claims.
func (c *Claims) HasPermission(perm string) bool {
	for _, p := range c.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// IsAdmin возвращает true если роль admin.
func (c *Claims) IsAdmin() bool { return c.Role == "admin" }
