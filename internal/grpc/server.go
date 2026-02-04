package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/psds-microservice/user-service/internal/auth"
	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/service"
	"github.com/psds-microservice/user-service/internal/validator"
	"github.com/psds-microservice/user-service/pkg/gen/user_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Deps — зависимости gRPC-сервера (все сервисы и инфраструктура, DI из application).
type Deps struct {
	User     *service.UserService
	Auth     *service.AuthService
	Operator *service.OperatorService
	Presence *service.PresenceService
	Session  *service.SessionService

	JWTConfig auth.Config
	Blacklist *auth.Blacklist
	Validate  *validator.Validator
}

// Server implements user_service.UserServiceServer.
type Server struct {
	user_service.UnimplementedUserServiceServer
	Deps
}

// NewServer создаёт gRPC-сервер с внедрёнными сервисами.
func NewServer(deps Deps) *Server {
	return &Server{Deps: deps}
}

func (s *Server) userIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		vals = md.Get("Authorization")
	}
	for _, v := range vals {
		if strings.HasPrefix(v, "Bearer ") {
			claims, err := s.JWTConfig.ValidateAccess(strings.TrimPrefix(v, "Bearer "))
			if err != nil {
				return ""
			}
			if s.Blacklist != nil && s.Blacklist.Contains(claims.ID) {
				return ""
			}
			return claims.UserID
		}
	}
	return ""
}

func (s *Server) bearerFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range []string{"authorization", "Authorization"} {
		vals := md.Get(key)
		for _, v := range vals {
			if strings.HasPrefix(v, "Bearer ") {
				return strings.TrimPrefix(v, "Bearer ")
			}
		}
	}
	return ""
}

func (s *Server) mapError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidOperatorStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, service.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrNotOperator),
		errors.Is(err, service.ErrOperatorNotVerifiedOrAvailable),
		errors.Is(err, service.ErrClientStreamingLimit),
		errors.Is(err, service.ErrMaxSessionsReached):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
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

func toProtoSessionResponse(r *dto.UserSessionResponse) *user_service.UserSessionResponse {
	if r == nil {
		return nil
	}
	out := &user_service.UserSessionResponse{
		Id:                r.ID,
		UserId:            r.UserID,
		SessionType:       r.SessionType,
		SessionExternalId: r.SessionExternalID,
		ParticipantRole:   r.ParticipantRole,
	}
	if !r.JoinedAt.IsZero() {
		out.JoinedAt = timestamppb.New(r.JoinedAt)
	}
	if r.LeftAt != nil {
		out.LeftAt = timestamppb.New(*r.LeftAt)
	}
	return out
}
