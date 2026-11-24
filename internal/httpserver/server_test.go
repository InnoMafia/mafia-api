package httpserver

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeShutdown(t *testing.T) {
	t.Parallel()

	type args struct {
		shutdownTimeout time.Duration
		finishRequest   bool
	}
	type serverResult struct {
		err             error
		handlerFinished bool
	}
	type requestResult struct {
		code int
		body string
		err  error
	}

	tests := []struct {
		name  string
		args  args
		check func(t *testing.T, server serverResult, request requestResult)
	}{
		{
			name: "drains an active request",
			args: args{
				shutdownTimeout: time.Second,
				finishRequest:   true,
			},
			check: func(t *testing.T, server serverResult, request requestResult) {
				t.Helper()
				if server.err != nil {
					t.Fatalf("Serve() error = %v", server.err)
				}
				if !server.handlerFinished {
					t.Error("Serve() returned before the active handler finished")
				}
				if request.err != nil {
					t.Fatalf("request error = %v", request.err)
				}
				if request.code != http.StatusOK || request.body != "done" {
					t.Errorf("response = %+v, want status 200 and body done", request)
				}
			},
		},
		{
			name: "closes a request after the shutdown deadline",
			args: args{
				shutdownTimeout: 20 * time.Millisecond,
				finishRequest:   false,
			},
			check: func(t *testing.T, server serverResult, request requestResult) {
				t.Helper()
				if !errors.Is(server.err, context.DeadlineExceeded) {
					t.Fatalf("Serve() error = %v, want deadline exceeded", server.err)
				}
				if request.err == nil {
					t.Errorf("request succeeded after forced shutdown: %+v", request)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Shutdown mutates the whole server, so each case needs its own listener.
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("listen: %v", err)
			}
			t.Cleanup(func() { _ = listener.Close() })

			started := make(chan struct{})
			finish := make(chan struct{})
			handlerFinished := make(chan struct{})
			shutdownStarted := make(chan struct{})
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(handlerFinished)
				close(started)
				select {
				case <-finish:
					_, _ = io.WriteString(w, "done")
				case <-r.Context().Done():
				}
			})
			logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
			server := New(handler, logger)
			t.Cleanup(func() { _ = server.Close() })
			server.RegisterOnShutdown(func() { close(shutdownStarted) })

			serverCtx, stop := context.WithCancel(ctx)
			defer stop()
			serverDone := make(chan serverResult, 1)
			go func() {
				result := serverResult{err: Serve(serverCtx, server, listener, tt.args.shutdownTimeout)}
				select {
				case <-handlerFinished:
					result.handlerFinished = true
				default:
				}
				serverDone <- result
			}()

			transport := &http.Transport{}
			t.Cleanup(transport.CloseIdleConnections)
			client := &http.Client{Transport: transport}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+listener.Addr().String(), nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			requestDone := make(chan requestResult, 1)
			go func() {
				result := requestResult{}
				response, err := client.Do(req)
				if err != nil {
					result.err = err
				} else {
					body, readErr := io.ReadAll(response.Body)
					result = requestResult{
						code: response.StatusCode,
						body: string(body),
						err:  errors.Join(readErr, response.Body.Close()),
					}
				}
				requestDone <- result
			}()

			receive(t, ctx, started)
			stop()
			receive(t, ctx, shutdownStarted)
			if tt.args.finishRequest {
				close(finish)
			}
			serveResult := receive(t, ctx, serverDone)
			responseResult := receive(t, ctx, requestDone)
			receive(t, ctx, handlerFinished)
			tt.check(t, serveResult, responseResult)
		})
	}
}

func TestServeFailure(t *testing.T) {
	t.Parallel()

	type args struct {
		addr            string
		shutdownTimeout time.Duration
	}
	tests := []struct {
		name  string
		args  args
		check func(t *testing.T, err error)
	}{
		{
			name: "reports a closed listener",
			args: args{
				addr:            "127.0.0.1:0",
				shutdownTimeout: time.Second,
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, net.ErrClosed) {
					t.Errorf("Serve() error = %v, want a closed listener", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			listener, err := net.Listen("tcp", tt.args.addr)
			if err != nil {
				t.Fatalf("listen: %v", err)
			}
			if err := listener.Close(); err != nil {
				t.Fatalf("close listener: %v", err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
			server := New(http.NotFoundHandler(), logger)

			err = Serve(ctx, server, listener, tt.args.shutdownTimeout)
			tt.check(t, err)
		})
	}
}

// Bound channel waits so lifecycle regressions fail instead of hanging the test run.
func receive[T any](t *testing.T, ctx context.Context, channel <-chan T) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-ctx.Done():
		t.Fatalf("waiting for HTTP lifecycle: %v", ctx.Err())
		var zero T
		return zero
	}
}
