package web

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"workshop/internal/adapters/http/middleware"
	"workshop/internal/domain/consent"
	"workshop/internal/domain/deletion"
)

// handlePrivacyDeletePage renders the data deletion request form (GET /privacy/delete)
// PRE: User must be authenticated as member, trial, coach, or admin
// POST: Renders deletion request form with member info
func handlePrivacyDeletePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	if !requireFeaturePage(w, r, sess, "privacy") {
		return
	}

	// Members and trials can only view their own deletion request page
	// Coaches and admins can view too but this is primarily for members
	if sess.Role != "member" && sess.Role != "trial" && sess.Role != "coach" && sess.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// Get member info for the requesting user
	member, err := stores.MemberStore.GetByEmail(ctx, sess.Email)
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	// Check for existing deletion request
	var existingRequest *deletion.Request
	existing, err := stores.DeletionRequestStore.GetByMemberID(ctx, member.ID)
	if err == nil {
		existingRequest = &existing
	}

	renderTemplate(w, r, "privacy_delete.html", map[string]any{
		"Member":          member,
		"ExistingRequest": existingRequest,
	})
}

// handlePrivacyDeleteRequest handles data deletion request submission (POST /api/privacy/delete)
// PRE: User must be authenticated as member, trial, coach, or admin
// POST: Creates a deletion request and returns confirmation
func handlePrivacyDeleteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	// Only members and trials can request deletion of their own data
	// Coaches and admins cannot use this endpoint (they have other tools)
	if sess.Role != "member" && sess.Role != "trial" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// Get member info
	member, err := stores.MemberStore.GetByEmail(ctx, sess.Email)
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	// Check if there's already a pending request
	existing, err := stores.DeletionRequestStore.GetByMemberID(ctx, member.ID)
	if err == nil && !existing.IsTerminal() {
		http.Error(w, "deletion request already pending", http.StatusConflict)
		return
	}

	// Create new deletion request
	req := deletion.Request{
		ID:             generateID(),
		MemberID:       member.ID,
		Email:          member.Email,
		Status:         deletion.StatusPending,
		RequestedAt:    time.Now(),
		GracePeriodEnd: time.Now().Add(7 * 24 * time.Hour), // 7 day grace period
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := stores.DeletionRequestStore.Save(ctx, req); err != nil {
		internalError(w, err)
		return
	}

	// Log audit event
	slog.Info("privacy_deletion_requested",
		"request_id", req.ID,
		"member_id", member.ID,
		"email", member.Email,
		"grace_period_end", req.GracePeriodEnd,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"request_id":       req.ID,
		"status":           req.Status,
		"grace_period_end": req.GracePeriodEnd,
		"message":          "Deletion request submitted. You have 7 days to cancel before processing begins.",
	})
}

// handlePrivacyDeleteCancel handles cancellation of a pending deletion request (POST /api/privacy/delete/cancel)
// PRE: User must be authenticated as member or trial; request must be in cancellable state
// POST: Cancels the deletion request and returns confirmation
func handlePrivacyDeleteCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	// Only members and trials can cancel their own deletion requests
	if sess.Role != "member" && sess.Role != "trial" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// Get member info
	member, err := stores.MemberStore.GetByEmail(ctx, sess.Email)
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	// Get existing request
	existing, err := stores.DeletionRequestStore.GetByMemberID(ctx, member.ID)
	if err != nil {
		http.Error(w, "no deletion request found", http.StatusNotFound)
		return
	}

	// Check if request can be cancelled
	if !existing.CanCancel() {
		http.Error(w, "deletion request cannot be cancelled at this stage", http.StatusConflict)
		return
	}

	// Cancel the request
	if err := existing.MarkCancelled(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := stores.DeletionRequestStore.Save(ctx, existing); err != nil {
		internalError(w, err)
		return
	}

	// Log audit event
	slog.Info("privacy_deletion_cancelled",
		"request_id", existing.ID,
		"member_id", member.ID,
		"email", member.Email,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"request_id": existing.ID,
		"status":     existing.Status,
		"message":    "Deletion request cancelled successfully.",
	})
}

// handlePrivacyConsentPage renders the consent management page (GET /privacy/consent)
// PRE: User must be authenticated as member, trial, coach, or admin
// POST: Renders consent preferences for the member
func handlePrivacyConsentPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	if sess.Role != "member" && sess.Role != "trial" && sess.Role != "coach" && sess.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	member, err := stores.MemberStore.GetByEmail(ctx, sess.Email)
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	consents, err := stores.ConsentStore.GetByMemberID(ctx, member.ID)
	if err != nil {
		internalError(w, err)
		return
	}

	renderTemplate(w, r, "privacy_consent.html", map[string]any{
		"Member":   member,
		"Consents": consents,
	})
}

// handlePrivacyConsentRevoke handles revoking marketing consent (POST /api/privacy/consent/revoke)
// PRE: User must be authenticated as member or trial
// POST: Revokes marketing consent for the member
func handlePrivacyConsentRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	if sess.Role != "member" && sess.Role != "trial" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	member, err := stores.MemberStore.GetByEmail(ctx, sess.Email)
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	consentType := r.URL.Query().Get("type")
	if consentType == "" {
		consentType = "marketing"
	}

	consent, err := stores.ConsentStore.GetByType(ctx, member.ID, consent.Type(consentType))
	if err != nil {
		http.Error(w, "no consent found to revoke", http.StatusNotFound)
		return
	}

	if err := consent.Revoke(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := stores.ConsentStore.Save(ctx, consent); err != nil {
		internalError(w, err)
		return
	}

	slog.Info("consent_revoked",
		"member_id", member.ID,
		"consent_type", consentType,
		"email", member.Email,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"consent_type": consentType,
		"granted":      false,
		"message":      "Consent revoked successfully.",
	})
}
