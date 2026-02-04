package grpc

import (
	"context"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/pkg/gen/user_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateUser(ctx context.Context, req *user_service.CreateUserRequest) (*user_service.UserResponse, error) {
	createReq := &dto.CreateUserRequest{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Password: req.GetPassword(),
	}
	if err := s.Validate.ValidateCreateUserRequest(createReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.User.CreateUser(ctx, createReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) GetUser(ctx context.Context, req *user_service.GetUserRequest) (*user_service.UserResponse, error) {
	resp, err := s.User.GetUser(ctx, req.GetId())
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) UpdateUser(ctx context.Context, req *user_service.UpdateUserRequest) (*user_service.UserResponse, error) {
	updateReq := &dto.UpdateUserRequest{
		ID:       req.GetId(),
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Password: req.GetPassword(),
		Status:   req.GetStatus(),
	}
	if err := s.Validate.ValidateUpdateUserRequest(updateReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.User.UpdateUser(ctx, updateReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) DeleteUser(ctx context.Context, req *user_service.DeleteUserRequest) (*user_service.DeleteUserResponse, error) {
	if err := s.User.DeleteUser(ctx, req.GetId()); err != nil {
		return nil, s.mapError(err)
	}
	return &user_service.DeleteUserResponse{Success: true}, nil
}

func (s *Server) GetMe(ctx context.Context, req *user_service.GetMeRequest) (*user_service.UserResponse, error) {
	userID := s.userIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	resp, err := s.User.GetUser(ctx, userID)
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) UpdateMe(ctx context.Context, req *user_service.UpdateUserRequest) (*user_service.UserResponse, error) {
	userID := s.userIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	resp, err := s.User.UpdateUser(ctx, &dto.UpdateUserRequest{
		ID:       userID,
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Password: req.GetPassword(),
		Status:   req.GetStatus(),
	})
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) Register(ctx context.Context, req *user_service.RegisterRequest) (*user_service.AuthResponse, error) {
	regReq := &dto.RegisterRequest{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Role:     req.GetRole(),
	}
	if err := s.Validate.ValidateRegisterRequest(regReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.User.CreateUser(ctx, &dto.CreateUserRequest{
		Username: regReq.Username,
		Email:    regReq.Email,
		Password: regReq.Password,
		Role:     regReq.Role,
	})
	if err != nil {
		return nil, s.mapError(err)
	}
	access, refresh, err := s.JWTConfig.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	return &user_service.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    900,
		User:         toProtoUserResponse(user),
	}, nil
}
