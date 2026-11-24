package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	type response struct {
		code         int
		body         string
		contentType  string
		cacheControl string
		allow        string
	}
	checkResponse := func(t *testing.T, got *httptest.ResponseRecorder, want response) {
		t.Helper()
		actual := response{
			code:         got.Code,
			body:         got.Body.String(),
			contentType:  got.Header().Get("Content-Type"),
			cacheControl: got.Header().Get("Cache-Control"),
			allow:        got.Header().Get("Allow"),
		}
		if actual != want {
			t.Errorf("response = %+v, want %+v", actual, want)
		}
	}

	tests := []struct {
		name  string
		req   *http.Request
		want  response
		check func(t *testing.T, got *httptest.ResponseRecorder, want response)
	}{
		{
			name: "reports liveness",
			req:  httptest.NewRequest(http.MethodGet, "/healthz", nil),
			want: response{
				code:         http.StatusOK,
				body:         "{\"status\":\"ok\"}\n",
				contentType:  "application/json",
				cacheControl: "no-store",
			},
			check: checkResponse,
		},
		{
			name: "reports readiness",
			req:  httptest.NewRequest(http.MethodGet, "/readyz", nil),
			want: response{
				code:         http.StatusOK,
				body:         "{\"status\":\"ok\"}\n",
				contentType:  "application/json",
				cacheControl: "no-store",
			},
			check: checkResponse,
		},
		{
			name: "supports head without a body",
			req:  httptest.NewRequest(http.MethodHead, "/healthz", nil),
			want: response{
				code:         http.StatusOK,
				contentType:  "application/json",
				cacheControl: "no-store",
			},
			check: checkResponse,
		},
		{
			name: "rejects a write method",
			req:  httptest.NewRequest(http.MethodPost, "/readyz", nil),
			want: response{
				code:        http.StatusMethodNotAllowed,
				body:        "Method Not Allowed\n",
				contentType: "text/plain; charset=utf-8",
				allow:       "GET, HEAD",
			},
			check: checkResponse,
		},
		{
			name: "does not expose unknown routes",
			req:  httptest.NewRequest(http.MethodGet, "/games", nil),
			want: response{
				code:        http.StatusNotFound,
				body:        "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
			},
			check: checkResponse,
		},
	}

	handler := NewHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, tt.req)
			tt.check(t, recorder, tt.want)
		})
	}
}
