package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	checkConfig := func(t *testing.T, got Config, err error, want Config, wantErr bool) {
		t.Helper()
		if (err != nil) != wantErr {
			t.Fatalf("Load() error = %v, want error = %v", err, wantErr)
		}
		if got != want {
			t.Errorf("Load() = %+v, want %+v", got, want)
		}
	}

	tests := []struct {
		name    string
		args    map[string]string
		want    Config
		wantErr bool
		check   func(t *testing.T, got Config, err error, want Config, wantErr bool)
	}{
		{
			name: "uses local defaults",
			args: map[string]string{},
			want: Config{
				HTTPAddr:        "127.0.0.1:8080",
				ShutdownTimeout: 10 * time.Second,
				LogLevel:        slog.LevelInfo,
			},
			check: checkConfig,
		},
		{
			name: "reads explicit configuration",
			args: map[string]string{
				"HTTP_ADDR":        "[::1]:9090",
				"SHUTDOWN_TIMEOUT": "1500ms",
				"LOG_LEVEL":        "DEBUG",
			},
			want: Config{
				HTTPAddr:        "[::1]:9090",
				ShutdownTimeout: 1500 * time.Millisecond,
				LogLevel:        slog.LevelDebug,
			},
			check: checkConfig,
		},
		{
			name: "allows container binding",
			args: map[string]string{
				"HTTP_ADDR": ":8080",
			},
			want: Config{
				HTTPAddr:        ":8080",
				ShutdownTimeout: 10 * time.Second,
				LogLevel:        slog.LevelInfo,
			},
			check: checkConfig,
		},
		{
			name:    "rejects an empty address",
			args:    map[string]string{"HTTP_ADDR": ""},
			wantErr: true,
			check:   checkConfig,
		},
		{
			name:    "rejects an address without a port",
			args:    map[string]string{"HTTP_ADDR": "127.0.0.1"},
			wantErr: true,
			check:   checkConfig,
		},
		{
			name:    "rejects a nonnumeric port",
			args:    map[string]string{"HTTP_ADDR": "127.0.0.1:http"},
			wantErr: true,
			check:   checkConfig,
		},
		{
			name:    "rejects port zero",
			args:    map[string]string{"HTTP_ADDR": "127.0.0.1:0"},
			wantErr: true,
			check:   checkConfig,
		},
		{
			name:    "rejects a port above the limit",
			args:    map[string]string{"HTTP_ADDR": "127.0.0.1:65536"},
			wantErr: true,
			check:   checkConfig,
		},
		{
			name:    "rejects an invalid duration",
			args:    map[string]string{"SHUTDOWN_TIMEOUT": "soon"},
			wantErr: true,
			check:   checkConfig,
		},
		{
			name:    "rejects a zero shutdown timeout",
			args:    map[string]string{"SHUTDOWN_TIMEOUT": "0s"},
			wantErr: true,
			check:   checkConfig,
		},
		{
			name:    "rejects a negative shutdown timeout",
			args:    map[string]string{"SHUTDOWN_TIMEOUT": "-1s"},
			wantErr: true,
			check:   checkConfig,
		},
		{
			name:    "rejects an unknown log level",
			args:    map[string]string{"LOG_LEVEL": "verbose"},
			wantErr: true,
			check:   checkConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Load(func(key string) (string, bool) {
				value, ok := tt.args[key]
				return value, ok
			})
			tt.check(t, got, err, tt.want, tt.wantErr)
		})
	}
}
