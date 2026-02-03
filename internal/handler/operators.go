package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/psds-microservice/user-service/internal/middleware"
	"github.com/psds-microservice/user-service/internal/service"
)

// OperatorsHandler — available, availability, verify, stats.
type OperatorsHandler struct {
	svc service.IUserService
}

func NewOperatorsHandler(svc service.IUserService) *OperatorsHandler {
	return &OperatorsHandler{svc: svc}
}

func (h *OperatorsHandler) Available(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	list, count, err := h.svc.ListAvailableOperators(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"operators": list,
		"total":     count,
	})
}

func (h *OperatorsHandler) Availability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		Available bool `json:"available"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.svc.UpdateAvailability(r.Context(), claims.UserID, req.Available)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *OperatorsHandler) Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !claims.IsAdmin() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	operatorID := r.PathValue("id")
	if operatorID == "" {
		http.Error(w, "operator id required", http.StatusBadRequest)
		return
	}
	var req struct {
		Status string `json:"status"` // pending, verified, blocked
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.svc.VerifyOperator(r.Context(), operatorID, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *OperatorsHandler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Список всех операторов с агрегатами (упрощённо — те же available с большим лимитом)
	list, total, err := h.svc.ListAvailableOperators(r.Context(), 100, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Дополнительно можно ListUsers с role=operator и считать — пока отдаём список
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"operators": list,
		"total":     total,
	})
}

func (h *OperatorsHandler) SetAvailabilityByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !claims.IsAdmin() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	operatorID := r.PathValue("id")
	if operatorID == "" {
		http.Error(w, "operator id required", http.StatusBadRequest)
		return
	}
	var req struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.svc.UpdateAvailability(r.Context(), operatorID, req.IsAvailable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
