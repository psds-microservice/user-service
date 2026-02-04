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

// Server implements user_service.UserServiceServer.
type Server struct {
	user_service.UnimplementedUserServiceServer
	svc       service.IUserService
	jwtCfg    auth.Config
	blacklist *auth.Blacklist
	validate  *validator.Validator
}

func NewServer(svc service.IUserService, jwtCfg auth.Config, blacklist *auth.Blacklist, v *validator.Validator) *Server {
	return &Server{svc: svc, jwtCfg: jwtCfg, blacklist: blacklist, validate: v}
}

// userIDFromContext читает Bearer из gRPC metadata и возвращает user_id или "".
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
			claims, err := s.jwtCfg.ValidateAccess(strings.TrimPrefix(v, "Bearer "))
			if err != nil {
				return ""
			}
			if s.blacklist != nil && s.blacklist.Contains(claims.ID) {
				return ""
			}
			return claims.UserID
		}
	}
	return ""
}

// mapError маппит доменные ошибки сервиса в gRPC-коды.
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
		// Сообщение можно упростить, чтобы не светить детали.
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, service.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrNotOperator),
		errors.Is(err, service.ErrOperatorNotVerifiedOrAvailable),
		errors.Is(err, service.ErrClientStreamingLimit),
		errors.Is(err, service.ErrMaxSessionsReached):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		// Для технических/неожиданных ошибок.
		return status.Error(codes.Internal, "internal error")
	}
}

