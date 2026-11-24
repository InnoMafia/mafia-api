package config

import (
	"errors"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	ShutdownTimeout time.Duration
	LogLevel        slog.Level
}

func Load(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{
		HTTPAddr:        "127.0.0.1:8080",
		ShutdownTimeout: 10 * time.Second,
		LogLevel:        slog.LevelInfo,
	}

	if value, ok := lookup("HTTP_ADDR"); ok {
		cfg.HTTPAddr = value
	}
	_, port, err := net.SplitHostPort(cfg.HTTPAddr)
	if err != nil || strings.TrimSpace(cfg.HTTPAddr) != cfg.HTTPAddr {
		return Config{}, errors.New("HTTP_ADDR must be a host:port address")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, errors.New("HTTP_ADDR must have a port between 1 and 65535")
	}

	if value, ok := lookup("SHUTDOWN_TIMEOUT"); ok {
		cfg.ShutdownTimeout, err = time.ParseDuration(value)
		if err != nil || cfg.ShutdownTimeout <= 0 {
			return Config{}, errors.New("SHUTDOWN_TIMEOUT must be a positive duration")
		}
	}

	if value, ok := lookup("LOG_LEVEL"); ok {
		switch strings.ToLower(value) {
		case "debug":
			cfg.LogLevel = slog.LevelDebug
		case "info":
			cfg.LogLevel = slog.LevelInfo
		case "warn":
			cfg.LogLevel = slog.LevelWarn
		case "error":
			cfg.LogLevel = slog.LevelError
		default:
			return Config{}, errors.New("LOG_LEVEL must be debug, info, warn, or error")
		}
	}

	return cfg, nil
}
