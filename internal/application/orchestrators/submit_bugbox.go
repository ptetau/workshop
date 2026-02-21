package orchestrators

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	bugboxStore "workshop/internal/adapters/storage/bugbox"
	domain "workshop/internal/domain/bugbox"
)

// SubmitBugBoxCommand holds the input for a Bug Box submission.
// PRE: Summary and Description are non-empty.
// POST: A GitHub issue is created and the submission is persisted.
// INVARIANT: Context must never include cookies, CSRF tokens, passwords, or raw session data.
type SubmitBugBoxCommand struct {
	ID               string
	Summary          string
	Description      string
	Steps            string
	Expected         string
	Actual           string
	Route            string
	UserAgent        string
	Viewport         string
	Role             string
	ImpersonatedRole string
	ScreenshotPath   string

	// GitHub integration config (injected from env vars — never hardcoded)
	GitHubToken string
	GitHubRepo  string
}

// SubmitBugBoxResult holds the outcome of a successful submission.
type SubmitBugBoxResult struct {
	SubmissionID      string
	GitHubIssueNumber int
	GitHubIssueURL    string
}

// SubmitBugBoxDeps are the external dependencies for this orchestrator.
type SubmitBugBoxDeps struct {
	BugBoxStore bugboxStore.Store
	HTTPClient  *http.Client
}

// ExecuteSubmitBugBox validates, creates a GitHub issue, and persists the submission.
// PRE: cmd.Summary and cmd.Description are non-empty; GitHubToken and GitHubRepo are set.
// POST: GitHub issue created, submission saved; returns issue number and URL.
func ExecuteSubmitBugBox(ctx context.Context, cmd SubmitBugBoxCommand, deps SubmitBugBoxDeps) (SubmitBugBoxResult, error) {
	sub := domain.Submission{
		ID:               cmd.ID,
		Summary:          cmd.Summary,
		Description:      cmd.Description,
		Steps:            cmd.Steps,
		Expected:         cmd.Expected,
		Actual:           cmd.Actual,
		Route:            cmd.Route,
		UserAgent:        cmd.UserAgent,
		Viewport:         cmd.Viewport,
		Role:             cmd.Role,
		ImpersonatedRole: cmd.ImpersonatedRole,
		ScreenshotPath:   cmd.ScreenshotPath,
		SubmittedAt:      time.Now().UTC(),
	}

	if err := sub.Validate(); err != nil {
		return SubmitBugBoxResult{}, fmt.Errorf("validation: %w", err)
	}

	if cmd.GitHubToken == "" || cmd.GitHubRepo == "" {
		return SubmitBugBoxResult{}, fmt.Errorf("GitHub integration not configured: GITHUB_TOKEN and GITHUB_REPO must be set")
	}

	issueNum, issueURL, err := createGitHubIssue(ctx, cmd, deps.HTTPClient)
	if err != nil {
		slog.Error("bugbox_github_issue_failed", "error", err.Error(), "submission_id", cmd.ID)
		return SubmitBugBoxResult{}, fmt.Errorf("failed to create GitHub issue: %w", err)
	}

	sub.GitHubIssueNumber = issueNum
	sub.GitHubIssueURL = issueURL

	if err := deps.BugBoxStore.Save(ctx, sub); err != nil {
		slog.Error("bugbox_save_failed", "error", err.Error(), "submission_id", cmd.ID)
		return SubmitBugBoxResult{}, fmt.Errorf("failed to save submission: %w", err)
	}

	slog.Info("bugbox_submitted", "submission_id", cmd.ID, "github_issue", issueNum)
	return SubmitBugBoxResult{
		SubmissionID:      cmd.ID,
		GitHubIssueNumber: issueNum,
		GitHubIssueURL:    issueURL,
	}, nil
}

// createGitHubIssue calls the GitHub REST API to create a new issue.
// PRE: token and repo are non-empty; ctx is valid.
// POST: returns the new issue number and HTML URL.
func createGitHubIssue(ctx context.Context, cmd SubmitBugBoxCommand, client *http.Client) (int, string, error) {
	body := buildIssueBody(cmd)

	payload := map[string]interface{}{
		"title":  fmt.Sprintf("[BugBox] %s", cmd.Summary),
		"body":   body,
		"labels": []string{"bug", "bugbox"},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return 0, "", fmt.Errorf("marshal payload: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/issues", cmd.GitHubRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(data))
	if err != nil {
		return 0, "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cmd.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("github api request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return 0, "", fmt.Errorf("github api returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, "", fmt.Errorf("parse github response: %w", err)
	}

	return result.Number, result.HTMLURL, nil
}

// buildIssueBody constructs the GitHub issue body from the submission command.
func buildIssueBody(cmd SubmitBugBoxCommand) string {
	var sb strings.Builder

	sb.WriteString("## Bug Report\n\n")
	sb.WriteString("### Description\n\n")
	sb.WriteString(cmd.Description)
	sb.WriteString("\n\n")

	if cmd.Steps != "" {
		sb.WriteString("### Steps to Reproduce\n\n")
		sb.WriteString(cmd.Steps)
		sb.WriteString("\n\n")
	}
	if cmd.Expected != "" {
		sb.WriteString("### Expected Behaviour\n\n")
		sb.WriteString(cmd.Expected)
		sb.WriteString("\n\n")
	}
	if cmd.Actual != "" {
		sb.WriteString("### Actual Behaviour\n\n")
		sb.WriteString(cmd.Actual)
		sb.WriteString("\n\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("### Browser Context\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|-------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Route | `%s` |\n", cmd.Route))
	sb.WriteString(fmt.Sprintf("| Role | `%s` |\n", cmd.Role))
	if cmd.ImpersonatedRole != "" {
		sb.WriteString(fmt.Sprintf("| Impersonating | `%s` |\n", cmd.ImpersonatedRole))
	}
	sb.WriteString(fmt.Sprintf("| Viewport | `%s` |\n", cmd.Viewport))
	sb.WriteString(fmt.Sprintf("| User-Agent | `%s` |\n", cmd.UserAgent))
	sb.WriteString(fmt.Sprintf("| Submitted At | `%s` |\n", time.Now().UTC().Format(time.RFC3339)))

	if cmd.ScreenshotPath != "" {
		sb.WriteString("\n### Screenshot\n\n")
		sb.WriteString("_A screenshot was attached to this report. An admin can retrieve it from the app._\n")
	}

	sb.WriteString("\n---\n_Submitted via Bug Box_\n")
	return sb.String()
}
