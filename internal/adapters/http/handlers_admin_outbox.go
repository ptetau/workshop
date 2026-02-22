package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"workshop/internal/adapters/http/middleware"
	"workshop/internal/application/orchestrators"
	"workshop/internal/domain/outbox"
)

// handleAdminOutbox handles admin endpoints for managing outbox entries.
// Routes: GET /admin/outbox (list failed entries), POST /admin/outbox/:id/retry (manual retry), POST /admin/outbox/:id/abandon
func handleAdminOutbox(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "outbox") {
		return
	}
	if sess.Role != "admin" {
		http.Error(w, "admin required", http.StatusForbidden)
		return
	}

	switch r.Method {
	case "GET":
		// List failed entries (permanently failed or max retries)
		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}

		status := r.URL.Query().Get("status")
		if status == "" {
			status = outbox.StatusFailed
		}

		var entries []outbox.Entry
		var err error

		if status == "all" {
			entries, err = stores.OutboxStore.ListPending(ctx, limit)
		} else {
			entries, err = stores.OutboxStore.ListFailed(ctx, limit)
		}

		if err != nil {
			internalError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)

	case "POST":
		// Extract entry ID from path: /admin/outbox/:id/:action
		path := r.URL.Path
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) < 4 || parts[0] != "admin" || parts[1] != "outbox" {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		entryID := parts[2]
		action := parts[3]

		processor := orchestrators.NewOutboxProcessor(stores.OutboxStore, nil)

		switch action {
		case "retry":
			if err := processor.ProcessSingle(ctx, entryID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "retry triggered"})

		case "abandon":
			if err := processor.AbandonEntry(ctx, entryID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "abandoned"})

		default:
			http.Error(w, "unknown action", http.StatusBadRequest)
		}
	}
}
