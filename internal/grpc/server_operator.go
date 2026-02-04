package grpc

import (
	"context"

	"github.com/psds-microservice/user-service/pkg/gen/user_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetAvailableOperators(ctx context.Context, req *user_service.GetAvailableOperatorsRequest) (*user_service.GetAvailableOperatorsResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.GetOffset())
	if offset < 0 {
		offset = 0
	}
	list, total, err := s.Operator.ListAvailableOperators(ctx, limit, offset)
	if err != nil {
		return nil, s.mapError(err)
	}
	operators := make([]*user_service.UserResponse, len(list))
	for i := range list {
		operators[i] = toProtoUserResponse(list[i])
	}
	return &user_service.GetAvailableOperatorsResponse{Operators: operators, Total: total}, nil
}

func (s *Server) UpdateOperatorAvailability(ctx context.Context, req *user_service.UpdateOperatorStatusRequest) (*user_service.UpdateOperatorStatusResponse, error) {
	userID := s.userIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	_, err := s.Operator.UpdateAvailability(ctx, userID, req.GetIsAvailable())
	if err != nil {
		return nil, s.mapError(err)
	}
	return &user_service.UpdateOperatorStatusResponse{Success: true}, nil
}

func (s *Server) UpdateOperatorStatus(ctx context.Context, req *user_service.UpdateOperatorStatusRequest) (*user_service.UpdateOperatorStatusResponse, error) {
	_, err := s.Operator.UpdateAvailability(ctx, req.GetUserId(), req.GetIsAvailable())
	if err != nil {
		return nil, s.mapError(err)
	}
	return &user_service.UpdateOperatorStatusResponse{Success: true}, nil
}

func (s *Server) VerifyOperator(ctx context.Context, req *user_service.VerifyOperatorRequest) (*user_service.UserResponse, error) {
	resp, err := s.Operator.VerifyOperator(ctx, req.GetId(), req.GetStatus())
	if err != nil {
		return nil, s.mapError(err)
	}
	return toProtoUserResponse(resp), nil
}

func (s *Server) GetOperatorStats(ctx context.Context, req *user_service.GetOperatorStatsRequest) (*user_service.GetOperatorStatsResponse, error) {
	list, total, err := s.Operator.ListAvailableOperators(ctx, 100, 0)
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
