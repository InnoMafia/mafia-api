package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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
		{"API status while database is down", "/api/v1/status", true, http.StatusServiceUnavailable, "unavailable"},
		{"API status while database is healthy", "/api/v1/status", false, http.StatusOK, "ready"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := New(func(ctx context.Context) error {
				if tc.path == "/healthz" {
					t.Fatal("liveness must be independent of the database")
				}
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("database check must have a deadline")
				}
				if tc.dbDown {
					return errors.New("database offline")
				}
				return nil
			})
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != tc.wantCode {
				t.Fatalf("got HTTP %d, want %d", recorder.Code, tc.wantCode)
			}
			if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			if tc.wantState == "" {
				return
			}
			var body struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid status JSON: %v", err)
			}
			if body.Status != tc.wantState {
				t.Errorf("status = %q, want %q", body.Status, tc.wantState)
			}
		})
	}
}
