package grpc

import (
	"context"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/pkg/gen/user_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUserSessions(ctx context.Context, req *user_service.GetUserSessionsRequest) (*user_service.GetUserSessionsResponse, error) {
	limit, offset := int(req.GetLimit()), int(req.GetOffset())
	if limit <= 0 {
		limit = 20
	}
	list, total, err := s.Session.GetUserSessions(ctx, req.GetId(), limit, offset)
	if err != nil {
		return nil, s.mapError(err)
	}
	out := &user_service.GetUserSessionsResponse{
		Sessions: make([]*user_service.UserSessionResponse, len(list)),
		Total:    total,
	}
	for i := range list {
		out.Sessions[i] = toProtoSessionResponse(list[i])
	}
	return out, nil
}

func (s *Server) GetActiveSessions(ctx context.Context, req *user_service.GetActiveSessionsRequest) (*user_service.GetActiveSessionsResponse, error) {
	list, err := s.Session.GetActiveSessions(ctx, req.GetId())
	if err != nil {
		return nil, s.mapError(err)
	}
	out := &user_service.GetActiveSessionsResponse{
		Sessions: make([]*user_service.UserSessionResponse, len(list)),
	}
	for i := range list {
		out.Sessions[i] = toProtoSessionResponse(list[i])
	}
	return out, nil
}

func (s *Server) CreateSession(ctx context.Context, req *user_service.CreateSessionRequest) (*user_service.UserSessionResponse, error) {
	userID := req.GetId()
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	createReq := &dto.CreateSessionRequest{
		SessionType:       req.GetSessionType(),
		SessionExternalID: req.GetSessionExternalId(),
		ParticipantRole:   req.GetParticipantRole(),
	}
	if err := s.Validate.ValidateCreateSessionRequest(createReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	session, err := s.Session.CreateSession(ctx, userID, createReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoSessionResponse(session), nil
}

func (s *Server) ValidateUserSession(ctx context.Context, req *user_service.ValidateUserSessionRequest) (*user_service.ValidateUserSessionResponse, error) {
	if err := s.Validate.ValidateSessionValidateRequest(req.GetUserId(), req.GetSessionExternalId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	allowed, err := s.Session.ValidateUserSession(ctx, req.GetUserId(), req.GetSessionExternalId(), req.GetParticipantRole())
	if err != nil {
		return nil, s.mapError(err)
	}
	return &user_service.ValidateUserSessionResponse{Allowed: allowed}, nil
}
