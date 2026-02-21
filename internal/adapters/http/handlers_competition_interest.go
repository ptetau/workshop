package web

import (
	"encoding/json"
	"net/http"
	"time"

	"workshop/internal/adapters/http/middleware"
	calendarDomain "workshop/internal/domain/calendar"
)

// handleCompetitionInterest handles POST/DELETE/GET for /api/calendar/interest
// Members can register/unregister interest in competitions. Admin/Coach can view who's interested.
func handleCompetitionInterest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "calendar") {
		return
	}

	// POST: Register interest (member only)
	if r.Method == "POST" {
		if sess.Role != "member" && sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "members only", http.StatusForbidden)
			return
		}
		var input struct {
			EventID string `json:"event_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if input.EventID == "" {
			http.Error(w, "event_id required", http.StatusBadRequest)
			return
		}

		// Get member ID for current account
		member, err := stores.MemberStore.GetByAccountID(ctx, sess.AccountID)
		if err != nil {
			http.Error(w, "member not found", http.StatusNotFound)
			return
		}

		ci := calendarDomain.CompetitionInterest{
			ID:        generateID(),
			EventID:   input.EventID,
			MemberID:  member.ID,
			CreatedAt: time.Now(),
		}
		if err := stores.CompetitionInterestStore.SaveInterest(ctx, ci); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
		return
	}

	// DELETE: Unregister interest (member only, their own)
	if r.Method == "DELETE" {
		if sess.Role != "member" && sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "members only", http.StatusForbidden)
			return
		}
		eventID := r.URL.Query().Get("event_id")
		if eventID == "" {
			http.Error(w, "event_id required", http.StatusBadRequest)
			return
		}

		member, err := stores.MemberStore.GetByAccountID(ctx, sess.AccountID)
		if err != nil {
			http.Error(w, "member not found", http.StatusNotFound)
			return
		}

		if err := stores.CompetitionInterestStore.DeleteInterest(ctx, eventID, member.ID); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// GET: List interested members for an event (admin/coach only)
	if r.Method == "GET" {
		if sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "admin/coach only", http.StatusForbidden)
			return
		}
		eventID := r.URL.Query().Get("event_id")
		if eventID == "" {
			http.Error(w, "event_id required", http.StatusBadRequest)
			return
		}

		interests, err := stores.CompetitionInterestStore.GetInterestsByEvent(ctx, eventID)
		if err != nil {
			internalError(w, err)
			return
		}

		// Enrich with member names
		type result struct {
			MemberID   string `json:"member_id"`
			MemberName string `json:"member_name"`
			CreatedAt  string `json:"created_at"`
		}
		var results []result
		for _, ci := range interests {
			member, err := stores.MemberStore.GetByID(ctx, ci.MemberID)
			if err != nil {
				continue // Skip if member not found
			}
			results = append(results, result{
				MemberID:   ci.MemberID,
				MemberName: member.Name,
				CreatedAt:  ci.CreatedAt.Format("2006-01-02"),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
