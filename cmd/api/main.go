package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/InnoMafia/mafia-api/internal/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("API stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		// A connection parsing error can include credentials.
		return errors.New("DATABASE_URL could not be parsed")
	}
	defer db.Close()

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "127.0.0.1:9090"
	}
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("listen for gRPC: %w", err)
	}
	defer grpcListener.Close()

	target := *grpcListener.Addr().(*net.TCPAddr)
	if target.IP.IsUnspecified() {
		if target.IP.To4() != nil {
			target.IP = net.IPv4(127, 0, 0, 1)
		} else {
			target.IP = net.IPv6loopback
		}
	}
	conn, err := grpc.NewClient(target.String(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithNoProxy())
	if err != nil {
		return fmt.Errorf("set up local gRPC client: %w", err)
	}
	defer conn.Close()

	rpc := grpc.NewServer()
	defer rpc.Stop()
	handler, err := httpapi.New(ctx, rpc, conn, db.Ping)
	if err != nil {
		return fmt.Errorf("register HTTP gateway: %w", err)
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = "127.0.0.1:8080"
	}
	server := &http.Server{
		Addr:              httpAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}
	done := make(chan error, 2)
	go func() {
		logger.Info("starting gRPC server", "address", grpcListener.Addr().String())
		done <- rpc.Serve(grpcListener)
	}()
	go func() {
		logger.Info("starting HTTP server", "address", httpAddr)
		done <- server.ListenAndServe()
	}()

	var serveErr error
	select {
	case serveErr = <-done:
	case <-ctx.Done():
	}

	// In-flight HTTP requests still need the gRPC server while HTTP drains.
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpErr := server.Shutdown(shutdown)
	graceful := make(chan struct{})
	go func() {
		rpc.GracefulStop()
		close(graceful)
	}()
	select {
	case <-graceful:
	case <-shutdown.Done():
		rpc.Stop()
	}
	return errors.Join(serveErr, httpErr)
}
