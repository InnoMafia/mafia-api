package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/InnoMafia/mafia-api"
	platformv1 "github.com/InnoMafia/mafia-api/gen/go/mafia/platform/v1"
	"github.com/InnoMafia/mafia-api/internal/handler/platform"
	"github.com/InnoMafia/mafia-api/internal/usecase"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/swaggest/swgui/v5emb"
	"google.golang.org/grpc"
)

func New(ctx context.Context, rpc *grpc.Server, conn *grpc.ClientConn, getServiceStatus *usecase.GetServiceStatus) (http.Handler, error) {
	platformv1.RegisterPlatformServiceServer(rpc, platform.New(getServiceStatus))

	gateway := runtime.NewServeMux()
	if err := platformv1.RegisterPlatformServiceHandler(ctx, gateway, conn); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !getServiceStatus.Execute(r.Context()) {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	const openAPIPath = "/api/openapi.json"
	mux.HandleFunc("GET "+openAPIPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, mafiaapi.OpenAPI())
	})
	const docsPath = "/api/docs/"
	mux.Handle("GET "+docsPath, v5emb.New("InnoMafia API", openAPIPath, docsPath))
	mux.Handle("/api/", gateway)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	}), nil
}
