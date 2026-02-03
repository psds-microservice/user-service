package handler

import (
	"net/http"

	"github.com/psds-microservice/user-service/internal/auth"
	"github.com/psds-microservice/user-service/internal/middleware"
	"github.com/psds-microservice/user-service/internal/service"
	"github.com/psds-microservice/user-service/pkg/constants"
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

	// Публичные (эквиваленты gRPC без JWT)
	mux.HandleFunc("POST "+constants.PathAuthRegister, authHandler.Register)
	mux.HandleFunc("POST "+constants.PathAuthLogin, authHandler.Login)
	mux.HandleFunc("POST "+constants.PathAuthRefresh, authHandler.Refresh)
	mux.HandleFunc("POST "+constants.PathAuthLogout, authHandler.Logout)
	mux.HandleFunc("POST "+constants.PathUsers, userHandler.CreateUserJSON)
	mux.HandleFunc("POST "+constants.PathSessionsValidate, sessionsHandler.ValidateSession)

	// Требуют JWT
	withAuth := middleware.JWTAuth(jwtConfig, blacklist)
	mux.Handle("GET "+constants.PathUsersMe, withAuth(http.HandlerFunc(meHandler.GetMe)))
	mux.Handle("PUT "+constants.PathUsersMe, withAuth(http.HandlerFunc(meHandler.UpdateMe)))

	// Пользователи: GET/PUT/DELETE по id, presence
	mux.Handle("GET "+constants.PathUsersID, withAuth(http.HandlerFunc(userHandler.GetUserByID)))
	mux.Handle("PUT "+constants.PathUsersPresence, withAuth(http.HandlerFunc(userHandler.UpdatePresence)))
	mux.Handle("PUT "+constants.PathUsersID, withAuth(http.HandlerFunc(userHandler.UpdateUserByID)))
	mux.Handle("DELETE "+constants.PathUsersID, withAuth(http.HandlerFunc(userHandler.DeleteUserByID)))

	// Операторы
	mux.Handle("GET "+constants.PathOperatorsAvailable, withAuth(http.HandlerFunc(operatorsHandler.Available)))
	mux.Handle("PUT "+constants.PathOperatorsAvailability, withAuth(http.HandlerFunc(operatorsHandler.Availability)))
	mux.Handle("PUT "+constants.PathOperatorsIDAvailability, withAuth(middleware.RequireRole(constants.RoleAdmin)(http.HandlerFunc(operatorsHandler.SetAvailabilityByID))))
	mux.Handle("POST "+constants.PathOperatorsIDVerify, withAuth(middleware.RequireRole(constants.RoleAdmin)(http.HandlerFunc(operatorsHandler.Verify))))
	mux.Handle("GET "+constants.PathOperatorsStats, withAuth(http.HandlerFunc(operatorsHandler.Stats)))

	// Сессии пользователя
	mux.Handle("GET "+constants.PathUsersSessions, withAuth(http.HandlerFunc(sessionsHandler.ListSessions)))
	mux.Handle("GET "+constants.PathUsersActiveSessions, withAuth(http.HandlerFunc(sessionsHandler.ListActiveSessions)))
	mux.Handle("POST "+constants.PathUsersSessions, withAuth(http.HandlerFunc(sessionsHandler.CreateSession)))

	return http.StripPrefix(constants.BasePathAPI, mux)
}
