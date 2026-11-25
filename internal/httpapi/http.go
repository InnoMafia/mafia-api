package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/InnoMafia/mafia-api/internal/httpapi/contract"
)

type service struct {
	ping func(context.Context) error
}

func New(ping func(context.Context) error) http.Handler {
	api := service{ping: ping}
	versioned := contract.HandlerWithOptions(contract.NewStrictHandler(api, nil), contract.ChiServerOptions{})
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
	mux.Handle("/api/v1/", versioned)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})
}

func (s service) ready(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.ping(ctx) == nil
}

func (s service) GetServiceStatus(ctx context.Context, _ contract.GetServiceStatusRequestObject) (contract.GetServiceStatusResponseObject, error) {
	if s.ready(ctx) {
		return contract.GetServiceStatus200JSONResponse{Status: contract.Ready}, nil
	}
	return contract.GetServiceStatus503JSONResponse{Status: contract.Unavailable}, nil
}
