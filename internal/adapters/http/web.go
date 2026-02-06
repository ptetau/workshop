package web

import (
	"net/http"

	"workshop/internal/adapters/http/middleware"
)

// NewMux wires HTTP handlers for the app.
func NewMux(staticDir string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	registerRoutes(mux)

	// Create CSRF protection (key should be env var in prod)
	csrfKey := []byte("01234567890123456789012345678901") // 32 bytes

	// Apply middleware: CSRF -> SecurityHeaders -> Mux
	return middleware.Chain(mux, middleware.SecurityHeaders, middleware.CSRF(csrfKey))
}
