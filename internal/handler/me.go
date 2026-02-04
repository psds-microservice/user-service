package handler

import (
	"encoding/json"
	"net/http"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/middleware"
	"github.com/psds-microservice/user-service/internal/service"
	"github.com/psds-microservice/user-service/internal/validator"
)

// MeHandler — GET/PUT /api/v1/users/me (текущий пользователь).
type MeHandler struct {
	svc      service.IUserService
	validate *validator.Validator
}

func NewMeHandler(svc service.IUserService, validate *validator.Validator) *MeHandler {
	return &MeHandler{svc: svc, validate: validate}
}

func (h *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.svc.GetUser(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *MeHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.ID = claims.UserID
	if err := h.validate.ValidateUpdateUserRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.svc.UpdateUser(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
