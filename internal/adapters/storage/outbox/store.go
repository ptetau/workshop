package outbox

import (
	"context"

	domain "workshop/internal/domain/outbox"
)

// Store defines the interface for outbox entry persistence.
type Store interface {
	// GetByID retrieves an outbox entry by its ID.
	// PRE: id is non-empty
	// POST: Returns the entry or an error if not found
	GetByID(ctx context.Context, id string) (domain.Entry, error)

	// Save persists an outbox entry to the database.
	// PRE: entity has been validated
	// POST: Entity is persisted (insert or update)
	Save(ctx context.Context, e domain.Entry) error

	// ListPending returns entries that need to be processed (pending or retrying).
	// PRE: limit > 0
	// POST: Returns up to limit entries ordered by created_at
	ListPending(ctx context.Context, limit int) ([]domain.Entry, error)

	// ListFailed returns entries that have permanently failed.
	// PRE: limit > 0
	// POST: Returns up to limit failed entries ordered by last_attempted_at desc
	ListFailed(ctx context.Context, limit int) ([]domain.Entry, error)

	// ListByActionType returns entries filtered by action type and status.
	// PRE: actionType is non-empty
	// POST: Returns matching entries ordered by created_at
	ListByActionType(ctx context.Context, actionType string, status string, limit int) ([]domain.Entry, error)

	// Delete removes an outbox entry (only for abandoned/terminal entries).
	// PRE: id is non-empty and entry is in terminal state
	// POST: Entry is removed from database
	Delete(ctx context.Context, id string) error
}
