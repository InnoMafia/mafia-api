package httpapi

import (
	"context"
	_ "embed"
	"io"
	"net/http"
	"time"

	platformv1 "github.com/InnoMafia/mafia-api/gen/go/mafia/platform/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:embed openapi/mafia/platform/v1/platform.openapi.json
var openAPI string

type service struct {
	platformv1.UnimplementedPlatformServiceServer
	ping func(context.Context) error
}

func New(ctx context.Context, rpc *grpc.Server, conn *grpc.ClientConn, ping func(context.Context) error) (http.Handler, error) {
	api := &service{ping: ping}
	platformv1.RegisterPlatformServiceServer(rpc, api)

	gateway := runtime.NewServeMux()
	if err := platformv1.RegisterPlatformServiceHandler(ctx, gateway, conn); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !api.ready(r.Context()) {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, openAPI)
	})
	mux.Handle("/api/", gateway)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	}), nil
}

func (s *service) ready(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.ping(ctx) == nil
}

func (s *service) GetServiceStatus(ctx context.Context, _ *platformv1.GetServiceStatusRequest) (*platformv1.GetServiceStatusResponse, error) {
	if !s.ready(ctx) {
		return nil, status.Error(codes.Unavailable, "database unavailable")
	}
	return &platformv1.GetServiceStatusResponse{Status: "ready"}, nil
}
