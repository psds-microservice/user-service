package grpc

import (
	"context"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/pkg/gen/user_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Login(ctx context.Context, req *user_service.LoginRequest) (*user_service.AuthResponse, error) {
	loginReq := &dto.LoginRequest{Email: req.GetEmail(), Password: req.GetPassword()}
	if err := s.Validate.ValidateLoginRequest(loginReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.Auth.Login(ctx, loginReq.Email, loginReq.Password)
	if err != nil {
		return nil, s.mapError(err)
	}
	access, refresh, err := s.JWTConfig.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	return &user_service.AuthResponse{
		AccessToken: access, RefreshToken: refresh, ExpiresIn: 900,
		User: toProtoUserResponse(user),
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *user_service.RefreshRequest) (*user_service.AuthResponse, error) {
	refreshReq := &dto.RefreshRequest{RefreshToken: req.GetRefreshToken()}
	if err := s.Validate.ValidateRefreshRequest(refreshReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userID, err := s.JWTConfig.ValidateRefresh(refreshReq.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	user, err := s.User.GetUser(ctx, userID)
	if err != nil || user == nil {
		return nil, s.mapError(err)
	}
	access, refresh, err := s.JWTConfig.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	return &user_service.AuthResponse{
		AccessToken: access, RefreshToken: refresh, ExpiresIn: 900,
		User: toProtoUserResponse(user),
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *user_service.LogoutRequest) (*user_service.LogoutResponse, error) {
	if token := s.bearerFromContext(ctx); token != "" {
		if claims, err := s.JWTConfig.ValidateAccess(token); err == nil && s.Blacklist != nil && claims.ExpiresAt != nil {
			s.Blacklist.Add(claims.ID, claims.ExpiresAt.Time)
		}
	}
	return &user_service.LogoutResponse{}, nil
}
