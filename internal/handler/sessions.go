package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/middleware"
	"github.com/psds-microservice/user-service/internal/service"
	"github.com/psds-microservice/user-service/internal/validator"
)

// SessionsHandler — GET/POST /api/v1/users/{id}/sessions, GET .../active-sessions.
type SessionsHandler struct {
	svc      service.IUserService
	validate *validator.Validator
}

func NewSessionsHandler(svc service.IUserService, validate *validator.Validator) *SessionsHandler {
	return &SessionsHandler{svc: svc, validate: validate}
}

func (h *SessionsHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	// Доступ: свой профиль или admin
	if claims.UserID != userID && !claims.IsAdmin() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	list, count, err := h.svc.GetUserSessions(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": list,
		"total":    count,
	})
}

func (h *SessionsHandler) ListActiveSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.UserID != userID && !claims.IsAdmin() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	list, err := h.svc.GetActiveSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"sessions": list})
}

func (h *SessionsHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.UserID != userID && !claims.IsAdmin() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req dto.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.ValidateCreateSessionRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	session, err := h.svc.CreateSession(r.Context(), userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (h *SessionsHandler) ValidateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		UserID            string `json:"user_id"`
		SessionExternalID string `json:"session_external_id"`
		ParticipantRole   string `json:"participant_role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.ValidateSessionValidateRequest(req.UserID, req.SessionExternalID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	allowed, err := h.svc.ValidateUserSession(r.Context(), req.UserID, req.SessionExternalID, req.ParticipantRole)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"allowed": false, "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"allowed": allowed})
}
