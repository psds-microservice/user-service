package handler

import (
	"net/http"

	"github.com/psds-microservice/user-service/internal/auth"
	"github.com/psds-microservice/user-service/internal/middleware"
	"github.com/psds-microservice/user-service/internal/service"
)

// APIv1 возвращает http.Handler для префикса /api/v1.
func APIv1(
	userSvc service.IUserService,
	jwtConfig auth.Config,
	blacklist *auth.Blacklist,
	userHandler *UserHandler,
) http.Handler {
	authHandler := NewAuthHandler(userSvc, jwtConfig, blacklist)
	meHandler := NewMeHandler(userSvc)
	operatorsHandler := NewOperatorsHandler(userSvc)
	sessionsHandler := NewSessionsHandler(userSvc)

	mux := http.NewServeMux()

	// Публичные
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	// Требуют JWT
	withAuth := middleware.JWTAuth(jwtConfig, blacklist)
	mux.Handle("GET /users/me", withAuth(http.HandlerFunc(meHandler.GetMe)))
	mux.Handle("PUT /users/me", withAuth(http.HandlerFunc(meHandler.UpdateMe)))

	// GET /users/{id} — ограниченный доступ (авторизованный видит больше)
	mux.Handle("GET /users/{id}", withAuth(http.HandlerFunc(userHandler.GetUserByID)))

	mux.Handle("GET /operators/available", withAuth(http.HandlerFunc(operatorsHandler.Available)))
	mux.Handle("PUT /operators/availability", withAuth(http.HandlerFunc(operatorsHandler.Availability)))
	mux.Handle("POST /operators/{id}/verify", withAuth(middleware.RequireRole("admin")(http.HandlerFunc(operatorsHandler.Verify))))
	mux.Handle("GET /operators/stats", withAuth(http.HandlerFunc(operatorsHandler.Stats)))

	mux.Handle("GET /users/{id}/sessions", withAuth(http.HandlerFunc(sessionsHandler.ListSessions)))
	mux.Handle("GET /users/{id}/active-sessions", withAuth(http.HandlerFunc(sessionsHandler.ListActiveSessions)))
	mux.Handle("POST /users/{id}/sessions", withAuth(http.HandlerFunc(sessionsHandler.CreateSession)))

	return http.StripPrefix("/api/v1", mux)
}
