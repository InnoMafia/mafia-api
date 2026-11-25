package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	platformv1 "github.com/InnoMafia/mafia-api/gen/go/mafia/platform/v1"
	"github.com/InnoMafia/mafia-api/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestProbes(t *testing.T) {
	for _, tc := range []struct {
		name      string
		path      string
		dbDown    bool
		wantCode  int
		wantState string
	}{
		{"liveness while database is down", "/healthz", true, http.StatusNoContent, ""},
		{"readiness while database is down", "/readyz", true, http.StatusServiceUnavailable, ""},
		{"readiness while database is healthy", "/readyz", false, http.StatusNoContent, ""},
		{"API status while database is down", "/api/v1/status", true, http.StatusServiceUnavailable, ""},
		{"API status while database is healthy", "/api/v1/status", false, http.StatusOK, "ready"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler, client := startTestServer(t, func(ctx context.Context) error {
				if tc.path == "/healthz" {
					t.Error("liveness must be independent of the database")
				}
				if _, ok := ctx.Deadline(); !ok {
					t.Error("database check must have a deadline")
				}
				if tc.dbDown {
					return errors.New("database offline")
				}
				return nil
			})
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequestWithContext(ctx, http.MethodGet, tc.path, nil))
			if response.Code != tc.wantCode {
				t.Fatalf("got HTTP %d, want %d", response.Code, tc.wantCode)
			}
			if got := response.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			if tc.path != "/api/v1/status" {
				return
			}

			var body struct {
				Status string `json:"status"`
				Code   int    `json:"code"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid status JSON: %v", err)
			}
			rpc, err := client.GetServiceStatus(ctx, &platformv1.GetServiceStatusRequest{})
			if tc.dbDown {
				if body.Code != int(codes.Unavailable) || status.Code(err) != codes.Unavailable {
					t.Errorf("HTTP and gRPC should both report unavailable, got %s and %v", response.Body, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("gRPC status: %v", err)
			}
			if body.Status != tc.wantState || rpc.GetStatus() != tc.wantState {
				t.Errorf("HTTP status = %q and gRPC status = %q, want %q", body.Status, rpc.GetStatus(), tc.wantState)
			}
		})
	}
}

func startTestServer(t *testing.T, ping func(context.Context) error) (http.Handler, platformv1.PlatformServiceClient) {
	t.Helper()

	listener := bufconn.Listen(1 << 20)
	rpc := grpc.NewServer()
	t.Cleanup(func() {
		rpc.Stop()
		_ = listener.Close()
	})

	conn, err := grpc.NewClient("passthrough:///test",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	getServiceStatus := usecase.NewGetServiceStatus(statusRepositoryFunc(ping))
	handler, err := New(t.Context(), rpc, conn, getServiceStatus)
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = rpc.Serve(listener) }()
	return handler, platformv1.NewPlatformServiceClient(conn)
}

type statusRepositoryFunc func(context.Context) error

func (f statusRepositoryFunc) Ping(ctx context.Context) error {
	return f(ctx)
}