func (s *Server) CreateUser(ctx context.Context, req *user_service.CreateUserRequest) (*user_service.UserResponse, error) {
	createReq := &dto.CreateUserRequest{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Password: req.GetPassword(),
	}
	if err := s.validate.ValidateCreateUserRequest(createReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.svc.CreateUser(ctx, createReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) GetUser(ctx context.Context, req *user_service.GetUserRequest) (*user_service.UserResponse, error) {
	resp, err := s.svc.GetUser(ctx, req.GetId())
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
	if err := s.validate.ValidateUpdateUserRequest(updateReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.svc.UpdateUser(ctx, updateReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) DeleteUser(ctx context.Context, req *user_service.DeleteUserRequest) (*user_service.DeleteUserResponse, error) {
	err := s.svc.DeleteUser(ctx, req.GetId())
	if err != nil {
		return nil, s.mapError(err)
	}
	return &user_service.DeleteUserResponse{Success: true}, nil
}

func (s *Server) Login(ctx context.Context, req *user_service.LoginRequest) (*user_service.AuthResponse, error) {
	loginReq := &dto.LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	if err := s.validate.ValidateLoginRequest(loginReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.svc.Login(ctx, loginReq.Email, loginReq.Password)
	if err != nil {
		return nil, s.mapError(err)
	}
	access, refresh, err := s.jwtCfg.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
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

func (s *Server) Register(ctx context.Context, req *user_service.RegisterRequest) (*user_service.AuthResponse, error) {
	regReq := &dto.RegisterRequest{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Role:     req.GetRole(),
	}
	if err := s.validate.ValidateRegisterRequest(regReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.svc.CreateUser(ctx, &dto.CreateUserRequest{
		Username: regReq.Username,
		Email:    regReq.Email,
		Password: regReq.Password,
		Role:     regReq.Role,
	})
	if err != nil {
		return nil, s.mapError(err)
	}
	access, refresh, err := s.jwtCfg.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
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

func (s *Server) Refresh(ctx context.Context, req *user_service.RefreshRequest) (*user_service.AuthResponse, error) {
	refreshReq := &dto.RefreshRequest{RefreshToken: req.GetRefreshToken()}
	if err := s.validate.ValidateRefreshRequest(refreshReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userID, err := s.jwtCfg.ValidateRefresh(refreshReq.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	user, err := s.svc.GetUser(ctx, userID)
	if err != nil || user == nil {
		return nil, s.mapError(err)
	}
	access, refresh, err := s.jwtCfg.GeneratePair(user.ID, user.Email, user.Role, user.OperatorStatus, user.IsAvailable)
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

func (s *Server) Logout(ctx context.Context, req *user_service.LogoutRequest) (*user_service.LogoutResponse, error) {
	// Инвалидация по Bearer из metadata
	if token := s.bearerFromContext(ctx); token != "" {
		if claims, err := s.jwtCfg.ValidateAccess(token); err == nil && s.blacklist != nil {
			exp := claims.ExpiresAt
			if exp != nil {
				s.blacklist.Add(claims.ID, exp.Time)
			}
		}
	}
	return &user_service.LogoutResponse{}, nil
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

func (s *Server) GetMe(ctx context.Context, req *user_service.GetMeRequest) (*user_service.UserResponse, error) {
	userID := s.userIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	resp, err := s.svc.GetUser(ctx, userID)
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
	resp, err := s.svc.UpdateUser(ctx, &dto.UpdateUserRequest{
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

func (s *Server) GetUserSessions(ctx context.Context, req *user_service.GetUserSessionsRequest) (*user_service.GetUserSessionsResponse, error) {
	limit, offset := int(req.GetLimit()), int(req.GetOffset())
	if limit <= 0 {
		limit = 20
	}
	list, total, err := s.svc.GetUserSessions(ctx, req.GetId(), limit, offset)
	if err != nil {
		return nil, s.mapError(err)
	}
	out := &user_service.GetUserSessionsResponse{Sessions: make([]*user_service.UserSessionResponse, len(list)), Total: total}
	for i := range list {
		out.Sessions[i] = toProtoSessionResponse(list[i])
	}
	return out, nil
}

func (s *Server) GetActiveSessions(ctx context.Context, req *user_service.GetActiveSessionsRequest) (*user_service.GetActiveSessionsResponse, error) {
	list, err := s.svc.GetActiveSessions(ctx, req.GetId())
	if err != nil {
		return nil, s.mapError(err)
	}
	out := &user_service.GetActiveSessionsResponse{Sessions: make([]*user_service.UserSessionResponse, len(list))}
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
	if err := s.validate.ValidateCreateSessionRequest(createReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	session, err := s.svc.CreateSession(ctx, userID, createReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoSessionResponse(session), nil
}

func (s *Server) UpdateOperatorAvailability(ctx context.Context, req *user_service.UpdateOperatorStatusRequest) (*user_service.UpdateOperatorStatusResponse, error) {
	userID := s.userIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	_, err := s.svc.UpdateAvailability(ctx, userID, req.GetIsAvailable())
	if err != nil {
		return nil, s.mapError(err)
	}
	return &user_service.UpdateOperatorStatusResponse{Success: true}, nil
}

func (s *Server) ValidateUserSession(ctx context.Context, req *user_service.ValidateUserSessionRequest) (*user_service.ValidateUserSessionResponse, error) {
	if err := s.validate.ValidateSessionValidateRequest(req.GetUserId(), req.GetSessionExternalId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	allowed, err := s.svc.ValidateUserSession(ctx, req.GetUserId(), req.GetSessionExternalId(), req.GetParticipantRole())
	if err != nil {
		return nil, s.mapError(err)
	}
	return &user_service.ValidateUserSessionResponse{Allowed: allowed}, nil
}

func (s *Server) UpdateUserPresence(ctx context.Context, req *user_service.UpdateUserPresenceRequest) (*user_service.UpdateUserPresenceResponse, error) {
	err := s.svc.UpdatePresence(ctx, req.GetUserId(), req.GetIsOnline())
	if err != nil {
		return nil, s.mapError(err)
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
		return nil, s.mapError(err)
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
		return nil, s.mapError(err)
	}
	return &user_service.UpdateOperatorStatusResponse{Success: true}, nil
}

func (s *Server) VerifyOperator(ctx context.Context, req *user_service.VerifyOperatorRequest) (*user_service.UserResponse, error) {
	resp, err := s.svc.VerifyOperator(ctx, req.GetId(), req.GetStatus())
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) GetOperatorStats(ctx context.Context, req *user_service.GetOperatorStatsRequest) (*user_service.GetOperatorStatsResponse, error) {
	list, total, err := s.svc.ListAvailableOperators(ctx, 100, 0)
	if err != nil {
		return nil, s.mapError(err)
	}
	var rating float64
	for _, u := range list {
		rating += u.Rating
	}
	if len(list) > 0 {
		rating /= float64(len(list))
	}
	return &user_service.GetOperatorStatsResponse{TotalSessions: total, Rating: rating}, nil
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
