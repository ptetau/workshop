package orchestrators

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	outboxStore "workshop/internal/adapters/storage/outbox"
	domain "workshop/internal/domain/outbox"
)

// OutboxProcessor handles retrying failed external integration actions.
type OutboxProcessor struct {
	store     outboxStore.Store
	executors map[string]ActionExecutor
	baseDelay time.Duration
	maxDelay  time.Duration
	batchSize int
}

// ActionExecutor executes a specific type of external action.
type ActionExecutor interface {
	// Execute runs the external action with the given payload.
	// Returns the external ID (e.g., GitHub issue number) and any error.
	Execute(ctx context.Context, payload string) (string, error)
}

// NewOutboxProcessor creates a new outbox processor.
func NewOutboxProcessor(store outboxStore.Store, executors map[string]ActionExecutor) *OutboxProcessor {
	return &OutboxProcessor{
		store:     store,
		executors: executors,
		baseDelay: 30 * time.Second,
		maxDelay:  1 * time.Hour,
		batchSize: 10,
	}
}

// ProcessPending processes pending outbox entries with retries.
// PRE: Context is valid
// POST: Pending entries are processed, failed entries marked for retry
func (p *OutboxProcessor) ProcessPending(ctx context.Context) error {
	entries, err := p.store.ListPending(ctx, p.batchSize)
	if err != nil {
		return fmt.Errorf("list pending outbox entries: %w", err)
	}

	for _, entry := range entries {
		if err := p.processEntry(ctx, entry); err != nil {
			slog.Error("outbox_process_failed", "entry_id", entry.ID, "action_type", entry.ActionType, "error", err.Error())
		}
	}

	return nil
}

// processEntry processes a single outbox entry.
func (p *OutboxProcessor) processEntry(ctx context.Context, entry domain.Entry) error {
	// Check if enough time has passed since last attempt
	if !entry.LastAttemptedAt.IsZero() {
		delay := entry.NextRetryDelay(p.baseDelay, p.maxDelay)
		if time.Since(entry.LastAttemptedAt) < delay {
			return nil // Not ready to retry yet
		}
	}

	executor, ok := p.executors[entry.ActionType]
	if !ok {
		entry.MarkFailed(fmt.Errorf("no executor registered for action type: %s", entry.ActionType))
		return p.store.Save(ctx, entry)
	}

	entry.MarkAttempt()
	externalID, err := executor.Execute(ctx, entry.Payload)
	if err != nil {
		entry.MarkFailed(err)
		slog.Warn("outbox_action_failed", "entry_id", entry.ID, "attempt", entry.Attempts, "error", err.Error())
	} else {
		entry.MarkSuccess(externalID)
		slog.Info("outbox_action_succeeded", "entry_id", entry.ID, "action_type", entry.ActionType, "external_id", externalID)
	}

	return p.store.Save(ctx, entry)
}

// ProcessSingle manually processes a single outbox entry (for admin retry).
// PRE: entryID is non-empty
// POST: Entry is processed, status updated
func (p *OutboxProcessor) ProcessSingle(ctx context.Context, entryID string) error {
	entry, err := p.store.GetByID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("get outbox entry: %w", err)
	}

	if entry.IsTerminal() {
		return fmt.Errorf("entry %s is in terminal state and cannot be retried", entryID)
	}

	executor, ok := p.executors[entry.ActionType]
	if !ok {
		return fmt.Errorf("no executor registered for action type: %s", entry.ActionType)
	}

	entry.MarkAttempt()
	externalID, err := executor.Execute(ctx, entry.Payload)
	if err != nil {
		entry.MarkFailed(err)
	} else {
		entry.MarkSuccess(externalID)
	}

	return p.store.Save(ctx, entry)
}

// AbandonEntry marks an entry as abandoned by admin.
// PRE: entryID is non-empty
// POST: Entry status set to abandoned
func (p *OutboxProcessor) AbandonEntry(ctx context.Context, entryID string) error {
	entry, err := p.store.GetByID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("get outbox entry: %w", err)
	}

	entry.MarkAbandoned()
	return p.store.Save(ctx, entry)
}

// --- GitHub Issue Executor ---

// GitHubIssuePayload is the JSON structure for GitHub issue creation.
type GitHubIssuePayload struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Labels      []string `json:"labels"`
	GitHubToken string   `json:"github_token"`
	GitHubRepo  string   `json:"github_repo"`
}

// GitHubIssueExecutor creates GitHub issues.
type GitHubIssueExecutor struct {
	HTTPClient *http.Client
}

// Execute creates a GitHub issue from the payload.
// PRE: payload is valid JSON matching GitHubIssuePayload
// POST: GitHub issue created, returns issue URL
// INVARIANT: outbox entry status managed by caller
func (e *GitHubIssueExecutor) Execute(ctx context.Context, payload string) (string, error) {
	var p GitHubIssuePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return "", fmt.Errorf("unmarshal payload: %w", err)
	}

	// Delegate to the existing bug box submission logic
	// This should be integrated with submit_bugbox.go
	return "", fmt.Errorf("github issue executor not yet fully implemented")
}

// --- Email Executor ---

// EmailPayload is the JSON structure for email sending.
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// EmailExecutor sends emails.
type EmailExecutor struct {
	// Email sender implementation
}

// Execute sends an email from the payload.
// PRE: payload is valid JSON matching EmailPayload
// POST: email sent via configured sender, returns message ID
// INVARIANT: outbox entry status managed by caller
func (e *EmailExecutor) Execute(ctx context.Context, payload string) (string, error) {
	var p EmailPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return "", fmt.Errorf("unmarshal payload: %w", err)
	}

	// Delegate to the existing email orchestrator
	return "", fmt.Errorf("email executor not yet fully implemented")
}

// --- Background Worker ---

// StartBackgroundWorker starts a background goroutine that periodically processes pending outbox entries.
// PRE: stopCh is provided to signal shutdown
// POST: Worker runs until stopCh is closed
func StartBackgroundWorker(processor *OutboxProcessor, interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				if err := processor.ProcessPending(ctx); err != nil {
					slog.Error("outbox_background_process_failed", "error", err.Error())
				}
				cancel()
			case <-stopCh:
				slog.Info("outbox_background_worker_stopped")
				return
			}
		}
	}()
}
