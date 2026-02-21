package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"workshop/internal/adapters/http/middleware"
	"workshop/internal/application/orchestrators"
)

// handleBugBoxSubmit handles POST /api/admin/bugbox.
// Accepts multipart form data: summary, description, steps, expected, actual,
// route, userAgent, viewport, and optional screenshot file.
// PRE: caller is Admin or Coach; CSRF token present.
// POST: GitHub issue created, submission persisted; returns JSON with issue URL.
func handleBugBoxSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !middleware.IsCoachOrAdmin(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "bugbox") {
		return
	}

	const maxUpload = 6 << 20 // 6 MB to allow for 5 MB image + form overhead
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		http.Error(w, "request too large or malformed", http.StatusBadRequest)
		return
	}

	summary := strings.TrimSpace(r.FormValue("summary"))
	description := strings.TrimSpace(r.FormValue("description"))
	if summary == "" || description == "" {
		http.Error(w, "summary and description are required", http.StatusBadRequest)
		return
	}

	// Validate screenshot if provided (type + size; no secrets in payload)
	var screenshotPath string
	if file, header, err := r.FormFile("screenshot"); err == nil {
		defer file.Close()
		const maxScreenshot = 5 << 20 // 5 MB
		if header.Size > maxScreenshot {
			http.Error(w, "screenshot must be under 5 MB", http.StatusBadRequest)
			return
		}
		ct := header.Header.Get("Content-Type")
		if ct != "image/png" && ct != "image/jpeg" && ct != "image/webp" && ct != "image/gif" {
			http.Error(w, "screenshot must be an image (png, jpeg, webp, gif)", http.StatusBadRequest)
			return
		}
		submissionID := generateID()
		screenshotPath = "bugbox/" + submissionID + "-screenshot"
		if saveErr := saveBugBoxScreenshot(screenshotPath, file); saveErr != nil {
			slog.Error("bugbox_screenshot_save_failed", "error", saveErr.Error())
			screenshotPath = ""
		}
	}

	impersonatedRole := ""
	if sess.IsImpersonating() {
		impersonatedRole = sess.RealRole
	}

	cmd := orchestrators.SubmitBugBoxCommand{
		ID:               generateID(),
		Summary:          summary,
		Description:      description,
		Steps:            strings.TrimSpace(r.FormValue("steps")),
		Expected:         strings.TrimSpace(r.FormValue("expected")),
		Actual:           strings.TrimSpace(r.FormValue("actual")),
		Route:            strings.TrimSpace(r.FormValue("route")),
		UserAgent:        strings.TrimSpace(r.FormValue("userAgent")),
		Viewport:         strings.TrimSpace(r.FormValue("viewport")),
		Role:             sess.Role,
		ImpersonatedRole: impersonatedRole,
		ScreenshotPath:   screenshotPath,
		GitHubToken:      os.Getenv("GITHUB_TOKEN"),
		GitHubRepo:       os.Getenv("GITHUB_REPO"),
	}

	deps := orchestrators.SubmitBugBoxDeps{
		BugBoxStore: stores.BugBoxStore,
	}

	result, err := orchestrators.ExecuteSubmitBugBox(ctx, cmd, deps)
	if err != nil {
		slog.Error("bugbox_submit_failed", "error", err.Error())
		http.Error(w, "Failed to submit bug report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"submissionID":      result.SubmissionID,
		"githubIssueNumber": result.GitHubIssueNumber,
		"githubIssueURL":    result.GitHubIssueURL,
	})
}

// handleBugBoxScreenshot handles GET /api/admin/bugbox/screenshot?id=<submission-id>.
// Returns the screenshot for a given submission. Admin-only.
// PRE: caller is Admin; id query param present.
// POST: returns image bytes or 404.
func handleBugBoxScreenshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	sess2, ok2 := middleware.GetSessionFromContext(ctx)
	if !ok2 {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !middleware.IsAdmin(ctx) {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess2, "bugbox") {
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	sub, err := stores.BugBoxStore.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if sub.ScreenshotPath == "" {
		http.Error(w, "no screenshot for this submission", http.StatusNotFound)
		return
	}

	data, err := loadBugBoxScreenshot(sub.ScreenshotPath)
	if err != nil {
		http.Error(w, "screenshot not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"bugbox-%s.png\"", id))
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}

// saveBugBoxScreenshot writes screenshot bytes to a local file under the uploads directory.
// PRE: relPath is a relative path under "uploads/"; src is a valid io.Reader.
// POST: file created at uploads/<relPath>.
func saveBugBoxScreenshot(relPath string, src io.Reader) error {
	fullPath := filepath.Join("uploads", relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, src); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// loadBugBoxScreenshot reads screenshot bytes from the uploads directory.
// PRE: relPath is a relative path under "uploads/".
// POST: returns file bytes or error if not found.
func loadBugBoxScreenshot(relPath string) ([]byte, error) {
	return os.ReadFile(filepath.Join("uploads", relPath))
}
