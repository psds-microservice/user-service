package grpc

import (
	"context"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/service"
	"github.com/psds-microservice/user-service/pkg/gen/user_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements user_service.UserServiceServer.
type Server struct {
	user_service.UnimplementedUserServiceServer
	svc service.IUserService
}

func NewServer(svc service.IUserService) *Server {
	return &Server{svc: svc}
}

func (s *Server) CreateUser(ctx context.Context, req *user_service.CreateUserRequest) (*user_service.UserResponse, error) {
	resp, err := s.svc.CreateUser(ctx, &dto.CreateUserRequest{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return &user_service.UserResponse{Error: err.Error()}, nil
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) GetUser(ctx context.Context, req *user_service.GetUserRequest) (*user_service.UserResponse, error) {
	resp, err := s.svc.GetUser(ctx, req.GetId())
	if err != nil {
		return &user_service.UserResponse{Error: err.Error()}, nil
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) UpdateUser(ctx context.Context, req *user_service.UpdateUserRequest) (*user_service.UserResponse, error) {
	resp, err := s.svc.UpdateUser(ctx, &dto.UpdateUserRequest{
		ID:       req.GetId(),
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Password: req.GetPassword(),
		Status:   req.GetStatus(),
	})
	if err != nil {
		return &user_service.UserResponse{Error: err.Error()}, nil
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) DeleteUser(ctx context.Context, req *user_service.DeleteUserRequest) (*user_service.DeleteUserResponse, error) {
	err := s.svc.DeleteUser(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user_service.DeleteUserResponse{Success: true}, nil
}

func (s *Server) Login(ctx context.Context, req *user_service.LoginRequest) (*user_service.UserResponse, error) {
	resp, err := s.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return &user_service.UserResponse{Error: err.Error()}, nil
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) ValidateUserSession(ctx context.Context, req *user_service.ValidateUserSessionRequest) (*user_service.ValidateUserSessionResponse, error) {
	allowed, err := s.svc.ValidateUserSession(ctx, req.GetUserId(), req.GetSessionExternalId(), req.GetParticipantRole())
	if err != nil {
		return &user_service.ValidateUserSessionResponse{Allowed: false, Error: err.Error()}, nil
	}
	return &user_service.ValidateUserSessionResponse{Allowed: allowed}, nil
}

func (s *Server) UpdateUserPresence(ctx context.Context, req *user_service.UpdateUserPresenceRequest) (*user_service.UpdateUserPresenceResponse, error) {
	err := s.svc.UpdatePresence(ctx, req.GetUserId(), req.GetIsOnline())
	if err != nil {
		return &user_service.UpdateUserPresenceResponse{Success: false, Error: err.Error()}, nil
	}
	return &user_service.UpdateUserPresenceResponse{Success: true}, nil
}

func (s *Server) GetAvailableOperators(ctx context.Context, req *user_service.GetAvailableOperatorsRequest) (*user_service.GetAvailableOperatorsResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.GetOffset())
	if offset < 0 {
		offset = 0
	}
	list, total, err := s.svc.ListAvailableOperators(ctx, limit, offset)
	if err != nil {
		return &user_service.GetAvailableOperatorsResponse{Error: err.Error()}, nil
	}
	operators := make([]*user_service.UserResponse, len(list))
	for i := range list {
		operators[i] = toProtoUserResponse(list[i])
	}
	return &user_service.GetAvailableOperatorsResponse{Operators: operators, Total: total}, nil
}

func (s *Server) UpdateOperatorStatus(ctx context.Context, req *user_service.UpdateOperatorStatusRequest) (*user_service.UpdateOperatorStatusResponse, error) {
	_, err := s.svc.UpdateAvailability(ctx, req.GetUserId(), req.GetIsAvailable())
	if err != nil {
		return &user_service.UpdateOperatorStatusResponse{Success: false, Error: err.Error()}, nil
	}
	return &user_service.UpdateOperatorStatusResponse{Success: true}, nil
}

func toProtoUserResponse(r *dto.UserResponse) *user_service.UserResponse {
	if r == nil {
		return nil
	}
	out := &user_service.UserResponse{
		Id:       r.ID,
		Username: r.Username,
		Email:    r.Email,
		Phone:    r.Phone,
		Status:   r.Status,
	}
	if !r.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(r.CreatedAt)
	}
	if !r.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(r.UpdatedAt)
	}
	return out
}
