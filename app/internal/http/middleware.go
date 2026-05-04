package http

import (
	"net/http"
	"os"
	"strings"
)

// Auth validates the Bearer token against AGENTE_API_KEY env var.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := os.Getenv("GATEWAY_API_KEY")
		if expected == "" {
			expected = "gateway-default-api-key-2025"
		}
		header := r.Header.Get("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		token = strings.TrimSpace(token)
		if token == "" || token != expected {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status":401,"message":"Authorization inválido ou ausente"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
