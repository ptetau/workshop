package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"workshop/internal/adapters/http/middleware"
	domain "workshop/internal/domain/personalgoal"
)

// handlePersonalGoals handles GET/POST/DELETE for /api/personal-goals
func handlePersonalGoals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "calendar") {
		return
	}

	// Get member ID for the current user
	member, err := stores.MemberStore.GetByAccountID(ctx, sess.AccountID)
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		goals, err := stores.PersonalGoalStore.ListByMemberID(ctx, member.ID)
		if err != nil {
			internalError(w, err)
			return
		}
		if goals == nil {
			goals = []domain.PersonalGoal{}
		}

		// Calculate auto-progress for hours-type goals
		for i, g := range goals {
			if g.IsAutoTracked() {
				hours, err := stores.AttendanceStore.SumMatHoursByMemberIDAndDateRange(ctx, member.ID,
					g.StartDate.Format("2006-01-02"), g.EndDate.Format("2006-01-02"))
				if err == nil {
					goals[i].Progress = int(hours)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(goals)

	case "POST":
		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Target      int    `json:"target"`
			Unit        string `json:"unit"`
			Type        string `json:"type"`
			StartDate   string `json:"start_date"`
			EndDate     string `json:"end_date"`
			Color       string `json:"color"`
			Progress    int    `json:"progress"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			http.Error(w, "invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}

		now := time.Now()
		goal := domain.PersonalGoal{
			ID:          uuid.New().String(),
			MemberID:    member.ID,
			Title:       input.Title,
			Description: input.Description,
			Target:      input.Target,
			Unit:        input.Unit,
			Type:        input.Type,
			StartDate:   startDate,
			EndDate:     endDate,
			Color:       input.Color,
			Progress:    input.Progress,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		goal.SetDefaultColor()
		goal.SetDefaultType()

		if err := goal.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := stores.PersonalGoalStore.Save(ctx, goal); err != nil {
			internalError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(goal)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		// Verify ownership before deleting
		goal, err := stores.PersonalGoalStore.GetByID(ctx, id)
		if err != nil {
			http.Error(w, "goal not found", http.StatusNotFound)
			return
		}
		if goal.MemberID != member.ID && sess.Role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if err := stores.PersonalGoalStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handlePersonalGoalProgress handles PUT for /api/personal-goals/progress
func handlePersonalGoalProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "calendar") {
		return
	}

	var input struct {
		ID       string `json:"id"`
		Progress int    `json:"progress"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Get member ID for the current user
	member, err := stores.MemberStore.GetByAccountID(ctx, sess.AccountID)
	if err != nil {
		http.Error(w, "member not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	goal, err := stores.PersonalGoalStore.GetByID(ctx, input.ID)
	if err != nil {
		http.Error(w, "goal not found", http.StatusNotFound)
		return
	}
	if goal.MemberID != member.ID && sess.Role != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	goal.UpdateProgress(input.Progress)
	if err := stores.PersonalGoalStore.Save(ctx, goal); err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}
