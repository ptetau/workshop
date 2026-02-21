package bugbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	storage "workshop/internal/adapters/storage"
	domain "workshop/internal/domain/bugbox"
)

type sqliteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore returns a Store backed by SQLite.
func NewSQLiteStore(db storage.SQLDB) Store {
	return &sqliteStore{db: db}
}

// Save persists a Submission.
// PRE: s.ID is non-empty and unique
// POST: row inserted into bugbox_submission
func (s *sqliteStore) Save(ctx context.Context, sub domain.Submission) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO bugbox_submission (
			id, summary, description, steps, expected, actual,
			route, user_agent, viewport, role, impersonated_role,
			submitted_at, screenshot_path, github_issue_number, github_issue_url
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		sub.ID,
		sub.Summary,
		sub.Description,
		sub.Steps,
		sub.Expected,
		sub.Actual,
		sub.Route,
		sub.UserAgent,
		sub.Viewport,
		sub.Role,
		sub.ImpersonatedRole,
		sub.SubmittedAt.UTC().Format(time.RFC3339),
		sub.ScreenshotPath,
		sub.GitHubIssueNumber,
		sub.GitHubIssueURL,
	)
	if err != nil {
		return fmt.Errorf("bugbox save: %w", err)
	}
	return nil
}

// GetByID retrieves a Submission by its ID.
// PRE: id is non-empty
// POST: returns domain.Submission or error if not found
func (s *sqliteStore) GetByID(ctx context.Context, id string) (domain.Submission, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, summary, description, steps, expected, actual,
		       route, user_agent, viewport, role, impersonated_role,
		       submitted_at, screenshot_path, github_issue_number, github_issue_url
		FROM bugbox_submission WHERE id = ?`, id)

	var sub domain.Submission
	var submittedAt string
	err := row.Scan(
		&sub.ID,
		&sub.Summary,
		&sub.Description,
		&sub.Steps,
		&sub.Expected,
		&sub.Actual,
		&sub.Route,
		&sub.UserAgent,
		&sub.Viewport,
		&sub.Role,
		&sub.ImpersonatedRole,
		&submittedAt,
		&sub.ScreenshotPath,
		&sub.GitHubIssueNumber,
		&sub.GitHubIssueURL,
	)
	if err == sql.ErrNoRows {
		return domain.Submission{}, fmt.Errorf("bugbox submission not found: %s", id)
	}
	if err != nil {
		return domain.Submission{}, fmt.Errorf("bugbox get: %w", err)
	}
	sub.SubmittedAt, _ = time.Parse(time.RFC3339, submittedAt)
	return sub, nil
}
