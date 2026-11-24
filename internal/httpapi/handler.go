package httpapi

import (
	"io"
	"net/http"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", status)
	mux.HandleFunc("GET /readyz", status)
	return mux
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodHead {
		_, _ = io.WriteString(w, "{\"status\":\"ok\"}\n")
	}
}
