package orchestrators

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"workshop/internal/adapters/storage/outbox"
	domainOutbox "workshop/internal/domain/outbox"
)

// OutboxRetryDeps provides the dependencies for retrying outbox entries.
type OutboxRetryDeps struct {
	OutboxStore outbox.Store
}

// ExecuteOutboxRetry processes pending and retryable failed outbox entries.
// It implements exponential backoff and respects max attempts.
// PRE: Deps are valid and store is connected
// POST: All eligible entries are processed, results logged
func ExecuteOutboxRetry(ctx context.Context, deps OutboxRetryDeps) error {
	// Fetch all entries that can be retried
	entries, err := deps.OutboxStore.ListPending(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to list retryable outbox entries: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	slog.Info("outbox_retry_start", "count", len(entries))

	var processed, succeeded, failed int
	baseDelay := 1 * time.Minute
	maxDelay := 1 * time.Hour

	for _, entry := range entries {
		processed++

		// Check if enough time has passed since last attempt
		if !entry.LastAttemptedAt.IsZero() {
			nextRetry := entry.LastAttemptedAt.Add(entry.NextRetryDelay(baseDelay, maxDelay))
			if time.Now().Before(nextRetry) {
				slog.Debug("outbox_retry_skipped_backoff", "entry_id", entry.ID, "next_retry", nextRetry)
				continue
			}
		}

		// Mark attempt
		entry.MarkAttempt()

		// Execute based on action type
		var err error
		switch entry.ActionType {
		case domainOutbox.ActionTypeGitHubIssue:
			err = retryGitHubIssue(ctx, entry)
		case domainOutbox.ActionTypeEmail:
			err = retryEmail(ctx, entry)
		default:
			err = fmt.Errorf("unknown action type: %s", entry.ActionType)
		}

		if err != nil {
			entry.MarkFailed(err)
			failed++
			slog.Error("outbox_retry_failed", "entry_id", entry.ID, "action", entry.ActionType, "attempt", entry.Attempts, "error", err)
		} else {
			entry.MarkSuccess("") // External ID set by specific handler
			succeeded++
			slog.Info("outbox_retry_succeeded", "entry_id", entry.ID, "action", entry.ActionType, "attempt", entry.Attempts)
		}

		// Save updated entry
		if saveErr := deps.OutboxStore.Save(ctx, entry); saveErr != nil {
			slog.Error("outbox_retry_save_failed", "entry_id", entry.ID, "error", saveErr)
		}
	}

	slog.Info("outbox_retry_complete", "processed", processed, "succeeded", succeeded, "failed", failed)
	return nil
}

// retryGitHubIssue attempts to create a GitHub issue from the outbox payload.
// PRE: Entry payload contains valid GitHub issue data
// POST: GitHub issue created or error returned
func retryGitHubIssue(ctx context.Context, entry domainOutbox.Entry) error {
	var payload struct {
		Title      string   `json:"title"`
		Body       string   `json:"body"`
		Labels     []string `json:"labels"`
		Repository string   `json:"repository"`
	}

	if err := json.Unmarshal([]byte(entry.Payload), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal github issue payload: %w", err)
	}

	// GitHub integration would be implemented here
	// For now, log that we would create the issue
	slog.Info("outbox_retry_github_issue", "title", payload.Title, "repository", payload.Repository)

	// TODO: Implement actual GitHub API call
	// This requires GitHub token configuration (issue #298)

	return fmt.Errorf("github integration not yet implemented (issue #298)")
}

// retryEmail attempts to send an email from the outbox payload.
// PRE: Entry payload contains valid email data
// POST: Email sent or error returned
func retryEmail(ctx context.Context, entry domainOutbox.Entry) error {
	var payload struct {
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Body    string   `json:"body"`
		HTML    bool     `json:"html"`
	}

	if err := json.Unmarshal([]byte(entry.Payload), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email payload: %w", err)
	}

	// Email sending would be implemented here
	// This requires email sender configuration
	slog.Info("outbox_retry_email", "to_count", len(payload.To), "subject", payload.Subject)

	// TODO: Implement actual email sending using configured email sender

	return fmt.Errorf("email retry not yet implemented")
}

// OutboxRetryConfig holds configuration for the retry scheduler.
type OutboxRetryConfig struct {
	Interval    time.Duration // How often to run retries
	MaxAttempts int           // Max attempts per entry (overrides entry's max)
	Enabled     bool
}

// DefaultOutboxRetryConfig returns sensible defaults.
func DefaultOutboxRetryConfig() OutboxRetryConfig {
	return OutboxRetryConfig{
		Interval:    5 * time.Minute,
		MaxAttempts: 10,
		Enabled:     true,
	}
}

// StartOutboxRetryScheduler starts a background goroutine that periodically retries outbox entries.
// PRE: Context is valid, deps are initialized
// POST: Goroutine started, returns cancel function
func StartOutboxRetryScheduler(ctx context.Context, deps OutboxRetryDeps, cfg OutboxRetryConfig) func() {
	if !cfg.Enabled {
		return func() {}
	}

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := ExecuteOutboxRetry(ctx, deps); err != nil {
					slog.Error("outbox_retry_scheduler_error", "error", err)
				}
			}
		}
	}()

	return cancel
}
