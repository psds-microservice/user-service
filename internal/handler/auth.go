package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/psds-microservice/user-service/internal/auth"
	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/service"
)

// AuthHandler — register, login, refresh, logout.
type AuthHandler struct {
	svc       service.IUserService
	jwtConfig auth.Config
	blacklist *auth.Blacklist
}

func NewAuthHandler(svc service.IUserService, jwtConfig auth.Config, blacklist *auth.Blacklist) *AuthHandler {
	return &AuthHandler{svc: svc, jwtConfig: jwtConfig, blacklist: blacklist}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	createReq := &dto.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}
	user, err := h.svc.CreateUser(r.Context(), createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	access, refresh, err := h.jwtConfig.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
	if err != nil {
		http.Error(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}
	sendTokenResponse(w, access, refresh, user, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	access, refresh, err := h.jwtConfig.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
	if err != nil {
		http.Error(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}
	sendTokenResponse(w, access, refresh, user, http.StatusOK)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	userID, err := h.jwtConfig.ValidateRefresh(req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}
	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	access, refresh, err := h.jwtConfig.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
	if err != nil {
		http.Error(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}
	sendTokenResponse(w, access, refresh, user, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Инвалидация по Bearer токену (добавляем jti в blacklist через middleware context или тело запроса)
	// Упрощённо: клиент при logout присылает access_token в заголовке — мы его инвалидируем
	header := r.Header.Get("Authorization")
	if header != "" && len(header) > 7 && header[:7] == "Bearer " {
		tokenString := header[7:]
		claims, err := h.jwtConfig.ValidateAccess(tokenString)
		if err == nil {
			exp := time.Now().Add(24 * time.Hour)
			if claims.ExpiresAt != nil {
				exp = claims.ExpiresAt.Time
			}
			h.blacklist.Add(claims.ID, exp)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func sendTokenResponse(w http.ResponseWriter, access, refresh string, user *dto.UserResponse, status int) {
	// ExpiresIn в секундах (типично 900 для 15m)
	expiresIn := 900
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
		User:         user,
	})
}
