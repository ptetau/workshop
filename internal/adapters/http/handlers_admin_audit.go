package web

import (
	"net/http"
	"strconv"

	"workshop/internal/adapters/http/middleware"
	auditStore "workshop/internal/adapters/storage/audit"
	auditDomain "workshop/internal/domain/audit"
)

// handleAdminAuditTrail renders the admin audit trail page (GET /admin/audit-trail)
// PRE: User must be authenticated as admin
// POST: Renders audit trail with optional filters
func handleAdminAuditTrail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	if !requireFeaturePage(w, r, sess, "audit_trail") {
		return
	}

	// Only admins can view audit trail
	if sess.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// Parse query parameters for filtering
	filter := auditStore.Filter{}

	if category := r.URL.Query().Get("category"); category != "" {
		cat := auditDomain.Category(category)
		filter.Category = &cat
	}
	if action := r.URL.Query().Get("action"); action != "" {
		act := auditDomain.Action(action)
		filter.Action = &act
	}
	if actorID := r.URL.Query().Get("actor_id"); actorID != "" {
		filter.ActorID = &actorID
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		sev := auditDomain.Severity(severity)
		filter.Severity = &sev
	}
	if resourceID := r.URL.Query().Get("resource_id"); resourceID != "" {
		filter.ResourceID = &resourceID
	}
	if fromDate := r.URL.Query().Get("from"); fromDate != "" {
		filter.FromDate = &fromDate
	}
	if toDate := r.URL.Query().Get("to"); toDate != "" {
		filter.ToDate = &toDate
	}

	// Parse limit, default to 100
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Fetch audit events
	events, err := stores.AuditStore.List(ctx, filter, limit)
	if err != nil {
		internalError(w, err)
		return
	}

	renderTemplate(w, r, "admin_audit_trail.html", map[string]any{
		"Events": events,
		"Filter": filter,
		"Limit":  limit,
	})
}
