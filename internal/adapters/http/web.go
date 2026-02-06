package web

import "net/http"

// NewMux wires HTTP handlers for the app.
func NewMux(staticDir string) *http.ServeMux {
    mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.Dir(staticDir)))
    registerRoutes(mux)
    return mux
}
