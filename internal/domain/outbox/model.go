package outbox

import (
	"errors"
	"time"
)

// Status constants for outbox entry lifecycle.
const (
	StatusPending   = "pending"
	StatusRetrying  = "retrying"
	StatusDone      = "done"
	StatusFailed    = "failed"
	StatusAbandoned = "abandoned"
)

// Action type constants for different external integrations.
const (
	ActionTypeGitHubIssue = "github_issue"
	ActionTypeEmail       = "email"
)

// Domain errors.
var (
	ErrEmptyActionType = errors.New("action type is required")
	ErrEmptyPayload    = errors.New("payload is required")
	ErrInvalidStatus   = errors.New("invalid status transition")
	ErrMaxRetries      = errors.New("max retry attempts reached")
)

// Entry represents a single external integration action in the outbox.
type Entry struct {
	ID              string
	ActionType      string // e.g., "github_issue", "email"
	Payload         string // JSON payload for replay
	Status          string // pending, retrying, done, failed, abandoned
	Attempts        int
	MaxAttempts     int
	LastAttemptedAt time.Time
	CreatedAt       time.Time
	ExternalID      string // ID of the external resource created (e.g., GitHub issue number)
	ErrorMessage    string // Last error message if failed
}

// Validate checks that the Entry has valid data.
// PRE: Entry struct is populated
// POST: Returns nil if valid, error otherwise
func (e *Entry) Validate() error {
	if e.ActionType == "" {
		return ErrEmptyActionType
	}
	if e.Payload == "" {
		return ErrEmptyPayload
	}
	if e.CreatedAt.IsZero() {
		return errors.New("created_at must be set")
	}
	if e.MaxAttempts <= 0 {
		e.MaxAttempts = 5 // Default max attempts
	}
	return nil
}

// CanRetry returns true if the entry can be retried.
// PRE: Status and Attempts fields are set
// POST: Returns true for pending/retrying/failed with attempts < max
func (e *Entry) CanRetry() bool {
	return (e.Status == StatusPending || e.Status == StatusRetrying || e.Status == StatusFailed) &&
		e.Attempts < e.MaxAttempts
}

// IsTerminal returns true if the entry has reached a terminal state.
// PRE: Status field is set
// POST: Returns true for done, failed (max retries), or abandoned
func (e *Entry) IsTerminal() bool {
	if e.Status == StatusDone || e.Status == StatusAbandoned {
		return true
	}
	if e.Status == StatusFailed && e.Attempts >= e.MaxAttempts {
		return true
	}
	return false
}

// MarkAttempt records a retry attempt.
// PRE: Entry is in a retryable state
// POST: Attempts incremented, LastAttemptedAt updated, status set to retrying
func (e *Entry) MarkAttempt() {
	e.Attempts++
	e.LastAttemptedAt = time.Now()
	e.Status = StatusRetrying
}

// MarkSuccess marks the entry as successfully completed.
// PRE: External action completed successfully
// POST: Status set to done, ExternalID can be set
func (e *Entry) MarkSuccess(externalID string) {
	e.Status = StatusDone
	e.ExternalID = externalID
	e.ErrorMessage = ""
}

// MarkFailed marks the entry as failed with an error message.
// PRE: External action failed
// POST: Status set to failed or remains retrying, ErrorMessage set
func (e *Entry) MarkFailed(err error) {
	e.ErrorMessage = err.Error()
	if e.Attempts >= e.MaxAttempts {
		e.Status = StatusFailed
	}
}

// MarkAbandoned marks the entry as abandoned by admin.
// PRE: Admin explicitly abandons the entry
// POST: Status set to abandoned
func (e *Entry) MarkAbandoned() {
	e.Status = StatusAbandoned
}

// NextRetryDelay calculates the delay before the next retry attempt.
// Uses exponential backoff: 2^attempts * baseDelay, capped at maxDelay.
// PRE: Attempts is set
// POST: Returns duration for next retry
func (e *Entry) NextRetryDelay(baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	delay := baseDelay * (1 << e.Attempts)
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}
