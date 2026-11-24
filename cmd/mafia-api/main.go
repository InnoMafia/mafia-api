package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/InnoMafia/mafia-api/internal/config"
	"github.com/InnoMafia/mafia-api/internal/httpapi"
	"github.com/InnoMafia/mafia-api/internal/httpserver"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		logger.Error("cannot listen", "error", err)
		os.Exit(1)
	}

	server := httpserver.New(httpapi.NewHandler(), logger)
	logger.Info("HTTP server listening", "addr", listener.Addr().String())

	if err := httpserver.Serve(ctx, server, listener, cfg.ShutdownTimeout); err != nil {
		logger.Error("HTTP server failed", "error", err)
		os.Exit(1)
	}
	logger.Info("HTTP server stopped")
}
