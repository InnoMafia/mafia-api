package handler

import (
	"context"

	platformv1 "github.com/InnoMafia/mafia-api/gen/go/mafia/platform/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) GetServiceStatus(ctx context.Context, _ *platformv1.GetServiceStatusRequest) (*platformv1.GetServiceStatusResponse, error) {
	if !s.getServiceStatus.Execute(ctx) {
		return nil, status.Error(codes.Unavailable, "database unavailable")
	}
	return &platformv1.GetServiceStatusResponse{Status: "ready"}, nil
}
