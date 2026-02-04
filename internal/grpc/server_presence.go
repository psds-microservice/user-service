package grpc

import (
	"context"

	"github.com/psds-microservice/user-service/pkg/gen/user_service"
)

func (s *Server) UpdateUserPresence(ctx context.Context, req *user_service.UpdateUserPresenceRequest) (*user_service.UpdateUserPresenceResponse, error) {
	if err := s.Presence.UpdatePresence(ctx, req.GetUserId(), req.GetIsOnline()); err != nil {
		return nil, s.mapError(err)
	}
	return &user_service.UpdateUserPresenceResponse{Success: true}, nil
}
